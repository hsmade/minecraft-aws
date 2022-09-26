package catalog

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	ecsTypes "github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	route53Types "github.com/aws/aws-sdk-go-v2/service/route53/types"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/pkg/errors"
	"os"
	"strings"
	"time"
)

var timeout = 20 * time.Second

type Server struct {
	Name          string // task definition family name
	Cluster       string
	DNSZoneID     string
	Tags          map[string]string
	EcsClient     ecsClient
	Route53Client route53Client
	Ec2Client     ec2Client
}

type ServerStatus struct {
	Name   string
	Status string
	taskID string
}

func (S Server) getRunningTask() (*ecsTypes.Task, error) {
	output, err := S.EcsClient.ListTasks(context.TODO(), &ecs.ListTasksInput{
		Cluster: &S.Cluster,
		Family:  &S.Name,
	})
	if err != nil {
		return nil, errors.Wrap(err, "listing tasks")
	}
	if len(output.TaskArns) == 0 {
		return nil, nil
	}

	tasks, err := S.EcsClient.DescribeTasks(context.TODO(), &ecs.DescribeTasksInput{
		Tasks:   output.TaskArns,
		Cluster: &S.Cluster,
	})
	if err != nil {
		return nil, errors.Wrap(err, "getting tasks")
	}

	for _, task := range tasks.Tasks {
		if *task.DesiredStatus != "RUNNING" {
			continue
		}
		return &task, nil
	}
	return nil, nil
}

func (S Server) Status() (*ServerStatus, error) {
	task, err := S.getRunningTask()
	if err != nil {
		return nil, errors.Wrap(err, "getting running task")
	}
	if task == nil {
		return nil, errors.New("no running task found")
	}

	status := "NONE"
	for _, container := range task.Containers {
		if *container.Name != "main" {
			continue
		}
		status = string(container.HealthStatus)
	}
	return &ServerStatus{
		Name:   S.Name,
		Status: status,
		taskID: *task.TaskArn,
	}, nil
}

func (S Server) Stop() error {
	task, err := S.getRunningTask()
	if err != nil {
		return errors.Wrap(err, "getting running task")
	}
	if task == nil {
		return errors.New("no running task found")
	}

	errDNS := S.deleteDNSRecord() // needs to happen first, save error for later

	_, err = S.EcsClient.StopTask(context.TODO(), &ecs.StopTaskInput{
		Task:    task.TaskArn,
		Cluster: &S.Cluster,
	})
	if err != nil {
		return errors.Wrap(err, "failed to stop task")
	}

	return errors.Wrap(errDNS, "deleting DNS record")
}

func (S Server) deleteDNSRecord() error {
	ip, err := S.getIP()
	if err != nil {
		return errors.Wrap(err, "getting IP")
	}

	output, err := S.Route53Client.ListResourceRecordSets(context.TODO(), &route53.ListResourceRecordSetsInput{
		HostedZoneId:    &S.DNSZoneID,
		MaxItems:        aws.Int32(1),
		StartRecordName: aws.String(S.Name + "." + os.Getenv("DNS_ZONE")),
		StartRecordType: "A",
	})
	if err != nil {
		return errors.Wrap(err, "listing recordsets")
	}

	if len(output.ResourceRecordSets) == 0 {
		return nil
	}
	return errors.Wrap(S.modifyDNSRecord(ip, route53Types.ChangeActionDelete), "deleting DNS record")
}

func (S Server) modifyDNSRecord(ip string, action route53Types.ChangeAction) error {
	record := route53Types.ResourceRecordSet{
		Name: aws.String(S.Name + "." + os.Getenv("DNS_ZONE")),
		Type: "A",
		TTL:  aws.Int64(10),
		ResourceRecords: []route53Types.ResourceRecord{
			{
				Value: &ip,
			},
		},
	}
	_, err := S.Route53Client.ChangeResourceRecordSets(context.TODO(), &route53.ChangeResourceRecordSetsInput{
		HostedZoneId: &S.DNSZoneID,
		ChangeBatch: &route53Types.ChangeBatch{
			Changes: []route53Types.Change{
				{
					Action:            action,
					ResourceRecordSet: &record,
				},
			},
		},
	})
	return err
}

func (S Server) createOrUpdateDNSRecord() error {
	ip, err := S.getIP()
	if err != nil {
		return errors.Wrap(err, "getting IP")
	}

	output, err := S.Route53Client.ListResourceRecordSets(context.TODO(), &route53.ListResourceRecordSetsInput{
		HostedZoneId:    &S.DNSZoneID,
		MaxItems:        aws.Int32(1),
		StartRecordName: aws.String(S.Name + "." + os.Getenv("DNS_ZONE")),
		StartRecordType: "A",
	})
	if err != nil {
		return errors.Wrap(err, "listing recordsets")
	}

	if len(output.ResourceRecordSets) == 0 {
		return errors.Wrap(S.modifyDNSRecord(ip, route53Types.ChangeActionCreate), "creating DNS record")
	}
	return errors.Wrap(S.modifyDNSRecord(ip, route53Types.ChangeActionUpsert), "updating DNS record")
}

func (S Server) getIP() (string, error) {
	ip := ""
	start := time.Now()

	for {
		if time.Now().After(start.Add(timeout)) {
			return ip, errors.New("timeout waiting for server to get IP")
		}

		task, err := S.getRunningTask()
		fmt.Printf("loop: task: %+v err: %v\n", task, err)

		if task == nil {
			continue // no running task yet
		}

		fmt.Print("looping over attachments\n")
		for _, attachment := range task.Attachments {
			for _, detail := range attachment.Details {
				fmt.Printf("looping over detail: %+v for attachment: %+v\n", detail, attachment)
				if *detail.Name == "networkInterfaceId" && detail.Value != nil && *detail.Value != "" {
					for {
						if time.Now().After(start.Add(timeout)) {
							return ip, errors.New("timeout waiting for server to get IP")
						}

						output, err := S.Ec2Client.DescribeNetworkInterfaces(context.TODO(), &ec2.DescribeNetworkInterfacesInput{
							NetworkInterfaceIds: []string{*detail.Value},
						})
						if err != nil {
							fmt.Printf("DescribeNetworkInterfaces got error: %v\n", err)
							continue
						}

						if len(output.NetworkInterfaces) == 0 {
							fmt.Println("no network interfaces yet")
							continue
						}

						if output.NetworkInterfaces[0].Association == nil {
							fmt.Println("association is still nil")
							continue
						}

						fmt.Printf("network interfaces: %+v\n", output.NetworkInterfaces)
						return *output.NetworkInterfaces[0].Association.PublicIp, nil
					}
				}
			}
		}
		time.Sleep(1 * time.Second)
	}
}

func (S Server) Start() error {
	task, err := S.getRunningTask()
	if err != nil {
		return errors.Wrap(err, "getting running task")
	}
	if task != nil {
		return errors.New(fmt.Sprintf("found already running task with ARN: %s", *task.TaskArn))
	}

	fmt.Print("creating task set\n")
	output, err := S.EcsClient.RunTask(context.TODO(), &ecs.RunTaskInput{
		Cluster:              &S.Cluster,
		TaskDefinition:       &S.Name,
		EnableECSManagedTags: true,
		//EnableExecuteCommand: true,
		Count: aws.Int32(1),
		NetworkConfiguration: &ecsTypes.NetworkConfiguration{
			AwsvpcConfiguration: &ecsTypes.AwsVpcConfiguration{
				AssignPublicIp: "ENABLED",
				Subnets:        strings.Split(os.Getenv("SUBNETS"), ","),
				SecurityGroups: nil,
			},
		},
	})
	fmt.Printf("RunTask output: %+v with error: %v\n", output, err)
	if err != nil {
		return errors.Wrap(err, "creating task set")
	}

	fmt.Print("creating DNS record\n")
	return errors.Wrap(S.createOrUpdateDNSRecord(), "setting DNS record")
}

package catalog

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	ecsTypes "github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	route53Types "github.com/aws/aws-sdk-go-v2/service/route53/types"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/pkg/errors"
	"time"
)

var timeout = 20 * time.Second

type Server struct {
	Name          string // task definition family name
	Cluster       string
	DNSZoneID     string
	EcsClient     ecsClient
	Route53Client route53Client
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

	return &ServerStatus{
		Name:   S.Name,
		Status: *task.LastStatus,
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

	_, err = S.EcsClient.StopTask(context.TODO(), &ecs.StopTaskInput{
		Task:    task.TaskArn,
		Cluster: &S.Cluster,
	})
	if err != nil {
		return errors.Wrap(err, "failed to stop task")
	}

	return errors.Wrap(S.deleteDNSRecord(), "deleting DNS record")
}

func (S Server) deleteDNSRecord() error {
	output, _ := S.Route53Client.ListResourceRecordSets(context.TODO(), &route53.ListResourceRecordSetsInput{
		HostedZoneId:    &S.DNSZoneID,
		MaxItems:        aws.Int32(1),
		StartRecordName: &S.Name,
		StartRecordType: "A",
	})

	if len(output.ResourceRecordSets) == 0 {
		return nil
	}
	return errors.Wrap(S.modifyDNSRecord("", route53Types.ChangeActionDelete), "deleting DNS record")
}

func (S Server) modifyDNSRecord(ip string, action route53Types.ChangeAction) error {
	record := route53Types.ResourceRecordSet{
		Name: &S.Name,
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

func (S Server) createOrUpdateDNSRecord(ip string) error {
	output, _ := S.Route53Client.ListResourceRecordSets(context.TODO(), &route53.ListResourceRecordSetsInput{
		HostedZoneId:    &S.DNSZoneID,
		MaxItems:        aws.Int32(1),
		StartRecordName: &S.Name,
		StartRecordType: "A",
	})
	if len(output.ResourceRecordSets) == 0 {
		return errors.Wrap(S.modifyDNSRecord(ip, route53Types.ChangeActionCreate), "creating DNS record")
	}
	return errors.Wrap(S.modifyDNSRecord(ip, route53Types.ChangeActionUpsert), "updating DNS record")
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
	output, err := S.EcsClient.CreateTaskSet(context.TODO(), &ecs.CreateTaskSetInput{
		Cluster:        &S.Cluster,
		TaskDefinition: &S.Name,
	})
	fmt.Printf("creatTaskSet output: %+v with error: %v\n", output, err)
	if err != nil {
		return errors.Wrap(err, "creating task set")
	}

	ip := ""
	start := time.Now()
	for {
		if time.Now().After(start.Add(timeout)) {
			return errors.New("timeout waiting for server to get IP")
		}
		task, err = S.getRunningTask()
		fmt.Printf("loop: task: %+v err: %v\n", task, err)
		if task == nil {
			continue // no running task yet
		}
		fmt.Print("looping over containers\n")
		for _, container := range task.Containers {
			for _, binding := range container.NetworkBindings {
				fmt.Printf("looping over binding: %+v for container: %+v\n", binding, container)
				if *binding.BindIP != "" {
					ip = *binding.BindIP
					break
				}
			}
			if ip != "" {
				break
			}
		}
		if ip != "" {
			break
		}
	}
	fmt.Print("creating DNS record\n")
	return errors.Wrap(S.createOrUpdateDNSRecord(ip), "setting DNS record")
}

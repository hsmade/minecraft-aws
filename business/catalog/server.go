package catalog

import (
	"context"
	_ "embed"
	"encoding/base64"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2Types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	efsTypes "github.com/aws/aws-sdk-go-v2/service/efs/types"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	route53Types "github.com/aws/aws-sdk-go-v2/service/route53/types"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/pkg/errors"
	"os"
	"time"
)

type Server struct {
	Name          string // task definition family name
	DNSZoneID     string
	FileSystem    efsTypes.FileSystemDescription
	Tags          map[string]string
	Route53Client route53Client
	Ec2Client     ec2Client
	EfsClient     efsClient
}

type ServerStatus struct {
	Name          string
	InstanceState string
	IP            string
}

//go:embed metadata.sh
var metadata []byte

func (S Server) getRunningInstance() (*ec2Types.Instance, error) {
	output, err := S.Ec2Client.DescribeInstances(context.TODO(), &ec2.DescribeInstancesInput{
		Filters: []ec2Types.Filter{
			{
				Name:   aws.String("tag:Name"),
				Values: []string{S.Name},
			},
		},
	})
	if err != nil {
		return nil, errors.Wrapf(err, "listing EC2 instances with name %s", S.Name)
	}

	for _, reservation := range output.Reservations {
		for _, instance := range reservation.Instances {
			fmt.Printf("found instance %s with state %s\n", *instance.InstanceId, instance.State.Name)
			if instance.State.Name == "running" {
				return &instance, nil
			}
		}
	}
	fmt.Println("no instances found")
	return nil, nil
}

// Status finds a running instance and returns its info
func (S Server) Status() (*ServerStatus, error) {
	fmt.Printf("getting status for server with name '%s'\n", S.Name)
	instance, err := S.getRunningInstance()
	if err != nil {
		return nil, errors.Wrap(err, "getting running instance")
	}
	if instance == nil {
		return nil, errors.New("no running instance found")
	}

	return &ServerStatus{
		Name:          S.Name,
		InstanceState: string(instance.State.Name),
		IP:            *instance.PublicIpAddress,
	}, nil
	// FIXME: add status check status (Initializing / ...?)
}

// Stop will terminate the running instance
func (S Server) Stop() error {
	instance, err := S.getRunningInstance()
	if err != nil {
		return errors.Wrap(err, "getting running instance")
	}
	if instance == nil {
		return errors.New("no running instance found")
	}

	errDNS := S.deleteDNSRecord() // needs to happen first, save error for later

	_, err = S.Ec2Client.TerminateInstances(context.TODO(), &ec2.TerminateInstancesInput{
		InstanceIds: []string{*instance.InstanceId},
	})
	if err != nil {
		return errors.Wrap(err, "failed to stop instance")
	}

	return errors.Wrap(errDNS, "deleting DNS record")
}

func (S Server) deleteDNSRecord() error {
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
	if len(output.ResourceRecordSets[0].ResourceRecords) == 0 {
		return nil
	}
	return errors.Wrap(S.modifyDNSRecord(*output.ResourceRecordSets[0].ResourceRecords[0].Value, route53Types.ChangeActionDelete), "deleting DNS record")
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

func (S Server) createOrUpdateDNSRecord(ip string) error {
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

func (S Server) Start() error {
	instance, err := S.getRunningInstance()
	if err != nil {
		return errors.Wrap(err, "getting running instance")
	}
	if instance != nil {
		return errors.New(fmt.Sprintf("found already running instance with ID: %s", *instance.InstanceId))
	}

	fmt.Print("creating instance\n")
	// find SG
	// find SSM role
	// find subnet
	// get instance type from env
	// find ubuntu image / get from env
	result, err := S.Ec2Client.RunInstances(context.TODO(), &ec2.RunInstancesInput{
		BlockDeviceMappings: nil,
		IamInstanceProfile: &ec2Types.IamInstanceProfileSpecification{
			Name: aws.String("ssm"),
		},
		ImageId:                           aws.String("ami-06d94a781b544c133"), // FIXME: get from env?
		InstanceInitiatedShutdownBehavior: "terminate",
		InstanceType:                      "t2.medium", // FIXME: get from env?
		SecurityGroups:                    []string{"minecraft"},
		MaxCount:                          aws.Int32(1),
		MinCount:                          aws.Int32(1),
		TagSpecifications: []ec2Types.TagSpecification{
			{
				ResourceType: ec2Types.ResourceTypeInstance,
				Tags: []ec2Types.Tag{
					{
						Key:   aws.String("Name"),
						Value: aws.String(S.Name),
					},
				},
			},
		},
		UserData: aws.String(base64.StdEncoding.EncodeToString(metadata)),
	})

	if err != nil {
		return errors.Wrap(err, "creating instance")
	}

	instance = &result.Instances[0] // FIXME: possible runtime error
	fmt.Printf("created instance with id %s\n", *instance.InstanceId)

	var IP *string
	startTime := time.Now()
	// FIXME: takes too long
	for {
		if time.Now().After(startTime.Add(time.Second * 30)) {
			return errors.New("timeout waiting for new instance")
		}
		time.Sleep(time.Millisecond * 250)
		output, err := S.Ec2Client.DescribeInstances(context.TODO(), &ec2.DescribeInstancesInput{
			Filters: []ec2Types.Filter{
				{
					Name:   aws.String("instance-id"),
					Values: []string{*instance.InstanceId},
				},
			},
		})
		if err != nil {
			fmt.Printf("failed to list new instance: %v\n", err)
			continue
		}

		if len(output.Reservations) != 1 {
			continue
		}
		if len(output.Reservations[0].Instances) != 1 {
			continue
		}
		if output.Reservations[0].Instances[0].PublicIpAddress == nil {
			continue
		}
		IP = output.Reservations[0].Instances[0].PublicIpAddress
		break
	}

	fmt.Print("creating DNS record\n")
	return errors.Wrap(S.createOrUpdateDNSRecord(*IP), "setting DNS record")
}

package catalog

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/base64"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2Types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/efs"
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
	Name             string
	InstanceState    string
	HealthcheckState string
	IP               string
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
			//fmt.Printf("found instance %s with state %s\n", *instance.InstanceId, instance.State.Name)
			if instance.State.Name == "running" || instance.State.Name == "pending" {
				return &instance, nil
			}
		}
	}
	//fmt.Println("no instances found")
	return nil, nil
}

// getHealthState returns the latest health check status for an instance
func (S Server) getHealthState(instanceId string) (string, error) {
	status := &ec2.DescribeInstanceStatusOutput{}
	var err error

	for {
		status, err = S.Ec2Client.DescribeInstanceStatus(context.TODO(), &ec2.DescribeInstanceStatusInput{
			InstanceIds: []string{instanceId},
			NextToken:   status.NextToken,
		})
		if err != nil {
			return "unknown", errors.Wrap(err, "getting instance status")
		}

		// we want to loop until we reach the last entry
		if status.NextToken == nil {
			break
		}
	}

	if len(status.InstanceStatuses) > 0 {
		// return last info we have
		return string(status.InstanceStatuses[len(status.InstanceStatuses)-1].InstanceStatus.Status), nil
	}

	return "unknown", nil
}

// Status finds a running instance and returns its info
func (S Server) Status() (*ServerStatus, error) {
	//fmt.Printf("getting status for server with name '%s'\n", S.Name)
	instance, err := S.getRunningInstance()
	if err != nil {
		return nil, errors.Wrap(err, "getting running instance")
	}
	if instance == nil {
		return nil, errors.New("no running instance found")
	}

	healthCheckState, _ := S.getHealthState(*instance.InstanceId)

	serverStatus := ServerStatus{
		Name:             S.Name,
		InstanceState:    string(instance.State.Name),
		HealthcheckState: healthCheckState,
		IP:               *instance.PublicIpAddress,
	}

	return &serverStatus, nil
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

	errDNS := S.deleteDNSRecord() // want to do this first, so it doesn't depend on instance termination

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

	if len(output.ResourceRecordSets) == 0 || len(output.ResourceRecordSets[0].ResourceRecords) == 0 {
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

	// find EFS for name, and get ID to add to metadata
	fsId, err := S.getEfsId()
	if err != nil {
		return errors.Wrap(err, "getting EFS ID")
	}

	fmt.Println("creating instance")
	result, err := S.Ec2Client.RunInstances(context.TODO(), &ec2.RunInstancesInput{
		BlockDeviceMappings: nil,
		IamInstanceProfile: &ec2Types.IamInstanceProfileSpecification{
			Name: aws.String("ssm"),
		},
		ImageId:                           aws.String("ami-06d94a781b544c133"), // FIXME: get from env?
		InstanceInitiatedShutdownBehavior: "terminate",
		InstanceType:                      "t3a.medium", // FIXME: get from env?
		SecurityGroups:                    []string{"minecraft", "minecraft-efs"},
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
		// add the userdata from metadata.sh, and replace the FSID with the actual EFS FS ID
		UserData: aws.String(base64.StdEncoding.EncodeToString(bytes.Replace(metadata, []byte("FSID"), []byte(fsId), 1))),
	})
	if err != nil {
		return errors.Wrap(err, "creating instance")
	}

	fmt.Println("waiting for IP")
	IP, err := S.waitForPublicIP(*result.Instances[0].InstanceId, 60*time.Second)
	if err != nil {
		return errors.Wrap(err, "getting IP for DNS record")
	}

	fmt.Printf("Updating DNS with ip %s\n", IP)
	return errors.Wrap(S.createOrUpdateDNSRecord(IP), "setting DNS record")
}

// waitForPublicIP waits until the public IP is available, and returns it
func (S Server) waitForPublicIP(instanceId string, timeoutDuration time.Duration) (string, error) {
	var IP *string
	timeout := time.Now().Add(timeoutDuration)

	for {
		time.Sleep(time.Second * 1)
		if time.Now().After(timeout) {
			return "", errors.New("timeout waiting for new instance")
		}

		output, err := S.Ec2Client.DescribeInstances(context.TODO(), &ec2.DescribeInstancesInput{
			Filters: []ec2Types.Filter{
				{
					Name:   aws.String("instance-id"),
					Values: []string{instanceId},
				},
			},
		})
		if err != nil {
			//fmt.Printf("failed to list new instance: %v\n", err)
			continue
		}

		// find running or pending
		for _, reservation := range output.Reservations {
			for _, instance := range reservation.Instances {
				if instance.State.Name == "pending" || instance.State.Name == "running" {
					IP = instance.PublicIpAddress
					break
				}
			}
		}

		if IP != nil {
			return *IP, nil
		}
	}
}

func (S Server) getEfsId() (string, error) {
	filesystems, err := S.EfsClient.DescribeFileSystems(context.TODO(), &efs.DescribeFileSystemsInput{})
	if err != nil {
		return "", errors.Wrap(err, "listing EFS filesystems")
	}

	for _, filesystem := range filesystems.FileSystems {
		if *filesystem.Name == S.Name {
			return *filesystem.FileSystemId, nil
		}
	}
	return "", errors.New(fmt.Sprintf("EFS Filesystem not found with name '%s'", S.Name))
}

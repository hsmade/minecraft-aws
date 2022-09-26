package catalog

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	ecsTypes "github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	"github.com/pkg/errors"
	"os"
)

type Servers struct {
	Cluster       string
	DNSZoneID     string
	EcsClient     ecsClient
	Route53Client route53Client
	Ec2Client     ec2Client
}

func New() (*Servers, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, errors.Wrap(err, "loading AWS config")
	}
	cluster := os.Getenv("CLUSTER")
	if cluster == "" {
		return nil, errors.New("Missing CLUSTER variable")
	}

	dnsZoneID := os.Getenv("DNS_ZONE_ID")
	if cluster == "" {
		return nil, errors.New("Missing DNS_ZONE_ID variable")
	}

	return &Servers{
		Cluster:       cluster,
		DNSZoneID:     dnsZoneID,
		EcsClient:     ecs.NewFromConfig(cfg),
		Route53Client: route53.NewFromConfig(cfg),
		Ec2Client:     ec2.NewFromConfig(cfg),
	}, nil
}

// ListServers gets the list of task definitions, and returns server instances
func (S Servers) ListServers() ([]*Server, error) {
	families, err := S.EcsClient.ListTaskDefinitionFamilies(context.TODO(), &ecs.ListTaskDefinitionFamiliesInput{})
	if err != nil {
		return nil, errors.Wrap(err, "getting task definition families")
	}
	var servers []*Server
	for _, name := range families.Families {
		output, err := S.EcsClient.DescribeTaskDefinition(context.TODO(), &ecs.DescribeTaskDefinitionInput{
			TaskDefinition: &name,
			Include:        []ecsTypes.TaskDefinitionField{"TAGS"},
		})
		if err != nil {
			fmt.Printf("could not get task definition for family '%s': %v\n", name, err)
		}
		fmt.Printf("task definition for family '%s': %+v\n", name, output)

		tags := make(map[string]string, len(output.Tags))
		for _, tag := range output.Tags {
			tags[*tag.Key] = *tag.Value
		}

		servers = append(servers, &Server{
			Name:          name,
			Cluster:       S.Cluster,
			DNSZoneID:     S.DNSZoneID,
			Tags:          tags,
			EcsClient:     S.EcsClient,
			Route53Client: S.Route53Client,
			Ec2Client:     S.Ec2Client,
		})
	}

	return servers, nil
}

// GetServer returns a server's instance
func (S Servers) GetServer(name string) (*Server, error) {
	// FIXME: check if server/task family exists
	return &Server{
		Name:          name,
		Cluster:       S.Cluster,
		DNSZoneID:     S.DNSZoneID,
		EcsClient:     S.EcsClient,
		Route53Client: S.Route53Client,
		Ec2Client:     S.Ec2Client,
	}, nil
}

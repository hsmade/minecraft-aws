package catalog

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	"github.com/pkg/errors"
	"os"
)

type Servers struct {
	Cluster       string
	DNSZoneID     string
	EcsClient     ecsClient
	Route53Client route53Client
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
		Route53Client: route53.New(route53.Options{}),
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
		servers = append(servers, &Server{
			Name:          name,
			Cluster:       S.Cluster,
			DNSZoneID:     S.DNSZoneID,
			EcsClient:     S.EcsClient,
			Route53Client: S.Route53Client,
		})
	}

	return servers, nil
}
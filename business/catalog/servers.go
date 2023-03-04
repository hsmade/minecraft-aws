package catalog

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/efs"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	"github.com/pkg/errors"
	"os"
)

type Servers struct {
	Cluster       string
	DNSZoneID     string
	Route53Client route53Client
	Ec2Client     ec2Client
	EfsClient     efsClient
}

func New() (*Servers, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, errors.Wrap(err, "loading AWS config")
	}

	dnsZoneID := os.Getenv("DNS_ZONE_ID")
	if dnsZoneID == "" {
		return nil, errors.New("Missing DNS_ZONE_ID variable")
	}

	return &Servers{
		DNSZoneID:     dnsZoneID,
		Route53Client: route53.NewFromConfig(cfg),
		Ec2Client:     ec2.NewFromConfig(cfg),
		EfsClient:     efs.NewFromConfig(cfg), // efs.New(mySession, aws.NewConfig().WithRegion("us-west-2"))
	}, nil
}

// ListServers gets the list of EFS shares
func (S Servers) ListServers() ([]*Server, error) {
	input := &efs.DescribeFileSystemsInput{}
	result, err := S.EfsClient.DescribeFileSystems(context.TODO(), input)
	if err != nil {
		return nil, errors.Wrap(err, "getting EFS filesystems")
	}

	var servers []*Server
	for _, fileSystem := range result.FileSystems {
		tags := make(map[string]string, len(fileSystem.Tags))
		for _, tag := range fileSystem.Tags {
			tags[*tag.Key] = *tag.Value
		}

		servers = append(servers, &Server{
			Name:          tags["Name"], // FIXME: runtime error
			DNSZoneID:     S.DNSZoneID,
			Tags:          tags,
			FileSystem:    fileSystem,
			Route53Client: S.Route53Client,
			Ec2Client:     S.Ec2Client,
			EfsClient:     S.EfsClient,
		})
	}

	return servers, nil
}

// GetServer returns a server's instance
func (S Servers) GetServer(name string) (*Server, error) {
	return &Server{
		Name:          name,
		DNSZoneID:     S.DNSZoneID,
		Route53Client: S.Route53Client,
		Ec2Client:     S.Ec2Client,
		EfsClient:     S.EfsClient,
	}, nil
}

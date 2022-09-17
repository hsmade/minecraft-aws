package storage

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/pkg/errors"
	"strings"
)

type Servers struct {
	bucket string // the S3 storage bucket to use
	region string // in which region S3 lives
	client ClientModel
}

type ClientModel interface {
	ListObjectsV2(ctx context.Context, params *s3.ListObjectsV2Input, optFns ...func(*s3.Options)) (*s3.ListObjectsV2Output, error)
}

type Server struct {
	Name       string
	ImageUrl   string // URL to thumbnail
	ArchiveUrl string // URL to archive
}

func New(bucket string) (*Servers, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, errors.Wrap(err, "getting client for S3")
	}
	client := s3.NewFromConfig(cfg)
	return &Servers{bucket: bucket, client: client, region: cfg.Region}, nil
}

func (S *Servers) List() ([]Server, error) {
	result, err := S.client.ListObjectsV2(context.TODO(), &s3.ListObjectsV2Input{
		Bucket: aws.String(S.bucket),
	})
	if err != nil {
		return nil, errors.Wrap(err, "getting contents of bucket")
	}

	var servers []Server
	for _, object := range result.Contents {
		fileName := aws.ToString(object.Key)
		if !strings.HasSuffix(fileName, ".tgz") {
			continue
		}
		serverName := strings.TrimSuffix(fileName, ".tgz")
		servers = append(servers, S.Get(serverName))
	}
	return servers, nil
}

func (S *Servers) Get(name string) Server {
	bucketBaseUrl := fmt.Sprintf("https://%s.s3.%s.amazonaws.com/", S.bucket, S.region)
	return Server{
		Name:       name,
		ImageUrl:   bucketBaseUrl + name + ".png",
		ArchiveUrl: bucketBaseUrl + name + ".tgz",
	}
}

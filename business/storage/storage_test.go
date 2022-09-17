package storage

// go:generate mockgen -destination mocks.go -package storage . ClientModel

import (
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/golang/mock/gomock"
	"gotest.tools/assert"
	"reflect"
	"strings"
	"testing"
)

func TestList(t *testing.T) {
	tests := []struct {
		name         string
		want         []Server
		wantError    bool
		listResponse []string
		listError    error
	}{
		{
			name: "happy path: 2 servers",
			want: []Server{
				{
					Name:       "server1",
					ImageUrl:   "https://bucket.s3.region.amazonaws.com/server1.png",
					ArchiveUrl: "https://bucket.s3.region.amazonaws.com/server1.tgz",
				},
				{
					Name:       "server2",
					ImageUrl:   "https://bucket.s3.region.amazonaws.com/server2.png",
					ArchiveUrl: "https://bucket.s3.region.amazonaws.com/server2.tgz",
				},
			},
			wantError: false,
			listError: nil,
			listResponse: []string{
				"server1.png",
				"server1.tgz",
				"server2.png",
				"server2.tgz",
				"index.html",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			client := NewMockClientModel(ctrl)
			s := Servers{client: client, bucket: "bucket", region: "region"}

			listResponse := s3.ListObjectsV2Output{}
			for _, key := range tt.listResponse {
				keyString := strings.Clone(key)
				listResponse.Contents = append(listResponse.Contents, types.Object{Key: &keyString})
			}
			client.EXPECT().
				ListObjectsV2(gomock.Any(), gomock.Any()).
				Return(&listResponse, tt.listError)

			got, err := s.List()
			if err != nil != tt.wantError {
				t.Errorf("List() returned unexpected err: %v", err)
			}
			assert.DeepEqual(t, tt.want, got)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("List() = %v, want %v", got, tt.want)
			}
		})
	}
}

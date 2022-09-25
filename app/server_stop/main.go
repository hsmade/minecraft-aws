package main

import (
	"context"
	"encoding/json"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	"minecraft-catalog/business/catalog"
	"net/http"
)

func wrapError(status int, err error) (events.APIGatewayProxyResponse, error) {
	return events.APIGatewayProxyResponse{
		StatusCode: status,
		Body:       err.Error(),
	}, nil
}

func HandleRequest(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return wrapError(http.StatusInternalServerError, err)
	}

	ecsClient := ecs.NewFromConfig(cfg)
	route53Client := route53.New(route53.Options{})

	server := catalog.Server{
		Name:          req.QueryStringParameters["name"],
		EcsClient:     ecsClient,
		Route53Client: route53Client,
	}
	err = server.Stop()

	body, err := json.Marshal(err)
	if err != nil {
		return wrapError(http.StatusInternalServerError, err)
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Headers:    map[string]string{"Content-Type": "application/json"},
		Body:       string(body),
	}, nil
}

func main() {
	lambda.Start(HandleRequest)
}

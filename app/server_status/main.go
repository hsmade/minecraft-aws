package main

import (
	"encoding/json"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"minecraft-catalog/business/catalog"
	"net/http"
	"os"
)

func wrapError(status int, err error) (events.APIGatewayProxyResponse, error) {
	return events.APIGatewayProxyResponse{
		StatusCode: status,
		Body:       err.Error(),
	}, nil
}

func HandleRequest(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	servers, err := catalog.New()
	if err != nil {
		return wrapError(http.StatusInternalServerError, err)
	}

	server, err := servers.GetServer(req.QueryStringParameters["name"])
	if err != nil {
		return wrapError(http.StatusInternalServerError, err)
	}

	status, err := server.Status()
	if err != nil {
		return wrapError(http.StatusInternalServerError, err)
	}

	body, err := json.Marshal(status)
	if err != nil {
		return wrapError(http.StatusInternalServerError, err)
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Headers: map[string]string{
			"Content-Type":                 "application/json",
			"Access-Control-Allow-Headers": "Content-Type,X-Amz-Date,Authorization,X-Api-Key,X-Amz-Security-Token",
			"Access-Control-Allow-Methods": "GET",
			"Access-Control-Allow-Origin":  os.Getenv("CORS_DOMAIN"),
		},
		Body: string(body),
	}, nil
}

func main() {
	lambda.Start(HandleRequest)
	//HandleRequest(events.APIGatewayProxyRequest{
	//	QueryStringParameters: map[string]string{"name": "test"},
	//})
}

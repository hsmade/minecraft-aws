package main

import (
	"encoding/json"
	"fmt"
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

type Response struct {
	Name          string            `json:"name"`
	Status        string            `json:"status"`
	DesiredStatus string            `json:"desired_status"`
	Tags          map[string]string `json:"tags"`
}

func HandleRequest(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	servers, err := catalog.New()
	if err != nil {
		return wrapError(http.StatusInternalServerError, err)
	}

	serverList, err := servers.ListServers()
	if err != nil {
		return wrapError(http.StatusInternalServerError, err)
	}

	var responses []Response
	for _, server := range serverList {
		response := Response{
			Name:          server.Name,
			Tags:          server.Tags,
			Status:        "NONE",
			DesiredStatus: "NONE",
		}

		fmt.Printf("checking status for server '%s'\n", server.Name)
		status, err := server.Status()
		if err == nil {
			response.Status = status.Status
			response.DesiredStatus = status.DesiredStatus
		}
		responses = append(responses, response)
	}

	body, err := json.Marshal(responses)
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
}

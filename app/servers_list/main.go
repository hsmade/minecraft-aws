package main

import (
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"minecraft-catalog/business/catalog"
	"net/http"
)

func wrapError(status int, err error) (events.APIGatewayProxyResponse, error) {
	return events.APIGatewayProxyResponse{
		StatusCode: status,
		Body:       err.Error(),
	}, nil
}

type Response struct {
	Name   string            `json:"name"`
	Status string            `json:"status"`
	Tags   map[string]string `json:"tags"`
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
			Name:   server.Name,
			Status: "NONE",
		}

		fmt.Printf("checking status for server '%s'\n", server.Name)
		status, err := server.Status()
		if err == nil {
			response.Status = status.Status
		}
		responses = append(responses, response)
	}

	body, err := json.Marshal(responses)
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

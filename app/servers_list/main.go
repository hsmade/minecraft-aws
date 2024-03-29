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
	Name             string            `json:"name"`
	InstanceState    string            `json:"instance_state"`
	HealthCheckState string            `json:"health_check_state"`
	Tags             map[string]string `json:"tags"`
	IP               string            `json:"ip"`
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
			Name:             server.Name,
			Tags:             server.Tags,
			InstanceState:    "NONE",
			HealthCheckState: "NONE",
		}

		fmt.Printf("checking status for server '%s'\n", server.Name)
		status, err := server.Status()
		if err == nil {
			response.InstanceState = status.InstanceState
			response.HealthCheckState = status.HealthcheckState
			response.IP = status.IP
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
	//HandleRequest(events.APIGatewayProxyRequest{})
}

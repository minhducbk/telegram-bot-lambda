package main

import (
	"fmt"
	"log"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"

	"telegram-bot/telegram"
)


func hanlder(event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	log.Println("Hello World", event)
	message := fmt.Sprintf("Resource %s. Body: %s. Path: %s", event.Resource, event.Body, event.Path)
	telegram.SendToTelegramChannel(message)

	return events.APIGatewayProxyResponse{StatusCode: 200, Body: event.Body}, nil
}

func main() {
	lambda.Start(hanlder)
}

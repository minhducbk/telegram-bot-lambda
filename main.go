package main

import (
	"log"
	"telegram-bot/telegram"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)


func hanlder(event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	log.Println("Hello World", event)
	telegram.SendToTelegramChannel("")
	return events.APIGatewayProxyResponse{StatusCode: 200, Body: event.Body}, nil
}

func main() {
	lambda.Start(hanlder)
}

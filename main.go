package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"

	"telegram-bot/pkg/ip"
	"telegram-bot/pkg/telegram"
)


func lambdaHanlder(event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	log.Println("Hello World", event)
	telegram.SendToTelegramChannel(event.Body)

	return events.APIGatewayProxyResponse{StatusCode: 200, Body: event.Body}, nil
}

func testRedisConnectivity() {
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		log.Fatal("REDIS_URL environment variable not set")
	}
	log.Println("Testing connectivity to Redis URL: ", redisURL)

	conn, err := net.DialTimeout("tcp", redisURL, 5*time.Second)
	if err != nil {
		log.Fatalf("Could not connect to Redis: %v", err)
	}
	defer conn.Close()

	log.Println("Successfully connected to Redis")
}

func ipHandler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	ip, err := ip.GetPublicIP()
	if err != nil {
		log.Fatalf("Error getting public IP: %v", err)
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       fmt.Sprintf("Error getting public IP: %v", err),
		}, nil
	}

	response := fmt.Sprintf("Public IP: %s", ip)
	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Body:       response,
	}, nil
}

func main() {
	// lambda.Start(func(ctx context.Context) error {
	// 	testRedisConnectivity()
	// 	return nil
	// })
	// lambda.Start(hanlder)
	// Load .env file

	// err := godotenv.Load()
	// if err != nil {
	// 	log.Fatalf("Error loading .env file")
	// 	}
	// trader := trader.NewBinanceTrader(os.Getenv("BINANCE_API_KEY"), os.Getenv("BINANCE_SECRET_KEY"))
	// otherMessage := fmt.Sprintf("Balances: \n")
	// balances := trader.GetBalances()
	// for asset, balance := range balances {
	// 	otherMessage += fmt.Sprintf("%s %s", balance, asset)
	// }
	// fmt.Println(otherMessage)
	// symbol := "MATICUSDT"
	// trader.PlaceMarketBuyOrder(symbol, "10")

	lambda.Start(lambdaHanlder)
}

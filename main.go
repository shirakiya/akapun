package main

import (
	"context"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

func HandleRequest(ctx context.Context, event events.IoTButtonEvent) (string, error) {
	fmt.Printf("%#v\n", event)

	return "OK", nil
}

func main() {
	lambda.Start(HandleRequest)
}

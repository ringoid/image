package main

import (
	"context"
	basicLambda "github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"os"
	"fmt"
	"github.com/aws/aws-sdk-go/aws/session"
	"encoding/json"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambdacontext"
	"errors"
	"github.com/ringoid/commons"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	"github.com/aws/aws-dax-go/dax"
)

var anlogger *commons.Logger
var daxClient dynamodbiface.DynamoDBAPI
var userPhotoTable string

var asyncTaskQueue string
var awsSqsClient *sqs.SQS

func init() {
	var env string
	var ok bool
	var papertrailAddress string
	var err error
	var awsSession *session.Session

	env, ok = os.LookupEnv("ENV")
	if !ok {
		fmt.Printf("lambda-initialization : handle_stream.go : env can not be empty ENV\n")
		os.Exit(1)
	}
	fmt.Printf("lambda-initialization : handle_stream.go : start with ENV = [%s]\n", env)

	papertrailAddress, ok = os.LookupEnv("PAPERTRAIL_LOG_ADDRESS")
	if !ok {
		fmt.Printf("lambda-initialization : handle_stream.go : env can not be empty PAPERTRAIL_LOG_ADDRESS\n")
		os.Exit(1)
	}
	fmt.Printf("lambda-initialization : handle_stream.go : start with PAPERTRAIL_LOG_ADDRESS = [%s]\n", papertrailAddress)

	anlogger, err = commons.New(papertrailAddress, fmt.Sprintf("%s-%s", env, "internal-handle-stream-image"))
	if err != nil {
		fmt.Errorf("lambda-initialization : handle_stream.go : error during startup : %v\n", err)
		os.Exit(1)
	}
	anlogger.Debugf(nil, "lambda-initialization : handle_stream.go : logger was successfully initialized")

	userPhotoTable, ok = os.LookupEnv("USER_PHOTO_TABLE")
	if !ok {
		anlogger.Fatalf(nil, "lambda-initialization : handle_stream.go : env can not be empty USER_PHOTO_TABLE")
	}
	anlogger.Debugf(nil, "lambda-initialization : handle_stream.go : start with USER_PHOTO_TABLE = [%s]", userPhotoTable)

	asyncTaskQueue, ok = os.LookupEnv("ASYNC_TASK_SQS_QUEUE")
	if !ok {
		anlogger.Fatalf(nil, "lambda-initialization : handle_stream.go : env can not be empty ASYNC_TASK_SQS_QUEUE")
	}
	anlogger.Debugf(nil, "lambda-initialization : handle_stream.go : start with ASYNC_TASK_SQS_QUEUE = [%s]", asyncTaskQueue)

	awsSession, err = session.NewSession(aws.NewConfig().
		WithRegion(commons.Region).WithMaxRetries(commons.MaxRetries).
		WithLogger(aws.LoggerFunc(func(args ...interface{}) { anlogger.AwsLog(args) })).WithLogLevel(aws.LogOff))
	if err != nil {
		anlogger.Fatalf(nil, "lambda-initialization : handle_stream.go : error during initialization : %v", err)
	}
	anlogger.Debugf(nil, "lambda-initialization : handle_stream.go : aws session was successfully initialized")

	daxEndpoint, ok := os.LookupEnv("DAX_ENDPOINT")
	if !ok {
		anlogger.Fatalf(nil, "lambda-initialization : handle_stream.go : env can not be empty DAX_ENDPOINT")
	}
	cfg := dax.DefaultConfig()
	cfg.HostPorts = []string{daxEndpoint}
	cfg.Region = commons.Region
	daxClient, err = dax.New(cfg)
	if err != nil {
		anlogger.Fatalf(nil, "lambda-initialization : handle_stream.go : error initialize DAX cluster")
	}
	anlogger.Debugf(nil, "lambda-initialization : handle_stream.go : dax client was successfully initialized")

	awsSqsClient = sqs.New(awsSession)
	anlogger.Debugf(nil, "lambda-initialization : handle_stream.go : sqs client was successfully initialized")
}

func handler(ctx context.Context, event events.KinesisEvent) (error) {
	lc, _ := lambdacontext.FromContext(ctx)

	anlogger.Debugf(lc, "handle_stream.go : start handle request with [%d] records", len(event.Records))

	for _, record := range event.Records {
		body := record.Kinesis.Data

		var aEvent commons.BaseInternalEvent
		err := json.Unmarshal(body, &aEvent)
		if err != nil {
			anlogger.Errorf(lc, "handle_stream.go : error unmarshal body [%s] to BaseInternalEvent : %v", body, err)
			return errors.New(fmt.Sprintf("error unmarshal body %s : %v", body, err))
		}
		anlogger.Debugf(lc, "handle_stream.go : handle record %v", aEvent)

		switch aEvent.EventType {
		case commons.LikePhotoInternalEvent:
			err = likePhotoUpdate(body, userPhotoTable, daxClient, lc, anlogger)
			if err != nil {
				return err
			}
		case commons.UserDeleteHimselfEvent:
			err = deleteAllPhotos(body, userPhotoTable, asyncTaskQueue, awsSqsClient, daxClient, lc, anlogger)
			if err != nil {
				return err
			}
		}
	}

	anlogger.Debugf(lc, "handle_stream.go : successfully complete handle request with [%d] records", len(event.Records))
	return nil
}

func main() {
	basicLambda.Start(handler)
}

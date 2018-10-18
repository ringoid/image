package main

import (
	"context"
	basicLambda "github.com/aws/aws-lambda-go/lambda"
	"../sys_log"
	"../apimodel"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/aws"
	"os"
	"fmt"
	"github.com/aws/aws-sdk-go/aws/session"
	"encoding/json"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambdacontext"
	"errors"
)

var anlogger *syslog.Logger
var awsDbClient *dynamodb.DynamoDB
var userPhotoTable string

func init() {
	var env string
	var ok bool
	var papertrailAddress string
	var err error
	var awsSession *session.Session

	env, ok = os.LookupEnv("ENV")
	if !ok {
		fmt.Printf("handle_stream.go : env can not be empty ENV")
		os.Exit(1)
	}
	fmt.Printf("handle_stream.go : start with ENV = [%s]", env)

	papertrailAddress, ok = os.LookupEnv("PAPERTRAIL_LOG_ADDRESS")
	if !ok {
		fmt.Printf("handle_stream.go : env can not be empty PAPERTRAIL_LOG_ADDRESS")
		os.Exit(1)
	}
	fmt.Printf("handle_stream.go : start with PAPERTRAIL_LOG_ADDRESS = [%s]", papertrailAddress)

	anlogger, err = syslog.New(papertrailAddress, fmt.Sprintf("%s-%s", env, "internal-handle-stream-image"))
	if err != nil {
		fmt.Errorf("handle_stream.go : error during startup : %v", err)
		os.Exit(1)
	}
	anlogger.Debugf(nil, "handle_stream.go : logger was successfully initialized")

	userPhotoTable, ok = os.LookupEnv("USER_PHOTO_TABLE")
	if !ok {
		fmt.Printf("handle_stream.go : env can not be empty USER_PHOTO_TABLE")
		os.Exit(1)
	}
	anlogger.Debugf(nil, "handle_stream.go : start with USER_PHOTO_TABLE = [%s]", userPhotoTable)

	awsSession, err = session.NewSession(aws.NewConfig().
		WithRegion(apimodel.Region).WithMaxRetries(apimodel.MaxRetries).
		WithLogger(aws.LoggerFunc(func(args ...interface{}) { anlogger.AwsLog(args) })).WithLogLevel(aws.LogOff))
	if err != nil {
		anlogger.Fatalf(nil, "handle_stream.go : error during initialization : %v", err)
	}
	anlogger.Debugf(nil, "handle_stream.go : aws session was successfully initialized")

	awsDbClient = dynamodb.New(awsSession)
	anlogger.Debugf(nil, "handle_stream.go : dynamodb client was successfully initialized")

}

func handler(ctx context.Context, event events.KinesisEvent) (error) {
	lc, _ := lambdacontext.FromContext(ctx)

	anlogger.Debugf(lc, "handle_stream.go : start handle request %v", event)

	for _, record := range event.Records {
		anlogger.Debugf(lc, "handle_stream.go : handle record %v", record)
		body := record.Kinesis.Data

		var aEvent apimodel.BaseInternalEvent
		err := json.Unmarshal(body, &aEvent)
		if err != nil {
			anlogger.Errorf(lc, "handle_stream.go : error unmarshal body [%s] to BaseInternalEvent : %v", body, err)
			return errors.New(fmt.Sprintf("error unmarshal body %s : %v", body, err))
		}
		switch aEvent.EventType {
		case apimodel.LikePhotoInternalEvent:
			err = likePhoto(body, userPhotoTable, awsDbClient, lc, anlogger)
			if err != nil {
				return err
			}
		}
	}

	anlogger.Debugf(lc, "handle_stream.go : successfully complete task %v", event)
	return nil
}

func main() {
	basicLambda.Start(handler)
}

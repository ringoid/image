package main

import (
	"context"
	basicLambda "github.com/aws/aws-lambda-go/lambda"
	"../apimodel"
	"github.com/aws/aws-sdk-go/aws"
	"os"
	"fmt"
	"github.com/aws/aws-sdk-go/aws/session"
	"encoding/json"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambdacontext"
	"errors"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/ringoid/commons"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	"github.com/aws/aws-dax-go/dax"
)

var anlogger *commons.Logger
var daxClient dynamodbiface.DynamoDBAPI
var awsS3Client *s3.S3
var downloader *s3manager.Downloader
var uploader *s3manager.Uploader

func init() {
	var env string
	var ok bool
	var papertrailAddress string
	var err error
	var awsSession *session.Session

	env, ok = os.LookupEnv("ENV")
	if !ok {
		fmt.Printf("lambda-initialization : internal_handle_task.go : env can not be empty ENV\n")
		os.Exit(1)
	}
	fmt.Printf("lambda-initialization : internal_handle_task.go : start with ENV = [%s]\n", env)

	papertrailAddress, ok = os.LookupEnv("PAPERTRAIL_LOG_ADDRESS")
	if !ok {
		fmt.Printf("lambda-initialization : internal_handle_task.go : env can not be empty PAPERTRAIL_LOG_ADDRESS\n")
		os.Exit(1)
	}
	fmt.Printf("lambda-initialization : internal_handle_task.go : start with PAPERTRAIL_LOG_ADDRESS = [%s]\n", papertrailAddress)

	anlogger, err = commons.New(papertrailAddress, fmt.Sprintf("%s-%s", env, "internal-handle-task-image"))
	if err != nil {
		fmt.Errorf("lambda-initialization : internal_handle_task.go : error during startup : %v\n", err)
		os.Exit(1)
	}
	anlogger.Debugf(nil, "lambda-initialization : internal_handle_task.go : logger was successfully initialized")

	awsSession, err = session.NewSession(aws.NewConfig().
		WithRegion(commons.Region).WithMaxRetries(commons.MaxRetries).
		WithLogger(aws.LoggerFunc(func(args ...interface{}) { anlogger.AwsLog(args) })).WithLogLevel(aws.LogOff))
	if err != nil {
		anlogger.Fatalf(nil, "lambda-initialization : internal_handle_task.go : error during initialization : %v", err)
	}
	anlogger.Debugf(nil, "lambda-initialization : internal_handle_task.go : aws session was successfully initialized")

	daxEndpoint, ok := os.LookupEnv("DAX_ENDPOINT")
	if !ok {
		anlogger.Fatalf(nil, "lambda-initialization : internal_handle_task.go : env can not be empty DAX_ENDPOINT")
	}
	cfg := dax.DefaultConfig()
	cfg.HostPorts = []string{daxEndpoint}
	cfg.Region = commons.Region
	daxClient, err = dax.New(cfg)
	if err != nil {
		anlogger.Fatalf(nil, "lambda-initialization : internal_handle_task.go : error initialize DAX cluster")
	}
	anlogger.Debugf(nil, "lambda-initialization : internal_handle_task.go : dax client was successfully initialized")

	awsS3Client = s3.New(awsSession)
	anlogger.Debugf(nil, "lambda-initialization : internal_handle_task.go : s3 client was successfully initialized")

	downloader = s3manager.NewDownloader(awsSession)
	anlogger.Debugf(nil, "lambda-initialization : internal_handle_task.go : s3 downloader was successfully initialized")

	uploader = s3manager.NewUploader(awsSession)
	anlogger.Debugf(nil, "lambda-initialization : internal_handle_task.go : s3 uploader was successfully initialized")
}

func handler(ctx context.Context, event events.SQSEvent) (error) {
	lc, _ := lambdacontext.FromContext(ctx)

	anlogger.Debugf(lc, "internal_handle_task.go : start handle request with [%d] records", len(event.Records))

	for _, record := range event.Records {
		body := record.Body
		var aTask apimodel.AsyncTask
		err := json.Unmarshal([]byte(body), &aTask)
		if err != nil {
			anlogger.Errorf(lc, "internal_handle_task.go : error unmarshal body [%s] to AsyncTask : %v", body, err)
			return errors.New(fmt.Sprintf("error unmarshal body %s : %v", body, err))
		}
		anlogger.Debugf(lc, "internal_handle_task.go : handle record %v", aTask)

		switch aTask.TaskType {
		case apimodel.ImageRemovePhotoTaskType:
			err = removePhoto([]byte(body), daxClient, lc, anlogger)
			if err != nil {
				return err
			}
		case apimodel.ImageResizePhotoTaskType:
			err = resizePhoto([]byte(body), downloader, uploader, daxClient, lc, anlogger)
			if err != nil {
				return err
			}
		case apimodel.ImageRemoveS3ObjectTaskType:
			err = removeS3Object([]byte(body), awsS3Client, lc, anlogger)
			if err != nil {
				return err
			}
		default:
			return errors.New(fmt.Sprintf("unsuported taks type %s", aTask.TaskType))
		}
	}

	anlogger.Debugf(lc, "internal_handle_task.go : successfully complete handle request with [%d] records", len(event.Records))
	return nil
}

func main() {
	basicLambda.Start(handler)
}

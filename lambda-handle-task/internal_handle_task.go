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

func init() {
	var env string
	var ok bool
	var papertrailAddress string
	var err error
	var awsSession *session.Session

	env, ok = os.LookupEnv("ENV")
	if !ok {
		fmt.Printf("internal_handle_task.go : env can not be empty ENV")
		os.Exit(1)
	}
	fmt.Printf("internal_handle_task.go : start with ENV = [%s]", env)

	papertrailAddress, ok = os.LookupEnv("PAPERTRAIL_LOG_ADDRESS")
	if !ok {
		fmt.Printf("internal_handle_task.go : env can not be empty PAPERTRAIL_LOG_ADDRESS")
		os.Exit(1)
	}
	fmt.Printf("internal_handle_task.go : start with PAPERTRAIL_LOG_ADDRESS = [%s]", papertrailAddress)

	anlogger, err = syslog.New(papertrailAddress, fmt.Sprintf("%s-%s", env, "internal-handle-task-image"))
	if err != nil {
		fmt.Errorf("internal_handle_task.go : error during startup : %v", err)
		os.Exit(1)
	}
	anlogger.Debugf(nil, "internal_handle_task.go : logger was successfully initialized")

	awsSession, err = session.NewSession(aws.NewConfig().
		WithRegion(apimodel.Region).WithMaxRetries(apimodel.MaxRetries).
		WithLogger(aws.LoggerFunc(func(args ...interface{}) { anlogger.AwsLog(args) })).WithLogLevel(aws.LogOff))
	if err != nil {
		anlogger.Fatalf(nil, "internal_handle_task.go : error during initialization : %v", err)
	}
	anlogger.Debugf(nil, "internal_handle_task.go : aws session was successfully initialized")

	awsDbClient = dynamodb.New(awsSession)
	anlogger.Debugf(nil, "internal_handle_task.go : dynamodb client was successfully initialized")
}

func handler(ctx context.Context, event events.SQSEvent) (error) {
	lc, _ := lambdacontext.FromContext(ctx)

	anlogger.Debugf(lc, "internal_handle_task.go : start handle request %v", event)

	for _, record := range event.Records {
		anlogger.Debugf(lc, "internal_handle_task.go : handle record %v", record)
		body := record.Body
		var aTask apimodel.AsyncTask
		err := json.Unmarshal([]byte(body), &aTask)
		if err != nil {
			anlogger.Errorf(lc, "internal_handle_task.go : error unmarshal body [%s] : %v", body, err)
			return errors.New(fmt.Sprintf("error unmarshal body %s : %v", body, err))
		}
		switch aTask.TaskType {
		case apimodel.RemovePhotoTaskType:
			ok, errStr := deletePhoto(body, lc)
			if !ok {
				return errors.New(errStr)
			}
		default:
			return errors.New(fmt.Sprintf("unsuported taks type %s", aTask.TaskType))
		}
	}

	anlogger.Debugf(lc, "internal_handle_task.go : successfully delete photos from request %v", event)
	return nil
}

//return ok and error string
func deletePhoto(body string, lc *lambdacontext.LambdaContext) (bool, string) {
	anlogger.Debugf(lc, "internal_handle_task.go : delete photo using body [%s]", body)
	var rTask apimodel.RemovePhotoAsyncTask
	err := json.Unmarshal([]byte(body), &rTask)
	anlogger.Debugf(lc, "internal_handle_task.go : result unmarshal %v", rTask)
	if err != nil {
		anlogger.Errorf(lc, "internal_handle_task.go : error unmarshal body [%s] : %v", body, err)
		return false, fmt.Sprintf("error unmarshal body [%s] : %v", body, err)
	}
	input := &dynamodb.DeleteItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			apimodel.UserIdColumnName: {
				S: aws.String(rTask.UserId),
			},
			apimodel.PhotoIdColumnName: {
				S: aws.String(rTask.PhotoId),
			},
		},
		TableName: aws.String(rTask.TableName),
	}
	_, err = awsDbClient.DeleteItem(input)
	if err != nil {
		anlogger.Errorf(lc, "internal_handle_task.go : error delete photo using task %v : %v", rTask, err)
		return false, fmt.Sprintf("error delete photo using task %v : %v", rTask, err)
	}
	anlogger.Debugf(lc, "internal_handle_task.go : successfully delete photo using task %v", rTask)
	return true, ""
}

func main() {
	basicLambda.Start(handler)
}

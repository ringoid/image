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
	"github.com/aws/aws-sdk-go/service/s3"
)

var anlogger *syslog.Logger
var awsDbClient *dynamodb.DynamoDB
var awsS3Client *s3.S3

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

	awsS3Client = s3.New(awsSession)
	anlogger.Debugf(nil, "internal_handle_task.go : s3 client was successfully initialized")
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
			anlogger.Errorf(lc, "internal_handle_task.go : error unmarshal body [%s] to AsyncTask : %v", body, err)
			return errors.New(fmt.Sprintf("error unmarshal body %s : %v", body, err))
		}
		switch aTask.TaskType {
		case apimodel.RemovePhotoTaskType:
			var rTask apimodel.RemovePhotoAsyncTask
			err := json.Unmarshal([]byte(body), &rTask)
			if err != nil {
				anlogger.Errorf(lc, "internal_handle_task.go : error unmarshal body [%s] to RemovePhotoTaskType: %v", body, err)
				return errors.New(fmt.Sprintf("error unmarshal body %s : %v", body, err))
			}
			userPhoto, ok, errStr := getUserPhoto(rTask.UserId, rTask.PhotoId, rTask.TableName, lc)
			if !ok {
				return errors.New(errStr)
			}

			ok, errStr = deleteFromS3(userPhoto.Bucket, userPhoto.Key, rTask.UserId, lc)
			if !ok {
				return errors.New(errStr)
			}

			ok, errStr = deletePhotoFromDynamo(rTask.UserId, rTask.PhotoId, rTask.TableName, lc)
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

//return userPhoto, ok and error string
func getUserPhoto(userId, photoId, tableName string, lc *lambdacontext.LambdaContext) (*apimodel.UserPhoto, bool, string) {
	anlogger.Debugf(lc, "internal_handle_task.go : get userPhoto for userId [%s], photoId [%s] from table [%s]",
		userId, photoId, tableName)
	input := &dynamodb.GetItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			apimodel.UserIdColumnName: {
				S: aws.String(userId),
			},
			apimodel.PhotoIdColumnName: {
				S: aws.String(photoId),
			},
		},
		ConsistentRead: aws.Bool(true),
		TableName:      aws.String(tableName),
	}
	result, err := awsDbClient.GetItem(input)
	if err != nil {
		anlogger.Errorf(lc, "internal_handle_task.go : error get item for userId [%s], photoId [%s] and table [%s] : %v",
			userId, photoId, tableName, err)
		return nil, false, apimodel.InternalServerError
	}
	if len(result.Item) == 0 {
		anlogger.Warnf(lc, "internal_handle_task.go : there is no item for userId [%s], photoId [%s] and table [%s]",
			userId, photoId, tableName)
		return nil, true, ""
	}

	res := apimodel.UserPhoto{
		Bucket: *result.Item[apimodel.PhotoBucketColumnName].S,
		Key:    *result.Item[apimodel.PhotoKeyColumnName].S,
	}
	anlogger.Debugf(lc, "internal_handle_task.go : successfully get userPhoto %v for userId [%s], photoId [%s] and table [%s]",
		res, userId, photoId, tableName)

	return &res, true, ""
}

//return ok and error string
func deleteFromS3(bucket, key, userId string, lc *lambdacontext.LambdaContext) (bool, string) {
	anlogger.Debugf(lc, "internal_handle_task.go : delete from s3 bucket [%s] with key [%s] for userId [%s]",
		bucket, key, userId)

	input := &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}

	_, err := awsS3Client.DeleteObject(input)
	if err != nil {
		anlogger.Errorf(lc, "internal_handle_task.go : error delete from s3 bucket [%s] with key [%s] for userId [%s] : %v",
			bucket, key, userId, err)
		return false, apimodel.InternalServerError
	}

	anlogger.Debugf(lc, "internal_handle_task.go : successfully delete from s3 bucket [%s] with key [%s] for userId [%s]",
		bucket, key, userId)
	return true, ""
}

//return ok and error string
func deletePhotoFromDynamo(userId, photoId, tableName string, lc *lambdacontext.LambdaContext) (bool, string) {
	anlogger.Debugf(lc, "internal_handle_task.go : delete photo using userId [%s] and photoId [%s] from tableName [%s]", userId, photoId, tableName)
	input := &dynamodb.DeleteItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			apimodel.UserIdColumnName: {
				S: aws.String(userId),
			},
			apimodel.PhotoIdColumnName: {
				S: aws.String(photoId),
			},
		},
		TableName: aws.String(tableName),
	}
	_, err := awsDbClient.DeleteItem(input)
	if err != nil {
		anlogger.Errorf(lc, "internal_handle_task.go : error delete photo using userId [%s] and photoId [%s] from tableName [%s] : %v",
			userId, photoId, tableName, err)
		return false, fmt.Sprintf("error delete photo using userId [%s] and photoId [%s] from tableName [%s] : %v",
			userId, photoId, tableName, err)
	}
	anlogger.Debugf(lc, "internal_handle_task.go : successfully delete photo userId [%s] and photoId [%s] from tableName [%s]",
		userId, photoId, tableName)
	return true, ""
}

func main() {
	basicLambda.Start(handler)
}

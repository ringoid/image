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
		fmt.Printf("lambda-initialization : clean.go : env can not be empty ENV\n")
		os.Exit(1)
	}
	fmt.Printf("lambda-initialization : clean.go : start with ENV = [%s]\n", env)

	//!!!START : VERY IMPORTANT CODE. NEVER DELETE OR MODIFY!!!
	if env != "test" {
		panic("use clean DB in not safe env")
	}
	//!!!END : VERY IMPORTANT CODE. NEVER DELETE OR MODIFY!!!

	papertrailAddress, ok = os.LookupEnv("PAPERTRAIL_LOG_ADDRESS")
	if !ok {
		fmt.Printf("lambda-initialization : clean.go : env can not be empty PAPERTRAIL_LOG_ADDRESS\n")
		os.Exit(1)
	}
	fmt.Printf("lambda-initialization : clean.go : start with PAPERTRAIL_LOG_ADDRESS = [%s]\n", papertrailAddress)

	anlogger, err = syslog.New(papertrailAddress, fmt.Sprintf("%s-%s", env, "internal-clean-db-image"))
	if err != nil {
		fmt.Errorf("lambda-initialization : clean.go : error during startup : %v\n", err)
		os.Exit(1)
	}
	anlogger.Debugf(nil, "lambda-initialization : clean.go : logger was successfully initialized")

	userPhotoTable, ok = os.LookupEnv("USER_PHOTO_TABLE")
	if !ok {
		fmt.Printf("lambda-initialization : clean.go : env can not be empty USER_PHOTO_TABLE")
		os.Exit(1)
	}
	anlogger.Debugf(nil, "lambda-initialization : clean.go : start with USER_PHOTO_TABLE = [%s]", userPhotoTable)

	awsSession, err = session.NewSession(aws.NewConfig().
		WithRegion(apimodel.Region).WithMaxRetries(apimodel.MaxRetries).
		WithLogger(aws.LoggerFunc(func(args ...interface{}) { anlogger.AwsLog(args) })).WithLogLevel(aws.LogOff))
	if err != nil {
		anlogger.Fatalf(nil, "lambda-initialization : clean.go : error during initialization : %v", err)
	}
	anlogger.Debugf(nil, "lambda-initialization : clean.go : aws session was successfully initialized")

	awsDbClient = dynamodb.New(awsSession)
	anlogger.Debugf(nil, "lambda-initialization : clean.go : dynamodb client was successfully initialized")
}

func handler(ctx context.Context) error {
	lc, _ := lambdacontext.FromContext(ctx)
	err := eraseTable(userPhotoTable, apimodel.UserIdColumnName, apimodel.PhotoIdColumnName, lc)
	if err != nil {
		return err
	}
	return nil
}

func eraseTable(tableName, partitionKeyColumnName, sortKeyColumnName string, lc *lambdacontext.LambdaContext) error {
	anlogger.Debugf(lc, "clean.go : start clean [%s] table", tableName)
	var lastEvaluatedKey map[string]*dynamodb.AttributeValue
	for {
		scanInput := &dynamodb.ScanInput{
			ConsistentRead:    aws.Bool(true),
			TableName:         aws.String(tableName),
			ExclusiveStartKey: lastEvaluatedKey,
		}
		scanResult, err := awsDbClient.Scan(scanInput)
		if err != nil {
			return errors.New(fmt.Sprintf("error during scan %s table", tableName))
		}
		items := scanResult.Items
		for _, item := range items {
			partitionKey := item[partitionKeyColumnName].S
			sortKey := item[sortKeyColumnName].S
			deleteInput := &dynamodb.DeleteItemInput{
				Key: map[string]*dynamodb.AttributeValue{
					partitionKeyColumnName: {
						S: partitionKey,
					},
					sortKeyColumnName: {
						S: sortKey,
					},
				},
				TableName: aws.String(tableName),
			}
			_, err = awsDbClient.DeleteItem(deleteInput)
			if err != nil {
				return errors.New(fmt.Sprintf("error during delete from %s table", tableName))
			}
		}
		lastEvaluatedKey = scanResult.LastEvaluatedKey
		if len(lastEvaluatedKey) == 0 {
			break
		}
	}
	anlogger.Debugf(lc, "clean.go : successfully clean [%s] table", tableName)
	return nil
}

func main() {
	basicLambda.Start(handler)
}

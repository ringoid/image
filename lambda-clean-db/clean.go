package main

import (
	"context"
	basicLambda "github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/aws"
	"os"
	"fmt"
	"github.com/aws/aws-lambda-go/lambdacontext"
	"errors"
	"github.com/ringoid/commons"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	"github.com/aws/aws-dax-go/dax"
)

var anlogger *commons.Logger
var awsDaxClient dynamodbiface.DynamoDBAPI
var userPhotoTable string

func init() {
	var env string
	var ok bool
	var papertrailAddress string
	var err error

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

	anlogger, err = commons.New(papertrailAddress, fmt.Sprintf("%s-%s", env, "internal-clean-db-image"))
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

	daxEndpoint, ok := os.LookupEnv("DAX_ENDPOINT")
	if !ok {
		anlogger.Fatalf(nil, "lambda-initialization : clean.go : env can not be empty DAX_ENDPOINT")
	}
	cfg := dax.DefaultConfig()
	cfg.HostPorts = []string{daxEndpoint}
	cfg.Region = commons.Region
	awsDaxClient, err = dax.New(cfg)
	if err != nil {
		anlogger.Fatalf(nil, "lambda-initialization : clean.go : error initialize DAX cluster")
	}
	anlogger.Debugf(nil, "lambda-initialization : clean.go : dax client was successfully initialized")
}

func handler(ctx context.Context) error {
	lc, _ := lambdacontext.FromContext(ctx)
	err := eraseTable(userPhotoTable, commons.UserIdColumnName, commons.PhotoIdColumnName, lc)
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
		anlogger.Debugf(lc, "clean.go : start scan a batch")
		scanResult, err := awsDaxClient.Scan(scanInput)
		if err != nil {
			anlogger.Errorf(lc, "clean.go : error during scan %s table : %v", tableName, err)
			return errors.New(fmt.Sprintf("error during scan %s table", tableName))
		}
		anlogger.Debugf(lc, "clean.go : finish scan a batch")
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
			anlogger.Debugf(lc, "clean.go : start delete using dax")
			_, err = awsDaxClient.DeleteItem(deleteInput)
			if err != nil {
				anlogger.Errorf(lc, "clean.go : error during delete from %s table using dax : %v", tableName, err)
				return errors.New(fmt.Sprintf("error during delete from %s table using dax", tableName))
			}
			anlogger.Debugf(lc, "clean.go : finish delete using dax")
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

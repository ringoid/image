package main

import (
	"../apimodel"
	"github.com/aws/aws-lambda-go/lambdacontext"
	"fmt"
	"encoding/json"
	"errors"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/ringoid/commons"
	"github.com/aws/aws-sdk-go/service/sqs"
)

func deleteAllPhotos(body []byte, userPhotoTable, asyncTaskQueue string, awsSqsClient *sqs.SQS,
	awsDbClient *dynamodb.DynamoDB, lc *lambdacontext.LambdaContext, anlogger *commons.Logger) error {

	anlogger.Debugf(lc, "delete_user.go : handle event and delete all photos, body %s", string(body))
	var aEvent commons.UserCallDeleteHimselfEvent
	err := json.Unmarshal([]byte(body), &aEvent)
	if err != nil {
		anlogger.Errorf(lc, "delete_user.go : error unmarshal body [%s] to UserCallDeleteHimselfEvent: %v", string(body), err)
		return errors.New(fmt.Sprintf("error unmarshal body %s : %v", string(body), err))
	}
	userPhotoMetaPartitionKey := aEvent.UserId + commons.PhotoPrimaryKeyMetaPostfix

	//first delete all photos
	photoIds, err := allUserPhotoId(aEvent.UserId, userPhotoTable, awsDbClient, lc, anlogger)
	if err != nil {
		return err
	}

	for _, eachPhotoIdP := range photoIds {
		val := *eachPhotoIdP.S
		ok, errStr := apimodel.MarkPhotoAsDel(aEvent.UserId, val, userPhotoTable, awsDbClient, anlogger, lc)
		if !ok {
			return errors.New(errStr)
		}
		if aEvent.UserReportStatus == commons.UserTakePartInReport && commons.IsItOriginPhoto(val) {
			anlogger.Infof(lc, "delete_user.go : user with userId [%s] was reported or was report initiator, so skip delete his origin photo [%s] from s3",
				aEvent.UserId, val)
			continue
		}

		task := apimodel.NewRemovePhotoAsyncTask(aEvent.UserId, val, userPhotoTable)
		ok, errStr = commons.SendAsyncTask(task, asyncTaskQueue, aEvent.UserId, 0, awsSqsClient, anlogger, lc)
		if !ok {
			return errors.New(errStr)
		}
	}

	//now lets delete all meta inf
	metaIds, err := allUserPhotoId(userPhotoMetaPartitionKey, userPhotoTable, awsDbClient, lc, anlogger)
	if err != nil {
		return err
	}

	for _, eachPhotoIdP := range metaIds {
		val := *eachPhotoIdP.S
		ok, errStr := apimodel.MarkPhotoAsDel(userPhotoMetaPartitionKey, val, userPhotoTable, awsDbClient, anlogger, lc)
		if !ok {
			return errors.New(errStr)
		}
	}
	anlogger.Debugf(lc, "delete_user.go : successfully handle event and delete all photos, body %s", string(body))
	return nil
}

func allUserPhotoId(partitionKey, userPhotoTable string, awsDbClient *dynamodb.DynamoDB, lc *lambdacontext.LambdaContext, anlogger *commons.Logger) ([]*dynamodb.AttributeValue, error) {
	var lastEvaluatedKey map[string]*dynamodb.AttributeValue

	var finalResult []*dynamodb.AttributeValue

	for {
		input := &dynamodb.QueryInput{
			ExpressionAttributeNames: map[string]*string{
				"#userId": aws.String(commons.UserIdColumnName),
			},
			ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
				":userIdV": {
					S: aws.String(partitionKey),
				},
			},
			ExclusiveStartKey:      lastEvaluatedKey,
			ConsistentRead:         aws.Bool(true),
			KeyConditionExpression: aws.String("#userId = :userIdV"),
			FilterExpression:       aws.String(fmt.Sprintf("attribute_not_exists(%s)", commons.PhotoDeletedAtColumnName)),
			TableName:              aws.String(userPhotoTable),
		}

		result, err := awsDbClient.Query(input)
		if err != nil {
			anlogger.Errorf(lc, "delete_user.go : error while query all photos for partitionKey [%s] : %v", partitionKey, err)
			return finalResult, errors.New(fmt.Sprintf("error query all photos for partition key %s : %v", partitionKey, err))
		}

		lastEvaluatedKey = result.LastEvaluatedKey

		for _, item := range result.Items {
			finalResult = append(finalResult, item[commons.PhotoIdColumnName])
		}

		if len(lastEvaluatedKey) == 0 {
			anlogger.Debugf(lc, "delete_user.go : all photo ids size is [%d] for partition key [%s]", len(finalResult), partitionKey)
			return finalResult, nil
		}
	}
}

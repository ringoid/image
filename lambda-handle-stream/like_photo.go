package main

import (
	"github.com/aws/aws-lambda-go/lambdacontext"
	"fmt"
	"encoding/json"
	"errors"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/ringoid/commons"
)

func likePhoto(body []byte, userPhotoTable string, awsDbClient *dynamodb.DynamoDB, lc *lambdacontext.LambdaContext, anlogger *commons.Logger) error {
	anlogger.Debugf(lc, "like_photo.go : handle event and like photo, body %s", string(body))
	var aEvent commons.PhotoLikeInternalEvent
	err := json.Unmarshal([]byte(body), &aEvent)
	if err != nil {
		anlogger.Errorf(lc, "like_photo.go : error unmarshal body [%s] to ImageRemovePhotoTaskType: %v", string(body), err)
		return errors.New(fmt.Sprintf("error unmarshal body %s : %v", string(body), err))
	}
	userPhotoMetaPartitionKey := aEvent.UserId + commons.PhotoPrimaryKeyMetaPostfix

	input :=
		&dynamodb.UpdateItemInput{
			ExpressionAttributeNames: map[string]*string{
				"#like": aws.String(commons.PhotoLikesColumnName),
			},
			ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
				":likeV": {
					N: aws.String("1"),
				},
			},
			Key: map[string]*dynamodb.AttributeValue{
				commons.UserIdColumnName: {
					S: aws.String(userPhotoMetaPartitionKey),
				},
				commons.PhotoIdColumnName: {
					S: aws.String(aEvent.OriginalPhotoId),
				},
			},
			TableName:        aws.String(userPhotoTable),
			UpdateExpression: aws.String("ADD #like :likeV"),
		}

	_, err = awsDbClient.UpdateItem(input)
	if err != nil {
		anlogger.Errorf(lc, "like_photo.go : error while update likes on photo with meta partition key [%s], original photoId [%s] for userId [%s] : %v",
			userPhotoMetaPartitionKey, aEvent.OriginalPhotoId, aEvent.UserId, err)
		return errors.New(fmt.Sprintf("error like photo %s : %v", string(body), err))
	}
	anlogger.Debugf(lc, "like_photo.go : successfully handle event and like photo, body %v", string(body))
	return nil
}

package main

import (
	"github.com/aws/aws-lambda-go/lambdacontext"
	"fmt"
	"encoding/json"
	"errors"
	"github.com/ringoid/commons"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	"../apimodel"
)

func likePhotoUpdate(body []byte, userPhotoTable string, dbClient dynamodbiface.DynamoDBAPI, lc *lambdacontext.LambdaContext, anlogger *commons.Logger) error {
	anlogger.Debugf(lc, "like_photo.go : handle event and like photo, body %s", string(body))
	var aEvent commons.PhotoLikeInternalEvent
	err := json.Unmarshal([]byte(body), &aEvent)
	if err != nil {
		anlogger.Errorf(lc, "like_photo.go : error unmarshal body [%s] to ImageRemovePhotoTaskType: %v", string(body), err)
		return errors.New(fmt.Sprintf("error unmarshal body %s : %v", string(body), err))
	}
	userPhotoMetaPartitionKey := aEvent.UserId + commons.PhotoPrimaryKeyMetaPostfix

	return apimodel.LikePhotoUpdate(aEvent.UserId, aEvent.OriginalPhotoId, userPhotoMetaPartitionKey, userPhotoTable, dbClient, anlogger, lc)
}

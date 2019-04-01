package main

import (
	"../apimodel"
	"github.com/aws/aws-lambda-go/lambdacontext"
	"fmt"
	"encoding/json"
	"errors"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/ringoid/commons"
	"github.com/aws/aws-sdk-go/service/sqs"
)

func hidePhoto(body []byte, userPhotoTable, asyncTaskQueue string, awsSqsClient *sqs.SQS,
	awsDbClient *dynamodb.DynamoDB, lc *lambdacontext.LambdaContext, anlogger *commons.Logger) error {

	anlogger.Debugf(lc, "hide_photo.go : handle event and hide photo, body %s", string(body))
	var aEvent commons.HidePhotoInternalEvent
	err := json.Unmarshal([]byte(body), &aEvent)
	if err != nil {
		anlogger.Errorf(lc, "hide_photo.go : error unmarshal body [%s] to HidePhotoInternalEvent: %v", string(body), err)
		return errors.New(fmt.Sprintf("error unmarshal body %s : %v", string(body), err))
	}

	photoIds, originPhotoId := apimodel.GetAllPhotoIdsBasedOnSource(aEvent.OriginalPhotoId, aEvent.UserId, anlogger, lc)
	for _, val := range photoIds {
		ok, errStr := apimodel.MarkPhotoAsHiddenInModerationProcess(aEvent.UserId, val, userPhotoTable, awsDbClient, anlogger, lc)
		if !ok {
			anlogger.Errorf(lc, "hide_photo.go : userId [%s], return %s to client", aEvent.UserId, errStr)
			return errors.New(fmt.Sprintf("error mark photo as deleted, body %s : %v", string(body), err))
		}

		//here we don't check does user take park in report coz hide photo can call only moderator
		if val == originPhotoId {
			continue
		}

		task := apimodel.NewRemovePhotoAsyncTask(aEvent.UserId, val, userPhotoTable)
		ok, errStr = commons.SendAsyncTask(task, asyncTaskQueue, aEvent.UserId, 0, awsSqsClient, anlogger, lc)
		if !ok {
			anlogger.Errorf(lc, "hide_photo.go : userId [%s], return %s to client", aEvent.UserId, errStr)
			return errors.New(fmt.Sprintf("error mark photo as deleted (durgin hide photo), body %s : %v", string(body), err))
		}
	}

	//Mark photo meta info like deleted also
	ok, errStr := apimodel.MarkPhotoAsHiddenInModerationProcess(aEvent.UserId+commons.PhotoPrimaryKeyMetaPostfix, originPhotoId, userPhotoTable, awsDbClient, anlogger, lc)
	if !ok {
		anlogger.Errorf(lc, "hide_photo.go : userId [%s], return %s to client", aEvent.UserId, errStr)
		return errors.New(fmt.Sprintf("error mark meta info about photo as deleted (durgin hide photo), body %s : %v", string(body), err))
	}

	anlogger.Debugf(lc, "hide_photo.go : successfully handle event and hide photo, body %s", string(body))
	return nil
}

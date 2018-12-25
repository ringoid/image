package main

import (
	"../apimodel"
	"github.com/aws/aws-lambda-go/lambdacontext"
	"fmt"
	"encoding/json"
	"errors"
	"github.com/ringoid/commons"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
)

func deleteAllPhotos(body []byte, userPhotoTable, asyncTaskQueue string, awsSqsClient *sqs.SQS,
	daxClient dynamodbiface.DynamoDBAPI, lc *lambdacontext.LambdaContext, anlogger *commons.Logger) error {

	anlogger.Debugf(lc, "delete_user.go : handle event and delete all photos, body %s", string(body))
	var aEvent commons.UserCallDeleteHimselfEvent
	err := json.Unmarshal([]byte(body), &aEvent)
	if err != nil {
		anlogger.Errorf(lc, "delete_user.go : error unmarshal body [%s] to UserCallDeleteHimselfEvent: %v", string(body), err)
		return errors.New(fmt.Sprintf("error unmarshal body %s : %v", string(body), err))
	}
	userPhotoMetaPartitionKey := aEvent.UserId + commons.PhotoPrimaryKeyMetaPostfix

	//first delete all photos
	photoIds, err := apimodel.AllUserPhotoIdQuery(aEvent.UserId, userPhotoTable, daxClient, anlogger, lc)
	if err != nil {
		return err
	}

	for _, eachPhotoIdP := range photoIds {
		val := *eachPhotoIdP.S
		ok, errStr := apimodel.MarkPhotoAsDelUpdate(aEvent.UserId, val, userPhotoTable, daxClient, anlogger, lc)
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
	metaIds, err := apimodel.AllUserPhotoIdQuery(userPhotoMetaPartitionKey, userPhotoTable, daxClient, anlogger, lc)
	if err != nil {
		return err
	}

	for _, eachPhotoIdP := range metaIds {
		val := *eachPhotoIdP.S
		ok, errStr := apimodel.MarkPhotoAsDelUpdate(userPhotoMetaPartitionKey, val, userPhotoTable, daxClient, anlogger, lc)
		if !ok {
			return errors.New(errStr)
		}
	}
	anlogger.Debugf(lc, "delete_user.go : successfully handle event and delete all photos, body %s", string(body))
	return nil
}

package main

import (
	"fmt"
	"github.com/aws/aws-lambda-go/lambdacontext"
	"../apimodel"
	"encoding/json"
	"errors"
	"github.com/ringoid/commons"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
)

func removePhoto(body []byte, daxClient dynamodbiface.DynamoDBAPI, lc *lambdacontext.LambdaContext, anlogger *commons.Logger) error {
	var rTask apimodel.RemovePhotoAsyncTask
	err := json.Unmarshal([]byte(body), &rTask)
	if err != nil {
		anlogger.Errorf(lc, "remove_photo.go : error unmarshal body [%s] to ImageRemovePhotoTaskType: %v", body, err)
		return errors.New(fmt.Sprintf("error unmarshal body %s : %v", body, err))
	}
	userPhoto, ok, errStr := apimodel.GetUserPhoto(rTask.UserId, rTask.PhotoId, rTask.TableName, daxClient, anlogger, lc)
	if !ok {
		return errors.New(errStr)
	}

	ok, errStr = commons.DeleteFromS3(userPhoto.Bucket, userPhoto.Key, rTask.UserId, awsS3Client, lc, anlogger)
	if !ok {
		return errors.New(errStr)
	}

	//There is no need to delete photo from DB, mark is enough
	//so
	//ok, errStr = deletePhotoFromDynamo(rTask.UserId, rTask.PhotoId, rTask.TableName, lc, anlogger)
	//if !ok {
	//	return errors.New(errStr)
	//}
	//
	////we need to delete meta info also
	//if strings.HasPrefix(rTask.PhotoId, "origin_") {
	//	ok, errStr = deletePhotoFromDynamo(rTask.UserId+apimodel.PhotoPrimaryKeyMetaPostfix, rTask.PhotoId, rTask.TableName, lc, anlogger)
	//	if !ok {
	//		return errors.New(errStr)
	//	}
	//}

	return nil
}

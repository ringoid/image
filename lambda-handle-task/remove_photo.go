package main

import (
	"fmt"
	"github.com/aws/aws-lambda-go/lambdacontext"
	"../apimodel"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/aws"
	"encoding/json"
	"errors"
	"github.com/ringoid/commons"
)

func removePhoto(body []byte, lc *lambdacontext.LambdaContext, anlogger *commons.Logger) error {
	var rTask apimodel.RemovePhotoAsyncTask
	err := json.Unmarshal([]byte(body), &rTask)
	if err != nil {
		anlogger.Errorf(lc, "remove_photo.go : error unmarshal body [%s] to ImageRemovePhotoTaskType: %v", body, err)
		return errors.New(fmt.Sprintf("error unmarshal body %s : %v", body, err))
	}
	userPhoto, ok, errStr := getUserPhoto(rTask.UserId, rTask.PhotoId, rTask.TableName, lc, anlogger)
	if !ok {
		return errors.New(errStr)
	}

	if userPhoto == nil {
		return nil
	}

	ok, errStr = commons.DeleteFromS3(userPhoto.Bucket, userPhoto.Key, rTask.UserId, awsS3Client, lc, anlogger)
	if !ok {
		return errors.New(errStr)
	}
	anlogger.Infof(lc, "remove_photo.go : successfully remove photo from bucket [%s] with key [%s] for userId [%s]", userPhoto.Bucket, userPhoto.Key, rTask.UserId)
	return nil
}

//return userPhoto, ok and error string
func getUserPhoto(userId, photoId, tableName string, lc *lambdacontext.LambdaContext, anlogger *commons.Logger) (*apimodel.UserPhoto, bool, string) {
	anlogger.Debugf(lc, "remove_photo.go : get userPhoto for userId [%s], photoId [%s] from table [%s]",
		userId, photoId, tableName)
	input := &dynamodb.GetItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			commons.UserIdColumnName: {
				S: aws.String(userId),
			},
			commons.PhotoIdColumnName: {
				S: aws.String(photoId),
			},
		},
		ConsistentRead: aws.Bool(true),
		TableName:      aws.String(tableName),
	}
	result, err := awsDbClient.GetItem(input)
	if err != nil {
		anlogger.Errorf(lc, "remove_photo.go : error get item for userId [%s], photoId [%s] and table [%s] : %v",
			userId, photoId, tableName, err)
		return nil, false, commons.InternalServerError
	}
	if len(result.Item) == 0 {
		anlogger.Warnf(lc, "remove_photo.go : there is no item for userId [%s], photoId [%s] and table [%s]",
			userId, photoId, tableName)
		return nil, true, ""
	}

	_, photoDeleted := result.Item[commons.PhotoDeletedAtColumnName]
	backetAttr, bucketExist := result.Item[commons.PhotoBucketColumnName]
	keyAttr, keyExist := result.Item[commons.PhotoKeyColumnName]

	if bucketExist && backetAttr != nil && keyExist && keyAttr != nil {
		res := apimodel.UserPhoto{
			Bucket: *result.Item[commons.PhotoBucketColumnName].S,
			Key:    *result.Item[commons.PhotoKeyColumnName].S,
		}
		anlogger.Debugf(lc, "remove_photo.go : successfully get userPhoto %v for userId [%s], photoId [%s] and table [%s]",
			res, userId, photoId, tableName)

		return &res, true, ""
	}

	if photoDeleted {
		//it just means that photo was deleted before was resized
		return nil, true, ""
	}

	//and this means an error in our soft
	anlogger.Errorf(lc, "remove_photo.go : try to remove photo with photoId [%s] without bucket or key in DB, userId [%s]",
		photoId, userId)
	return nil, false, commons.InternalServerError
}

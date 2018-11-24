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

	//todo: we need to check if your was reported and don't delete origin photo

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

	res := apimodel.UserPhoto{
		Bucket: *result.Item[commons.PhotoBucketColumnName].S,
		Key:    *result.Item[commons.PhotoKeyColumnName].S,
	}
	anlogger.Debugf(lc, "remove_photo.go : successfully get userPhoto %v for userId [%s], photoId [%s] and table [%s]",
		res, userId, photoId, tableName)

	return &res, true, ""
}

//return ok and error string
func deletePhotoFromDynamo(userId, photoId, tableName string, lc *lambdacontext.LambdaContext, anlogger *commons.Logger) (bool, string) {
	anlogger.Debugf(lc, "remove_photo.go : delete photo using userId [%s] and photoId [%s] from tableName [%s]", userId, photoId, tableName)
	input := &dynamodb.DeleteItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			commons.UserIdColumnName: {
				S: aws.String(userId),
			},
			commons.PhotoIdColumnName: {
				S: aws.String(photoId),
			},
		},
		TableName: aws.String(tableName),
	}
	_, err := awsDbClient.DeleteItem(input)
	if err != nil {
		anlogger.Errorf(lc, "remove_photo.go : error delete photo using userId [%s] and photoId [%s] from tableName [%s] : %v",
			userId, photoId, tableName, err)
		return false, fmt.Sprintf("error delete photo using userId [%s] and photoId [%s] from tableName [%s] : %v",
			userId, photoId, tableName, err)
	}
	anlogger.Debugf(lc, "remove_photo.go : successfully delete photo userId [%s] and photoId [%s] from tableName [%s]",
		userId, photoId, tableName)
	return true, ""
}

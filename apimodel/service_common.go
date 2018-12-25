package apimodel

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-lambda-go/lambdacontext"
	"time"
	"strconv"
	"fmt"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/ringoid/commons"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	"strings"
	"errors"
)

//make Update
//return ok and error string
func SavePhotoUpdate(userPhoto *UserPhoto, userPhotoTable string, dbClient dynamodbiface.DynamoDBAPI, anlogger *commons.Logger, lc *lambdacontext.LambdaContext) (bool, string) {
	anlogger.Debugf(lc, "common_action.go : save photo %v for userId [%s]", userPhoto, userPhoto.UserId)
	input :=
		&dynamodb.UpdateItemInput{
			ExpressionAttributeNames: map[string]*string{
				"#photoType":   aws.String(commons.PhotoTypeColumnName),
				"#photoBucket": aws.String(commons.PhotoBucketColumnName),
				"#photoKey":    aws.String(commons.PhotoKeyColumnName),
				"#photoSize":   aws.String(commons.PhotoSizeColumnName),
				"#updatedAt":   aws.String(commons.UpdatedTimeColumnName),
			},
			ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
				":photoTypeV": {
					S: aws.String(userPhoto.PhotoType),
				},
				":photoBucketV": {
					S: aws.String(userPhoto.Bucket),
				},
				":photoKeyV": {
					S: aws.String(userPhoto.Key),
				},
				":photoSizeV": {
					N: aws.String(strconv.FormatInt(userPhoto.Size, 10)),
				},
				":updatedAtV": {
					S: aws.String(time.Now().UTC().Format("2006-01-02-15-04-05.000")),
				},
			},
			Key: map[string]*dynamodb.AttributeValue{
				commons.UserIdColumnName: {
					S: aws.String(userPhoto.UserId),
				},
				commons.PhotoIdColumnName: {
					S: aws.String(userPhoto.PhotoId),
				},
			},
			TableName:           aws.String(userPhotoTable),
			ConditionExpression: aws.String(fmt.Sprintf("attribute_not_exists(%s)", commons.PhotoDeletedAtColumnName)),
			UpdateExpression:    aws.String("SET #photoType = :photoTypeV, #photoBucket = :photoBucketV, #photoKey = :photoKeyV, #photoSize = :photoSizeV, #updatedAt = :updatedAtV"),
		}

	if userPhoto.PhotoSourceUri != "" {
		input.ExpressionAttributeNames["#photoUri"] = aws.String(commons.PhotoSourceUriColumnName)
		input.ExpressionAttributeValues[":photoUriV"] = &dynamodb.AttributeValue{
			S: aws.String(userPhoto.PhotoSourceUri),
		}
		input.UpdateExpression = aws.String("SET #photoUri = :photoUriV, #photoType = :photoTypeV, #photoBucket = :photoBucketV, #photoKey = :photoKeyV, #photoSize = :photoSizeV, #updatedAt = :updatedAtV")
	}

	_, err := dbClient.UpdateItem(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case dynamodb.ErrCodeConditionalCheckFailedException:
				anlogger.Debugf(lc, "common_action.go : photo [%v] was already deleted for userId [%s]", userPhoto, userPhoto.UserId)
				return false, ""
			}
		}
		anlogger.Errorf(lc, "common_action.go : error while save photo %v for userId [%s] : %v", userPhoto, userPhoto.UserId, err)
		return false, commons.InternalServerError
	}

	anlogger.Debugf(lc, "common_action.go : successfully save photo %v for userId [%s]", userPhoto, userPhoto.UserId)
	return true, ""
}

//make Update
//return ok and error string
func MarkPhotoAsDelUpdate(userId, photoId, tableName string, dbClient dynamodbiface.DynamoDBAPI, anlogger *commons.Logger, lc *lambdacontext.LambdaContext) (bool, string) {
	anlogger.Debugf(lc, "common_action.go : mark photoId [%s] as deleted for userId [%s]", photoId, userId)
	input :=
		&dynamodb.UpdateItemInput{
			ExpressionAttributeNames: map[string]*string{
				"#deletedAt": aws.String(commons.PhotoDeletedAtColumnName),
			},
			ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
				":deletedAtV": {
					S: aws.String(time.Now().UTC().Format("2006-01-02-15-04-05.000")),
				},
			},
			Key: map[string]*dynamodb.AttributeValue{
				commons.UserIdColumnName: {
					S: aws.String(userId),
				},
				commons.PhotoIdColumnName: {
					S: aws.String(photoId),
				},
			},
			TableName:        aws.String(tableName),
			UpdateExpression: aws.String("SET #deletedAt = :deletedAtV"),
		}
	_, err := dbClient.UpdateItem(input)
	if err != nil {
		anlogger.Errorf(lc, "common_action.go : error while mark photo as deleted photoId [%s] for userId [%s] : %v", photoId, userId, err)
		return false, commons.InternalServerError
	}
	anlogger.Debugf(lc, "common_action.go : successfully mark photoId [%s] as deleted for userId [%s]", photoId, userId)
	return true, ""
}

//make Query
//return own photos, ok and error string
//todo:keep in mind that we should use ExclusiveStartKey later, if somebody will have > 100K photos
func GetOwnPhotosQuery(userId, resolution, userPhotoTable string, dbClient dynamodbiface.DynamoDBAPI, anlogger *commons.Logger, lc *lambdacontext.LambdaContext) ([]*UserPhoto, bool, string) {
	anlogger.Debugf(lc, "common_action.go : get all own photos for userId [%s] with resolution [%s]", userId, resolution)
	input := &dynamodb.QueryInput{
		ExpressionAttributeNames: map[string]*string{
			"#userId":  aws.String(commons.UserIdColumnName),
			"#photoId": aws.String(commons.PhotoIdColumnName),
		},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":userIdV": {
				S: aws.String(userId),
			},
			":photoIdV": {
				S: aws.String(resolution),
			},
		},
		FilterExpression: aws.String(fmt.Sprintf("attribute_not_exists(%s)", commons.PhotoDeletedAtColumnName)),
		//for using with DAX need to set false
		ConsistentRead:         aws.Bool(true),
		KeyConditionExpression: aws.String("#userId = :userIdV AND begins_with(#photoId, :photoIdV)"),
		TableName:              aws.String(userPhotoTable),
	}
	result, err := dbClient.Query(input)
	if err != nil {
		anlogger.Errorf(lc, "common_action.go : error while query all own photos userId [%s] with resolution [%s] : %v", userId, resolution, err)
		return make([]*UserPhoto, 0), false, commons.InternalServerError
	}

	if *result.Count == 0 {
		anlogger.Debugf(lc, "common_action.go : there is no photo for userId [%s] with resolution [%s]", userId, resolution)
		return make([]*UserPhoto, 0), true, ""
	}

	items := make([]*UserPhoto, 0)
	for _, v := range result.Items {
		originPhotoId := strings.Replace(*v[commons.PhotoIdColumnName].S, resolution, "origin", 1)
		items = append(items, &UserPhoto{
			UserId:         *v[commons.UserIdColumnName].S,
			PhotoId:        *v[commons.PhotoIdColumnName].S,
			PhotoSourceUri: *v[commons.PhotoSourceUriColumnName].S,
			PhotoType:      *v[commons.PhotoTypeColumnName].S,
			Bucket:         *v[commons.PhotoBucketColumnName].S,
			Key:            *v[commons.PhotoKeyColumnName].S,
			UpdatedAt:      *v[commons.UpdatedTimeColumnName].S,
			OriginPhotoId:  originPhotoId,
		})
	}
	anlogger.Debugf(lc, "common_action.go : successfully fetch [%v] photos for userId [%s] and resolution [%s], result=%v",
		*result.Count, userId, resolution, items)
	return items, true, ""
}

//make Query
//return photo's meta infs, ok and error string
//todo:keep in mind that we should use ExclusiveStartKey later, if somebody will have > 100K photos
func GetMetaInfsQuery(userId, userPhotoTable string, dbClient dynamodbiface.DynamoDBAPI, anlogger *commons.Logger, lc *lambdacontext.LambdaContext) (map[string]*UserPhotoMetaInf, bool, string) {
	anlogger.Debugf(lc, "common_action.go : get all photo's meta infs for userId [%s]", userId)
	metaInfPartitionKey := userId + commons.PhotoPrimaryKeyMetaPostfix
	input := &dynamodb.QueryInput{
		ExpressionAttributeNames: map[string]*string{
			"#userId": aws.String(commons.UserIdColumnName),
		},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":userIdV": {
				S: aws.String(metaInfPartitionKey),
			},
		},
		FilterExpression: aws.String(fmt.Sprintf("attribute_not_exists(%s)", commons.PhotoDeletedAtColumnName)),
		//for using with DAX need to set false
		ConsistentRead:         aws.Bool(true),
		KeyConditionExpression: aws.String("#userId = :userIdV"),
		TableName:              aws.String(userPhotoTable),
	}
	result, err := dbClient.Query(input)
	if err != nil {
		anlogger.Errorf(lc, "common_action.go : error while query all photo's meta infs for userId [%s] : %v", userId, err)
		return make(map[string]*UserPhotoMetaInf, 0), false, commons.InternalServerError
	}

	if *result.Count == 0 {
		anlogger.Debugf(lc, "common_action.go : there is no photo's meta info for userId [%s]", userId)
		return make(map[string]*UserPhotoMetaInf, 0), true, ""
	}

	anlogger.Debugf(lc, "common_action.go : there is [%d] photo's meta info for userId [%s]", *result.Count, userId)

	items := make(map[string]*UserPhotoMetaInf, 0)
	for _, v := range result.Items {
		photoId := *v[commons.PhotoIdColumnName].S

		likes, err := strconv.Atoi(*v[commons.PhotoLikesColumnName].N)
		if err != nil {
			anlogger.Errorf(lc, "common_action.go : error while convert likes from photo meta inf to int, photoId [%s] for userId [%s] : %v", photoId, userId, err)
			return make(map[string]*UserPhotoMetaInf, 0), false, commons.InternalServerError
		}

		items[photoId] = &UserPhotoMetaInf{
			UserId:        *v[commons.UserIdColumnName].S,
			OriginPhotoId: photoId,
			Likes:         likes,
		}
	}

	anlogger.Debugf(lc, "common_action.go : successfully fetch [%v] photo meta inf for userId [%s], result=%v",
		*result.Count, userId, items)

	return items, true, ""
}

//make GetItem
//return userPhoto, ok and error string
func GetUserPhoto(userId, photoId, tableName string, dbClient dynamodbiface.DynamoDBAPI, anlogger *commons.Logger, lc *lambdacontext.LambdaContext) (*UserPhoto, bool, string) {
	anlogger.Debugf(lc, "common_action.go : get userPhoto for userId [%s], photoId [%s] from table [%s]",
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
		//if use consistent read then we will read from dynamodb directly
		//ConsistentRead: aws.Bool(true),
		TableName: aws.String(tableName),
	}
	result, err := dbClient.GetItem(input)
	if err != nil {
		anlogger.Errorf(lc, "common_action.go : error get item for userId [%s], photoId [%s] and table [%s] : %v",
			userId, photoId, tableName, err)
		return nil, false, commons.InternalServerError
	}
	if len(result.Item) == 0 {
		anlogger.Warnf(lc, "common_action.go : there is no item for userId [%s], photoId [%s] and table [%s]",
			userId, photoId, tableName)
		return nil, true, ""
	}

	res := UserPhoto{
		Bucket: *result.Item[commons.PhotoBucketColumnName].S,
		Key:    *result.Item[commons.PhotoKeyColumnName].S,
	}
	anlogger.Debugf(lc, "common_action.go : successfully get userPhoto %v for userId [%s], photoId [%s] and table [%s]",
		res, userId, photoId, tableName)

	return &res, true, ""
}

//make GetItem
//return userId (owner), was everything ok and error string
func GetPhotoOwner(objectKey, photoUserMappingTableName string, dbClient dynamodbiface.DynamoDBAPI, anlogger *commons.Logger, lc *lambdacontext.LambdaContext) (string, bool, string) {
	anlogger.Debugf(lc, "common_action.go : find owner of object with a key [%s]", objectKey)
	input := &dynamodb.GetItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			commons.PhotoIdColumnName: {
				S: aws.String(objectKey),
			},
		},
		//if use consistent read then we will read from dynamodb directly
		//ConsistentRead: aws.Bool(true),
		TableName: aws.String(photoUserMappingTableName),
	}

	result, err := dbClient.GetItem(input)
	if err != nil {
		anlogger.Errorf(lc, "common_action.go : error reading owner by object key [%s] : %v", objectKey, err)
		return "", false, commons.InternalServerError
	}

	anlogger.Debugf(lc, "result : %v", result.Item)

	if len(result.Item) == 0 {
		anlogger.Warnf(lc, "common_action.go : there is no owner for object with key [%s]", objectKey)
		//we need such coz s3 call function async and in this case we don't need to retry
		return "", true, ""
	}

	userId := *result.Item[commons.UserIdColumnName].S
	anlogger.Debugf(lc, "common_action.go : found owner with userId [%s] for object key [%s]", userId, objectKey)
	return userId, true, ""
}

//make Update
//return was mapping created, need to retry and error string
func CreatePhotoIdUserIdMappingUpdate(photoId, userId, photoUserMappingTableName string, dbClient dynamodbiface.DynamoDBAPI, anlogger *commons.Logger, lc *lambdacontext.LambdaContext) (bool, bool, string) {
	anlogger.Debugf(lc, "common_action.go : create mapping between photoId [%s] and userId [%s]", photoId, userId)

	input :=
		&dynamodb.UpdateItemInput{
			ExpressionAttributeNames: map[string]*string{
				"#userId": aws.String(commons.UserIdColumnName),
				"#time":   aws.String(commons.UpdatedTimeColumnName),
			},
			ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
				":uV": {
					S: aws.String(userId),
				},
				":tV": {
					S: aws.String(time.Now().UTC().Format("2006-01-02-15-04-05.000")),
				},
			},
			Key: map[string]*dynamodb.AttributeValue{
				commons.PhotoIdColumnName: {
					S: aws.String(photoId),
				},
			},
			ConditionExpression: aws.String(fmt.Sprintf("attribute_not_exists(%v)", commons.PhotoIdColumnName)),

			TableName:        aws.String(photoUserMappingTableName),
			UpdateExpression: aws.String("SET #userId = :uV, #time = :tV"),
		}

	_, err := dbClient.UpdateItem(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case dynamodb.ErrCodeConditionalCheckFailedException:
				anlogger.Debugf(lc, "common_action.go : such photoId [%s] already in use, userId [%s]", photoId, userId)
				return false, true, ""
			}
		}
		anlogger.Errorf(lc, "common_action.go : error while create mapping between photoId [%s] and userId [%s] : %v", photoId, userId, err)
		return false, false, commons.InternalServerError
	}
	anlogger.Debugf(lc, "common_action.go : successfully create mapping between photoId [%s] and userId [%s]", photoId, userId)
	return true, false, ""
}

//make Update
func LikePhotoUpdate(userId, originPhotoId, userPhotoMetaPartitionKey, userPhotoTable string, dbClient dynamodbiface.DynamoDBAPI,
	anlogger *commons.Logger, lc *lambdacontext.LambdaContext) error {

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
					S: aws.String(originPhotoId),
				},
			},
			TableName:        aws.String(userPhotoTable),
			UpdateExpression: aws.String("ADD #like :likeV"),
		}

	_, err := dbClient.UpdateItem(input)
	if err != nil {
		anlogger.Errorf(lc, "common_action.go : error while update likes on photo with meta partition key [%s], original photoId [%s] for userId [%s] : %v",
			userPhotoMetaPartitionKey, originPhotoId, userId, err)
		return errors.New(fmt.Sprintf("error like photo %s : %v", originPhotoId, err))
	}
	anlogger.Debugf(lc, "common_action.go : successfully like photo [%s] for userId [%s]", originPhotoId, userId)
	return nil
}

//make Query
func AllUserPhotoIdQuery(partitionKey, userPhotoTable string, dbClient dynamodbiface.DynamoDBAPI, anlogger *commons.Logger, lc *lambdacontext.LambdaContext) ([]*dynamodb.AttributeValue, error) {
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
			ExclusiveStartKey: lastEvaluatedKey,
			//for using with DAX need to set false
			ConsistentRead:         aws.Bool(true),
			KeyConditionExpression: aws.String("#userId = :userIdV"),
			FilterExpression:       aws.String(fmt.Sprintf("attribute_not_exists(%s)", commons.PhotoDeletedAtColumnName)),
			TableName:              aws.String(userPhotoTable),
		}

		result, err := dbClient.Query(input)
		if err != nil {
			anlogger.Errorf(lc, "common_action.go : error while query all photos for partitionKey [%s] : %v", partitionKey, err)
			return finalResult, errors.New(fmt.Sprintf("error query all photos for partition key %s : %v", partitionKey, err))
		}

		lastEvaluatedKey = result.LastEvaluatedKey

		for _, item := range result.Items {
			finalResult = append(finalResult, item[commons.PhotoIdColumnName])
		}

		if len(lastEvaluatedKey) == 0 {
			anlogger.Debugf(lc, "common_action.go : all photo ids size is [%d] for partition key [%s]", len(finalResult), partitionKey)
			return finalResult, nil
		}
	}
}

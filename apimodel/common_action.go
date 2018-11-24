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
)

//return ok and error string
func SavePhoto(userPhoto *UserPhoto, userPhotoTable string, awsDbClient *dynamodb.DynamoDB, anlogger *commons.Logger, lc *lambdacontext.LambdaContext) (bool, string) {
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

	_, err := awsDbClient.UpdateItem(input)
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

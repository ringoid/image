package main

import (
	"context"
	basicLambda "github.com/aws/aws-lambda-go/lambda"
	"../sys_log"
	"../apimodel"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/firehose"
	"os"
	"fmt"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-lambda-go/events"
	"time"
	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/aws/aws-sdk-go/service/lambda"
	"errors"
	"strings"
	"strconv"
	"github.com/aws/aws-sdk-go/service/s3"
)

var anlogger *syslog.Logger
var awsDbClient *dynamodb.DynamoDB
var awsDeliveryStreamClient *firehose.Firehose
var deliveryStreamName string
var internalAuthFunctionName string
var presignFunctionName string
var clientLambda *lambda.Lambda
var photoUserMappingTableName string
var originPhotoBucketName string
var publicPhotoBucketName string
var userPhotoTable string
var awsS3Client *s3.S3

func init() {
	var env string
	var ok bool
	var papertrailAddress string
	var err error
	var awsSession *session.Session

	env, ok = os.LookupEnv("ENV")
	if !ok {
		fmt.Printf("internal_handle_upload.go : env can not be empty ENV")
		os.Exit(1)
	}
	fmt.Printf("internal_handle_upload.go : start with ENV = [%s]", env)

	papertrailAddress, ok = os.LookupEnv("PAPERTRAIL_LOG_ADDRESS")
	if !ok {
		fmt.Printf("internal_handle_upload.go : env can not be empty PAPERTRAIL_LOG_ADDRESS")
		os.Exit(1)
	}
	fmt.Printf("internal_handle_upload.go : start with PAPERTRAIL_LOG_ADDRESS = [%s]", papertrailAddress)

	anlogger, err = syslog.New(papertrailAddress, fmt.Sprintf("%s-%s", env, "internal-handle-upload-image"))
	if err != nil {
		fmt.Errorf("internal_handle_upload.go : error during startup : %v", err)
		os.Exit(1)
	}
	anlogger.Debugf(nil, "internal_handle_upload.go : logger was successfully initialized")

	internalAuthFunctionName, ok = os.LookupEnv("INTERNAL_AUTH_FUNCTION_NAME")
	if !ok {
		fmt.Printf("internal_handle_upload.go : env can not be empty INTERNAL_AUTH_FUNCTION_NAME")
		os.Exit(1)
	}
	anlogger.Debugf(nil, "internal_handle_upload.go : start with INTERNAL_AUTH_FUNCTION_NAME = [%s]", internalAuthFunctionName)

	presignFunctionName, ok = os.LookupEnv("PRESIGN_FUNCTION_NAME")
	if !ok {
		fmt.Printf("internal_handle_upload.go : env can not be empty PRESIGN_FUNCTION_NAME")
		os.Exit(1)
	}
	anlogger.Debugf(nil, "internal_handle_upload.go : start with PRESIGN_FUNCTION_NAME = [%s]", presignFunctionName)

	photoUserMappingTableName, ok = os.LookupEnv("PHOTO_USER_MAPPING_TABLE")
	if !ok {
		fmt.Printf("internal_handle_upload.go : env can not be empty PHOTO_USER_MAPPING_TABLE")
		os.Exit(1)
	}
	anlogger.Debugf(nil, "internal_handle_upload.go : start with PHOTO_USER_MAPPING_TABLE = [%s]", photoUserMappingTableName)

	originPhotoBucketName, ok = os.LookupEnv("ORIGIN_PHOTO_BUCKET_NAME")
	if !ok {
		fmt.Printf("internal_handle_upload.go : env can not be empty ORIGIN_PHOTO_BUCKET_NAME")
		os.Exit(1)
	}
	anlogger.Debugf(nil, "internal_handle_upload.go : start with ORIGIN_PHOTO_BUCKET_NAME = [%s]", originPhotoBucketName)

	publicPhotoBucketName, ok = os.LookupEnv("PUBLIC_PHOTO_BUCKET_NAME")
	if !ok {
		fmt.Printf("internal_handle_upload.go : env can not be empty PUBLIC_PHOTO_BUCKET_NAME")
		os.Exit(1)
	}
	anlogger.Debugf(nil, "internal_handle_upload.go : start with PUBLIC_PHOTO_BUCKET_NAME = [%s]", publicPhotoBucketName)

	userPhotoTable, ok = os.LookupEnv("USER_PHOTO_TABLE")
	if !ok {
		fmt.Printf("internal_handle_upload.go : env can not be empty USER_PHOTO_TABLE")
		os.Exit(1)
	}
	anlogger.Debugf(nil, "internal_handle_upload.go : start with USER_PHOTO_TABLE = [%s]", userPhotoTable)

	awsSession, err = session.NewSession(aws.NewConfig().
		WithRegion(apimodel.Region).WithMaxRetries(apimodel.MaxRetries).
		WithLogger(aws.LoggerFunc(func(args ...interface{}) { anlogger.AwsLog(args) })).WithLogLevel(aws.LogOff))
	if err != nil {
		anlogger.Fatalf(nil, "internal_handle_upload.go : error during initialization : %v", err)
	}
	anlogger.Debugf(nil, "internal_handle_upload.go : aws session was successfully initialized")

	awsDbClient = dynamodb.New(awsSession)
	anlogger.Debugf(nil, "internal_handle_upload.go : dynamodb client was successfully initialized")

	clientLambda = lambda.New(awsSession)
	anlogger.Debugf(nil, "internal_handle_upload.go : lambda client was successfully initialized")

	awsS3Client = s3.New(awsSession)
	anlogger.Debugf(nil, "internal_handle_upload.go : s3 client was successfully initialized")

	deliveryStreamName, ok = os.LookupEnv("DELIVERY_STREAM")
	if !ok {
		anlogger.Fatalf(nil, "internal_handle_upload.go : env can not be empty DELIVERY_STREAM")
		os.Exit(1)
	}
	anlogger.Debugf(nil, "internal_handle_upload.go : start with DELIVERY_STREAM = [%s]", deliveryStreamName)

	awsDeliveryStreamClient = firehose.New(awsSession)
	anlogger.Debugf(nil, "internal_handle_upload.go : firehose client was successfully initialized")
}

func handler(ctx context.Context, request events.S3Event) (error) {
	lc, _ := lambdacontext.FromContext(ctx)
	anlogger.Debugf(lc, "internal_handle_upload.go : start handle request %v", request)

	for _, record := range request.Records {
		objectBucket := record.S3.Bucket.Name
		objectKey := record.S3.Object.Key
		decodedObjectKey := record.S3.Object.URLDecodedKey
		objectSize := record.S3.Object.Size
		anlogger.Debugf(lc, "internal_handle_upload.go : object was uploaded with bucket [%s], objectKey [%s], decodedObjectKey [%s], objectSize [%v]",
			objectBucket, objectKey, decodedObjectKey, objectSize)

		userId, ok, errStr := getOwner(objectKey, lc)
		if !ok {
			return errors.New(errStr)
		}

		//it means that there is no owner for this photo
		if userId == "" {
			return nil
		}

		//now construct photo object
		originPhotoId := strings.Split(objectKey, "_photo")[0]
		photoId := "origin_" + originPhotoId

		userPhoto := apimodel.UserPhoto{
			UserId:    userId,
			PhotoId:   photoId,
			PhotoType: "origin",
			Bucket:    objectBucket,
			Key:       objectKey,
			Size:      objectSize,
		}

		ok, errStr = savePhoto(&userPhoto, lc)
		if !ok {
			return errors.New(errStr)
		}
		anlogger.Infof(lc, "internal_handle_upload.go : successfully save origin photo %v for userId [%s]", userPhoto, userPhoto.UserId)

		event := apimodel.NewUserUploadedPhotoEvent(userPhoto)
		apimodel.SendAnalyticEvent(event, userPhoto.UserId, deliveryStreamName, awsDeliveryStreamClient, anlogger, lc)

		//Now transform foto to fake 640x480
		userPhoto = apimodel.UserPhoto{
			UserId:    userId,
			PhotoId:   "640x480_" + originPhotoId,
			PhotoType: "640x480",
			Bucket:    publicPhotoBucketName,
			Key:       "640x480_" + objectKey,
			Size:      objectSize,
		}
		link, ok, errStr := transformImage(
			userPhoto.Bucket, userPhoto.Key, objectBucket, objectKey, userPhoto.UserId, lc)
		if !ok {
			return errors.New(errStr)
		}

		userPhoto.PhotoSourceUri = link
		ok, errStr = savePhoto(&userPhoto, lc)
		if !ok {
			return errors.New(errStr)
		}
		anlogger.Infof(lc, "internal_handle_upload.go : successfully save 640x480 photo %v for userId [%s]", userPhoto, userPhoto.UserId)
		//end transformation
	}

	return nil
}

//return public link, ok and error string
func transformImage(targetBucket, targetKey, sourceBucket, sourceKey, userId string, lc *lambdacontext.LambdaContext) (string, bool, string) {
	anlogger.Debugf(lc, "internal_handle_upload.go : copy [%s/%s] to [%s/%s] for userId [%s]",
		sourceBucket, sourceKey, targetBucket, targetKey, userId)
	_, err := awsS3Client.CopyObject(&s3.CopyObjectInput{Bucket: aws.String(targetBucket),
		CopySource: aws.String(sourceBucket + "/" + sourceKey), Key: aws.String(targetKey), ACL: aws.String("public-read")})
	if err != nil {
		anlogger.Errorf(lc, "internal_handle_upload.go : error copy [%s/%s] to [%s/%s] for userId [%s] : %v",
			sourceBucket, sourceKey, targetBucket, targetKey, userId, err)
		return "", false, apimodel.InternalServerError
	}
	err = awsS3Client.WaitUntilObjectExists(&s3.HeadObjectInput{Bucket: aws.String(targetBucket), Key: aws.String(targetKey)})
	if err != nil {
		anlogger.Errorf(lc, "internal_handle_upload.go : error waiting while copy [%s/%s] to [%s/%s] for userId [%s] complete : %v",
			sourceBucket, sourceKey, targetBucket, targetKey, userId, err)
		return "", false, apimodel.InternalServerError
	}
	link := fmt.Sprintf("https://s3-eu-west-1.amazonaws.com/%s/%s", targetBucket, targetKey)
	anlogger.Debugf(lc, "internal_handle_upload.go : successfully copy [%s/%s] to [%s/%s] with public link [%s] for userId [%s]",
		sourceBucket, sourceKey, targetBucket, targetKey, link, userId)

	return link, true, ""
}

//return userId (owner), was everything ok and error string
func getOwner(objectKey string, lc *lambdacontext.LambdaContext) (string, bool, string) {
	anlogger.Debugf(lc, "internal_handle_upload.go : find owner of object with a key [%s]", objectKey)
	input := &dynamodb.GetItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			apimodel.PhotoIdColumnName: {
				S: aws.String(objectKey),
			},
		},
		ConsistentRead: aws.Bool(true),
		TableName:      aws.String(photoUserMappingTableName),
	}

	result, err := awsDbClient.GetItem(input)
	if err != nil {
		anlogger.Errorf(lc, "internal_handle_upload.go : error reading owner by object key [%s] : %v", objectKey, err)
		return "", false, apimodel.InternalServerError
	}

	anlogger.Debugf(lc, "result : %v", result.Item)

	if len(result.Item) == 0 {
		anlogger.Warnf(lc, "internal_handle_upload.go : there is no owner for object with key [%s]", objectKey)
		//we need such coz s3 call function async and in this case we don't need to retry
		return "", true, ""
	}

	userId := *result.Item[apimodel.UserIdColumnName].S
	anlogger.Debugf(lc, "internal_handle_upload.go : found owner with userId [%s] for object key [%s]", userId, objectKey)
	return userId, true, ""
}

//return ok and error string
func savePhoto(userPhoto *apimodel.UserPhoto, lc *lambdacontext.LambdaContext) (bool, string) {
	anlogger.Debugf(lc, "internal_handle_upload.go : save photo %v for userId [%s]", userPhoto, userPhoto.UserId)
	input :=
		&dynamodb.UpdateItemInput{
			ExpressionAttributeNames: map[string]*string{
				"#photoType":   aws.String(apimodel.PhotoTypeColumnName),
				"#photoBucket": aws.String(apimodel.PhotoBucketColumnName),
				"#photoKey":    aws.String(apimodel.PhotoKeyColumnName),
				"#photoSize":   aws.String(apimodel.PhotoSizeColumnName),
				"#updatedAt":   aws.String(apimodel.UpdatedTimeColumnName),
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
				apimodel.UserIdColumnName: {
					S: aws.String(userPhoto.UserId),
				},
				apimodel.PhotoIdColumnName: {
					S: aws.String(userPhoto.PhotoId),
				},
			},
			TableName:        aws.String(userPhotoTable),
			UpdateExpression: aws.String("SET #photoType = :photoTypeV, #photoBucket = :photoBucketV, #photoKey = :photoKeyV, #photoSize = :photoSizeV, #updatedAt = :updatedAtV"),
		}

	if userPhoto.PhotoSourceUri != "" {
		input.ExpressionAttributeNames["#photoUri"] = aws.String(apimodel.PhotoSourceUriColumnName)
		input.ExpressionAttributeValues[":photoUriV"] = &dynamodb.AttributeValue{
			S: aws.String(userPhoto.PhotoSourceUri),
		}
		input.UpdateExpression = aws.String("SET #photoUri = :photoUriV, #photoType = :photoTypeV, #photoBucket = :photoBucketV, #photoKey = :photoKeyV, #photoSize = :photoSizeV, #updatedAt = :updatedAtV")
	}

	_, err := awsDbClient.UpdateItem(input)
	if err != nil {
		anlogger.Errorf(lc, "internal_handle_upload.go : error while save photo %v for userId [%s] : %v", userPhoto, userPhoto.UserId, err)
		return false, apimodel.InternalServerError
	}
	anlogger.Debugf(lc, "internal_handle_upload.go : successfully save photo %v for userId [%s]", userPhoto, userPhoto.UserId)
	return true, ""
}

func main() {
	basicLambda.Start(handler)
}

package main

import (
	"context"
	basicLambda "github.com/aws/aws-lambda-go/lambda"
	"../apimodel"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/firehose"
	"os"
	"fmt"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/aws/aws-sdk-go/service/lambda"
	"errors"
	"strings"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/aws/aws-sdk-go/service/kinesis"
	"github.com/ringoid/commons"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	"github.com/aws/aws-dax-go/dax"
)

var anlogger *commons.Logger
var daxClient dynamodbiface.DynamoDBAPI
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
var downloader *s3manager.Downloader
var uploader *s3manager.Uploader
var awsSqsClient *sqs.SQS
var asyncTaskQueue string
var commonStreamName string
var awsKinesisClient *kinesis.Kinesis

func init() {
	var env string
	var ok bool
	var papertrailAddress string
	var err error
	var awsSession *session.Session

	env, ok = os.LookupEnv("ENV")
	if !ok {
		fmt.Printf("lambda-initialization : internal_handle_upload.go : env can not be empty ENV\n")
		os.Exit(1)
	}
	fmt.Printf("lambda-initialization : internal_handle_upload.go : start with ENV = [%s]\n", env)

	papertrailAddress, ok = os.LookupEnv("PAPERTRAIL_LOG_ADDRESS")
	if !ok {
		fmt.Printf("lambda-initialization : internal_handle_upload.go : env can not be empty PAPERTRAIL_LOG_ADDRESS\n")
		os.Exit(1)
	}
	fmt.Printf("lambda-initialization : internal_handle_upload.go : start with PAPERTRAIL_LOG_ADDRESS = [%s]\n", papertrailAddress)

	anlogger, err = commons.New(papertrailAddress, fmt.Sprintf("%s-%s", env, "internal-handle-upload-image"))
	if err != nil {
		fmt.Errorf("lambda-initialization : internal_handle_upload.go : error during startup : %v\n", err)
		os.Exit(1)
	}
	anlogger.Debugf(nil, "lambda-initialization : internal_handle_upload.go : logger was successfully initialized")

	internalAuthFunctionName, ok = os.LookupEnv("INTERNAL_AUTH_FUNCTION_NAME")
	if !ok {
		anlogger.Fatalf(nil, "lambda-initialization : internal_handle_upload.go : env can not be empty INTERNAL_AUTH_FUNCTION_NAME")
	}
	anlogger.Debugf(nil, "lambda-initialization : internal_handle_upload.go : start with INTERNAL_AUTH_FUNCTION_NAME = [%s]", internalAuthFunctionName)

	presignFunctionName, ok = os.LookupEnv("PRESIGN_FUNCTION_NAME")
	if !ok {
		anlogger.Fatalf(nil, "lambda-initialization : internal_handle_upload.go : env can not be empty PRESIGN_FUNCTION_NAME")
	}
	anlogger.Debugf(nil, "lambda-initialization : internal_handle_upload.go : start with PRESIGN_FUNCTION_NAME = [%s]", presignFunctionName)

	photoUserMappingTableName, ok = os.LookupEnv("PHOTO_USER_MAPPING_TABLE")
	if !ok {
		anlogger.Fatalf(nil, "lambda-initialization : internal_handle_upload.go : env can not be empty PHOTO_USER_MAPPING_TABLE")
	}
	anlogger.Debugf(nil, "lambda-initialization : internal_handle_upload.go : start with PHOTO_USER_MAPPING_TABLE = [%s]", photoUserMappingTableName)

	originPhotoBucketName, ok = os.LookupEnv("ORIGIN_PHOTO_BUCKET_NAME")
	if !ok {
		anlogger.Fatalf(nil, "lambda-initialization : internal_handle_upload.go : env can not be empty ORIGIN_PHOTO_BUCKET_NAME")
	}
	anlogger.Debugf(nil, "lambda-initialization : internal_handle_upload.go : start with ORIGIN_PHOTO_BUCKET_NAME = [%s]", originPhotoBucketName)

	publicPhotoBucketName, ok = os.LookupEnv("PUBLIC_PHOTO_BUCKET_NAME")
	if !ok {
		anlogger.Fatalf(nil, "lambda-initialization : internal_handle_upload.go : env can not be empty PUBLIC_PHOTO_BUCKET_NAME")
	}
	anlogger.Debugf(nil, "lambda-initialization : internal_handle_upload.go : start with PUBLIC_PHOTO_BUCKET_NAME = [%s]", publicPhotoBucketName)

	userPhotoTable, ok = os.LookupEnv("USER_PHOTO_TABLE")
	if !ok {
		anlogger.Fatalf(nil, "lambda-initialization : internal_handle_upload.go : env can not be empty USER_PHOTO_TABLE")
	}
	anlogger.Debugf(nil, "lambda-initialization : internal_handle_upload.go : start with USER_PHOTO_TABLE = [%s]", userPhotoTable)

	asyncTaskQueue, ok = os.LookupEnv("ASYNC_TASK_SQS_QUEUE")
	if !ok {
		anlogger.Fatalf(nil, "lambda-initialization : internal_handle_upload.go : env can not be empty ASYNC_TASK_SQS_QUEUE")
	}
	anlogger.Debugf(nil, "lambda-initialization : internal_handle_upload.go : start with ASYNC_TASK_SQS_QUEUE = [%s]", asyncTaskQueue)

	awsSession, err = session.NewSession(aws.NewConfig().
		WithRegion(commons.Region).WithMaxRetries(commons.MaxRetries).
		WithLogger(aws.LoggerFunc(func(args ...interface{}) { anlogger.AwsLog(args) })).WithLogLevel(aws.LogOff))
	if err != nil {
		anlogger.Fatalf(nil, "lambda-initialization : internal_handle_upload.go : error during initialization : %v", err)
	}
	anlogger.Debugf(nil, "lambda-initialization : internal_handle_upload.go : aws session was successfully initialized")

	daxEndpoint, ok := os.LookupEnv("DAX_ENDPOINT")
	if !ok {
		anlogger.Fatalf(nil, "lambda-initialization : internal_handle_upload.go : env can not be empty DAX_ENDPOINT")
	}
	cfg := dax.DefaultConfig()
	cfg.HostPorts = []string{daxEndpoint}
	cfg.Region = commons.Region
	daxClient, err = dax.New(cfg)
	if err != nil {
		anlogger.Fatalf(nil, "lambda-initialization : internal_handle_upload.go : error initialize DAX cluster")
	}
	anlogger.Debugf(nil, "lambda-initialization : internal_handle_upload.go : dax client was successfully initialized")

	clientLambda = lambda.New(awsSession)
	anlogger.Debugf(nil, "lambda-initialization : internal_handle_upload.go : lambda client was successfully initialized")

	awsS3Client = s3.New(awsSession)
	anlogger.Debugf(nil, "lambda-initialization : internal_handle_upload.go : s3 client was successfully initialized")

	downloader = s3manager.NewDownloader(awsSession)
	anlogger.Debugf(nil, "lambda-initialization : internal_handle_upload.go : s3 downloader was successfully initialized")

	uploader = s3manager.NewUploader(awsSession)
	anlogger.Debugf(nil, "lambda-initialization : internal_handle_upload.go : s3 uploader was successfully initialized")

	deliveryStreamName, ok = os.LookupEnv("DELIVERY_STREAM")
	if !ok {
		anlogger.Fatalf(nil, "lambda-initialization : internal_handle_upload.go : env can not be empty DELIVERY_STREAM")
		os.Exit(1)
	}
	anlogger.Debugf(nil, "lambda-initialization : internal_handle_upload.go : start with DELIVERY_STREAM = [%s]", deliveryStreamName)

	awsDeliveryStreamClient = firehose.New(awsSession)
	anlogger.Debugf(nil, "lambda-initialization : internal_handle_upload.go : firehose client was successfully initialized")

	commonStreamName, ok = os.LookupEnv("COMMON_STREAM")
	if !ok {
		anlogger.Fatalf(nil, "lambda-initialization : internal_handle_upload.go : env can not be empty COMMON_STREAM")
		os.Exit(1)
	}
	anlogger.Debugf(nil, "lambda-initialization : internal_handle_upload.go : start with DELIVERY_STREAM = [%s]", commonStreamName)

	awsKinesisClient = kinesis.New(awsSession)
	anlogger.Debugf(nil, "lambda-initialization : internal_handle_upload.go : kinesis client was successfully initialized")

	awsSqsClient = sqs.New(awsSession)
	anlogger.Debugf(nil, "lambda-initialization : internal_handle_upload.go : sqs client was successfully initialized")
}

func handler(ctx context.Context, request events.S3Event) (error) {
	lc, _ := lambdacontext.FromContext(ctx)
	anlogger.Debugf(lc, "internal_handle_upload.go : start handle upload photo request %v", request)

	for _, record := range request.Records {
		objectBucket := record.S3.Bucket.Name
		objectKey := record.S3.Object.Key
		objectSize := record.S3.Object.Size

		anlogger.Debugf(lc, "internal_handle_upload.go : object was uploaded with bucket [%s], objectKey [%s], objectSize [%v]",
			objectBucket, objectKey, objectSize)

		userId, ok, errStr := apimodel.GetPhotoOwner(objectKey, photoUserMappingTableName, daxClient, anlogger, lc)
		if !ok {
			return errors.New(errStr)
		}

		//it means that there is no owner for this photo
		if userId == "" {
			return nil
		}

		//todo: uncomment before prod
		//if objectSize >= commons.DefaultMaxPhotoSize {
		//	anlogger.Warnf(lc, "internal_handle_upload.go : uploaded object too big, bucket [%s], objectKey [%s], objectSize [%v] for userId [%s]",
		//		objectBucket, objectKey, objectSize, userId)
		//	task := apimodel.NewRemoveS3ObjectAsyncTask(objectBucket, objectKey)
		//	ok, errStr = apimodel.SendAsyncTask(task, asyncTaskQueue, userId, awsSqsClient, anlogger, lc)
		//	if !ok {
		//		return errors.New(errStr)
		//	}
		//	event := apimodel.NewRemoveTooLargeObjectEvent(userId, objectBucket, objectKey, objectSize)
		//	apimodel.SendAnalyticEvent(event, userId, deliveryStreamName, awsDeliveryStreamClient, anlogger, lc)
		//	return nil
		//}

		//now construct photo object
		arr := strings.Split(objectKey, "_photo")
		originS3PhotoId := arr[0]
		extension := arr[1]

		photoId := "origin_" + originS3PhotoId

		userPhoto := apimodel.UserPhoto{
			UserId:    userId,
			PhotoId:   photoId,
			PhotoType: "origin",
			Bucket:    objectBucket,
			Key:       objectKey,
			Size:      objectSize,
		}

		ok, errStr = apimodel.SavePhotoUpdate(&userPhoto, userPhotoTable, daxClient, anlogger, lc)
		if !ok && len(errStr) != 0 {
			return errors.New(errStr)
		} else if !ok && len(errStr) == 0 {
			anlogger.Warnf(lc, "internal_handle_upload.go : uploaded object was already deleted, bucket [%s], objectKey [%s], objectSize [%v] for userId [%s]",
				objectBucket, objectKey, objectSize, userId)
			task := apimodel.NewRemoveS3ObjectAsyncTask(objectBucket, objectKey)
			ok, errStr = commons.SendAsyncTask(task, asyncTaskQueue, userId, 0, awsSqsClient, anlogger, lc)
			if !ok {
				return errors.New(errStr)
			}
			anlogger.Debugf(lc, "internal_handle_upload.go : successfully handle photo upload request %v", request)
			return nil
		}

		anlogger.Infof(lc, "internal_handle_upload.go : successfully save origin photo %v for userId [%s]", userPhoto, userPhoto.UserId)

		event := commons.NewUserUploadedPhotoEvent(userPhoto.UserId, userPhoto.Bucket, userPhoto.Key, userPhoto.PhotoId, userPhoto.PhotoType, userPhoto.Size)
		commons.SendAnalyticEvent(event, userPhoto.UserId, deliveryStreamName, awsDeliveryStreamClient, anlogger, lc)

		partitionKey := userId
		ok, errStr = commons.SendCommonEvent(event, userId, commonStreamName, partitionKey, awsKinesisClient, anlogger, lc)
		if !ok {
			return errors.New(errStr)
		}

		for resolution := range commons.AllowedPhotoResolution {
			width := commons.ResolutionValues[resolution+"_width"]
			height := commons.ResolutionValues[resolution+"_height"]
			resizedPhotoId := resolution + "_" + originS3PhotoId
			targetKey := originS3PhotoId + "_" + resolution + extension
			task := apimodel.NewResizePhotoAsyncTask(userId, resizedPhotoId, resolution, objectBucket, objectKey, publicPhotoBucketName, targetKey, userPhotoTable, width, height)
			ok, errStr = commons.SendAsyncTask(task, asyncTaskQueue, userId, 0, awsSqsClient, anlogger, lc)
			if !ok {
				return errors.New(errStr)
			}
		}
	}
	anlogger.Debugf(lc, "internal_handle_upload.go : successfully handle photo upload request %v", request)
	return nil
}

func main() {
	basicLambda.Start(handler)
}

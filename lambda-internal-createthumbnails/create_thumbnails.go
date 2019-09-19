package main

import (
	"context"
	basicLambda "github.com/aws/aws-lambda-go/lambda"
	"../apimodel"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/firehose"
	"os"
	"fmt"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/aws/aws-sdk-go/service/lambda"
	"errors"
	"strings"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/aws/aws-sdk-go/service/kinesis"
	"github.com/ringoid/commons"
	"encoding/json"
)

var anlogger *commons.Logger
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
var downloader *s3manager.Downloader
var uploader *s3manager.Uploader
var awsSqsClient *sqs.SQS
var asyncTaskQueue string
var commonStreamName string
var awsKinesisClient *kinesis.Kinesis

var env string

var getConvertObjectsFunctionNameMap map[string]string
var iamFunctionNameMap map[string]string
var originBucketName map[string]string

func init() {
	var ok bool
	var papertrailAddress string
	var err error
	var awsSession *session.Session

	env, ok = os.LookupEnv("ENV")
	if !ok {
		fmt.Printf("lambda-initialization : create_thumbnails.go : env can not be empty ENV\n")
		os.Exit(1)
	}
	fmt.Printf("lambda-initialization : create_thumbnails.go : start with ENV = [%s]\n", env)

	papertrailAddress, ok = os.LookupEnv("PAPERTRAIL_LOG_ADDRESS")
	if !ok {
		fmt.Printf("lambda-initialization : create_thumbnails.go : env can not be empty PAPERTRAIL_LOG_ADDRESS\n")
		os.Exit(1)
	}
	fmt.Printf("lambda-initialization : create_thumbnails.go : start with PAPERTRAIL_LOG_ADDRESS = [%s]\n", papertrailAddress)

	anlogger, err = commons.New(papertrailAddress, fmt.Sprintf("%s-%s", env, "internal-create-thumbnails-image"), apimodel.IsDebugLogEnabled)
	if err != nil {
		fmt.Errorf("lambda-initialization : create_thumbnails.go : error during startup : %v\n", err)
		os.Exit(1)
	}
	anlogger.Debugf(nil, "lambda-initialization : create_thumbnails.go : logger was successfully initialized")

	internalAuthFunctionName, ok = os.LookupEnv("INTERNAL_AUTH_FUNCTION_NAME")
	if !ok {
		anlogger.Fatalf(nil, "lambda-initialization : create_thumbnails.go : env can not be empty INTERNAL_AUTH_FUNCTION_NAME")
	}
	anlogger.Debugf(nil, "lambda-initialization : create_thumbnails.go : start with INTERNAL_AUTH_FUNCTION_NAME = [%s]", internalAuthFunctionName)

	presignFunctionName, ok = os.LookupEnv("PRESIGN_FUNCTION_NAME")
	if !ok {
		anlogger.Fatalf(nil, "lambda-initialization : create_thumbnails.go : env can not be empty PRESIGN_FUNCTION_NAME")
	}
	anlogger.Debugf(nil, "lambda-initialization : create_thumbnails.go : start with PRESIGN_FUNCTION_NAME = [%s]", presignFunctionName)

	photoUserMappingTableName, ok = os.LookupEnv("PHOTO_USER_MAPPING_TABLE")
	if !ok {
		anlogger.Fatalf(nil, "lambda-initialization : create_thumbnails.go : env can not be empty PHOTO_USER_MAPPING_TABLE")
	}
	anlogger.Debugf(nil, "lambda-initialization : create_thumbnails.go : start with PHOTO_USER_MAPPING_TABLE = [%s]", photoUserMappingTableName)

	originPhotoBucketName, ok = os.LookupEnv("ORIGIN_PHOTO_BUCKET_NAME")
	if !ok {
		anlogger.Fatalf(nil, "lambda-initialization : create_thumbnails.go : env can not be empty ORIGIN_PHOTO_BUCKET_NAME")
	}
	anlogger.Debugf(nil, "lambda-initialization : create_thumbnails.go : start with ORIGIN_PHOTO_BUCKET_NAME = [%s]", originPhotoBucketName)

	publicPhotoBucketName, ok = os.LookupEnv("PUBLIC_PHOTO_BUCKET_NAME")
	if !ok {
		anlogger.Fatalf(nil, "lambda-initialization : create_thumbnails.go : env can not be empty PUBLIC_PHOTO_BUCKET_NAME")
	}
	anlogger.Debugf(nil, "lambda-initialization : create_thumbnails.go : start with PUBLIC_PHOTO_BUCKET_NAME = [%s]", publicPhotoBucketName)

	userPhotoTable, ok = os.LookupEnv("USER_PHOTO_TABLE")
	if !ok {
		anlogger.Fatalf(nil, "lambda-initialization : create_thumbnails.go : env can not be empty USER_PHOTO_TABLE")
	}
	anlogger.Debugf(nil, "lambda-initialization : create_thumbnails.go : start with USER_PHOTO_TABLE = [%s]", userPhotoTable)

	asyncTaskQueue, ok = os.LookupEnv("ASYNC_TASK_SQS_QUEUE")
	if !ok {
		anlogger.Fatalf(nil, "lambda-initialization : create_thumbnails.go : env can not be empty ASYNC_TASK_SQS_QUEUE")
	}
	anlogger.Debugf(nil, "lambda-initialization : create_thumbnails.go : start with ASYNC_TASK_SQS_QUEUE = [%s]", asyncTaskQueue)

	awsSession, err = session.NewSession(aws.NewConfig().
		WithRegion(commons.Region).WithMaxRetries(commons.MaxRetries).
		WithLogger(aws.LoggerFunc(func(args ...interface{}) { anlogger.AwsLog(args) })).WithLogLevel(aws.LogOff))
	if err != nil {
		anlogger.Fatalf(nil, "lambda-initialization : create_thumbnails.go : error during initialization : %v", err)
	}
	anlogger.Debugf(nil, "lambda-initialization : create_thumbnails.go : aws session was successfully initialized")

	awsDbClient = dynamodb.New(awsSession)
	anlogger.Debugf(nil, "lambda-initialization : create_thumbnails.go : dynamodb client was successfully initialized")

	clientLambda = lambda.New(awsSession)
	anlogger.Debugf(nil, "lambda-initialization : create_thumbnails.go : lambda client was successfully initialized")

	awsS3Client = s3.New(awsSession)
	anlogger.Debugf(nil, "lambda-initialization : create_thumbnails.go : s3 client was successfully initialized")

	downloader = s3manager.NewDownloader(awsSession)
	anlogger.Debugf(nil, "lambda-initialization : create_thumbnails.go : s3 downloader was successfully initialized")

	uploader = s3manager.NewUploader(awsSession)
	anlogger.Debugf(nil, "lambda-initialization : create_thumbnails.go : s3 uploader was successfully initialized")

	deliveryStreamName, ok = os.LookupEnv("DELIVERY_STREAM")
	if !ok {
		anlogger.Fatalf(nil, "lambda-initialization : create_thumbnails.go : env can not be empty DELIVERY_STREAM")
		os.Exit(1)
	}
	anlogger.Debugf(nil, "lambda-initialization : create_thumbnails.go : start with DELIVERY_STREAM = [%s]", deliveryStreamName)

	awsDeliveryStreamClient = firehose.New(awsSession)
	anlogger.Debugf(nil, "lambda-initialization : create_thumbnails.go : firehose client was successfully initialized")

	commonStreamName, ok = os.LookupEnv("COMMON_STREAM")
	if !ok {
		anlogger.Fatalf(nil, "lambda-initialization : create_thumbnails.go : env can not be empty COMMON_STREAM")
		os.Exit(1)
	}
	anlogger.Debugf(nil, "lambda-initialization : create_thumbnails.go : start with DELIVERY_STREAM = [%s]", commonStreamName)

	awsKinesisClient = kinesis.New(awsSession)
	anlogger.Debugf(nil, "lambda-initialization : create_thumbnails.go : kinesis client was successfully initialized")

	awsSqsClient = sqs.New(awsSession)
	anlogger.Debugf(nil, "lambda-initialization : create_thumbnails.go : sqs client was successfully initialized")

	getConvertObjectsFunctionNameMap = make(map[string]string)
	getConvertObjectsFunctionNameMap["test"] = "test-internal-getConvertObjectsFunction-relationships"
	getConvertObjectsFunctionNameMap["stage"] = "stage-internal-getConvertObjectsFunction-relationships"
	getConvertObjectsFunctionNameMap["prod"] = "prod-internal-getConvertObjectsFunction-relationships"

	iamFunctionNameMap = make(map[string]string)
	iamFunctionNameMap["test"] = "test-internal-create-thumbnailsFunction-image"
	iamFunctionNameMap["stage"] = "stage-internal-create-thumbnailsFunction-image"
	iamFunctionNameMap["prod"] = "prod-internal-create-thumbnailsFunction-image"

	originBucketName = make(map[string]string)
	originBucketName["test"] = "test-ringoid-origin-photo"
	originBucketName["stage"] = "stage-ringoid-origin-photo"
	originBucketName["prod"] = "prod-ringoid-origin-photo"
}

type ConvertRequest struct {
	Skip  int `json:"skip"`
	Limit int `json:"limit"`
}

type ConvertResponse struct {
	Objects []ConvertObject `json:"objects"`
}

type ConvertObject struct {
	UserId    string `json:"userId"`
	ObjectKey string `json:"objectKey"`
}

func handler(ctx context.Context, request ConvertRequest) (error) {

	lc, _ := lambdacontext.FromContext(ctx)

	if request.Limit <= 0 || request.Limit >= 1000 {
		request.Limit = 1000
	}

	anlogger.Infof(lc, "create_thumbnails.go : start handle create thumbnails for request %v", request)

	convertResponse, err := sendRequest(request.Skip, request.Limit, getConvertObjectsFunctionNameMap[env], lc)
	if err != nil {
		return err
	}

	for _, each := range convertResponse.Objects {

		objectBucket := originBucketName[env]
		objectKey := each.ObjectKey
		userId := each.UserId

		anlogger.Debugf(lc, "create_thumbnails.go : object was uploaded with bucket [%s], objectKey [%s]",
			objectBucket, objectKey)

		//now construct photo object
		arr := strings.Split(objectKey, "_photo")
		originS3PhotoId := arr[0]
		extension := arr[1]

		photoId := "origin_" + originS3PhotoId

		//make thumbnail
		thumbnailResizedPhotoId := commons.ThumbnailPhotoType + "_" + originS3PhotoId
		thumbnailTargetKey := originS3PhotoId + "_" + commons.ThumbnailPhotoType + extension
		task := apimodel.NewResizePhotoAsyncTask(userId, photoId, thumbnailResizedPhotoId, commons.ThumbnailPhotoType, commons.ThumbnailJPEGQuality, objectBucket, objectKey, publicPhotoBucketName, thumbnailTargetKey, userPhotoTable, commons.ThumbnailPhotoWidth, commons.ThumbnailPhotoHeight)
		ok, errStr := commons.SendAsyncTask(task, asyncTaskQueue, userId, 0, awsSqsClient, anlogger, lc)
		if !ok {
			return errors.New(errStr)
		}

		anlogger.Debugf(lc, "create_thumbnails.go : original photo %v was send to create thumbnail", each)
	}

	anlogger.Infof(lc, "create_thumbnails.go : successfully complete ConvertRequest %v", request)
	if len(convertResponse.Objects) >= request.Limit {
		request.Skip = request.Skip + len(convertResponse.Objects)
		anlogger.Infof(lc, "create_thumbnails.go : call myself with new ConvertRequest %v", request)
		err = callMySelf(request.Skip, request.Limit, iamFunctionNameMap[env], lc)
		if err != nil {
			return err
		}
	} else {
		anlogger.Infof(lc, "create_thumbnails.go : successfully complete final ConvertRequest %v", request)
	}
	return nil
}

func callMySelf(skip, limit int, functionName string, lc *lambdacontext.LambdaContext) (error) {

	request := ConvertRequest{
		Skip:  skip,
		Limit: limit,
	}

	jsonBody, err := json.Marshal(request)
	if err != nil {
		anlogger.Errorf(lc, "create_thumbnails.go : error marshaling request %v into json : %v", request, err)
		return fmt.Errorf("error marshaling request %v into json : %v", request, err)
	}

	_, err = clientLambda.Invoke(&lambda.InvokeInput{FunctionName: aws.String(functionName), Payload: jsonBody, InvocationType: aws.String("Event")})
	if err != nil {
		anlogger.Errorf(lc, "create_thumbnails.go : error invoke function [%s] to get convert objects with body %s : %v",
			functionName, jsonBody, err)
		return fmt.Errorf("create_thumbnails.go : error invoke function [%s] to get convert objects with body %s : %v",
			functionName, jsonBody, err)
	}

	return nil
}

func sendRequest(skip, limit int, functionName string, lc *lambdacontext.LambdaContext) (*ConvertResponse, error) {

	request := ConvertRequest{
		Skip:  skip,
		Limit: limit,
	}

	jsonBody, err := json.Marshal(request)
	if err != nil {
		anlogger.Errorf(lc, "create_thumbnails.go : error marshaling request %v into json : %v", request, err)
		return nil, fmt.Errorf("error marshaling request %v into json : %v", request, err)
	}

	resp, err := clientLambda.Invoke(&lambda.InvokeInput{FunctionName: aws.String(functionName), Payload: jsonBody})
	if err != nil {
		anlogger.Errorf(lc, "create_thumbnails.go : error invoke function [%s] to get push objects with body %s : %v",
			functionName, jsonBody, err)
		return nil, fmt.Errorf("create_thumbnails.go : error invoke function [%s] to get push objects with body %s : %v",
			functionName, jsonBody, err)
	}

	if *resp.StatusCode != 200 {
		anlogger.Errorf(lc, "create_thumbnails.go : status code = %d, response body %s for request %s (function name [%s])",
			*resp.StatusCode, string(resp.Payload), jsonBody, functionName)
		return nil, fmt.Errorf("create_thumbnails.go : error invoke function [%s] with body %s : %v",
			functionName, jsonBody, err)
	}

	var response ConvertResponse
	err = json.Unmarshal(resp.Payload, &response)
	if err != nil {
		anlogger.Errorf(lc, "create_thumbnails.go : error unmarshaling response %s into commons.PushResponse : %v",
			string(resp.Payload), err)
		return nil, fmt.Errorf("error unmarshaling response %v into json : %v", string(resp.Payload), err)
	}

	anlogger.Debugf(lc, "create_thumbnails.go : successfully receive [%v] convert objects from [%s]", response, functionName)
	return &response, nil
}

func main() {
	basicLambda.Start(handler)
}

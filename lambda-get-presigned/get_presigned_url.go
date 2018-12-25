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
	"encoding/json"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/aws/aws-sdk-go/service/lambda"
	"crypto/sha1"
	"github.com/satori/go.uuid"
	"github.com/ringoid/commons"
	"github.com/aws/aws-dax-go/dax"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
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

func init() {
	var env string
	var ok bool
	var papertrailAddress string
	var err error
	var awsSession *session.Session

	env, ok = os.LookupEnv("ENV")
	if !ok {
		fmt.Printf("lambda-initialization : get_presigned_url.go : env can not be empty ENV\n")
		os.Exit(1)
	}
	fmt.Printf("lambda-initialization : get_presigned_url.go : start with ENV = [%s]\n", env)

	papertrailAddress, ok = os.LookupEnv("PAPERTRAIL_LOG_ADDRESS")
	if !ok {
		fmt.Printf("lambda-initialization : get_presigned_url.go : env can not be empty PAPERTRAIL_LOG_ADDRESS\n")
		os.Exit(1)
	}
	fmt.Printf("lambda-initialization : get_presigned_url.go : start with PAPERTRAIL_LOG_ADDRESS = [%s]\n", papertrailAddress)

	anlogger, err = commons.New(papertrailAddress, fmt.Sprintf("%s-%s", env, "get-presign-url-image"))
	if err != nil {
		fmt.Errorf("lambda-initialization : get_presigned_url.go : error during startup : %v\n", err)
		os.Exit(1)
	}
	anlogger.Debugf(nil, "lambda-initialization : get_presigned_url.go : logger was successfully initialized")

	internalAuthFunctionName, ok = os.LookupEnv("INTERNAL_AUTH_FUNCTION_NAME")
	if !ok {
		anlogger.Fatalf(nil, "lambda-initialization : get_presigned_url.go : env can not be empty INTERNAL_AUTH_FUNCTION_NAME")
	}
	anlogger.Debugf(nil, "lambda-initialization : get_presigned_url.go : start with INTERNAL_AUTH_FUNCTION_NAME = [%s]", internalAuthFunctionName)

	presignFunctionName, ok = os.LookupEnv("PRESIGN_FUNCTION_NAME")
	if !ok {
		anlogger.Fatalf(nil, "lambda-initialization : get_presigned_url.go : env can not be empty PRESIGN_FUNCTION_NAME")
	}
	anlogger.Debugf(nil, "lambda-initialization : get_presigned_url.go : start with PRESIGN_FUNCTION_NAME = [%s]", presignFunctionName)

	photoUserMappingTableName, ok = os.LookupEnv("PHOTO_USER_MAPPING_TABLE")
	if !ok {
		anlogger.Fatalf(nil, "lambda-initialization : get_presigned_url.go : env can not be empty PHOTO_USER_MAPPING_TABLE")
	}
	anlogger.Debugf(nil, "lambda-initialization : get_presigned_url.go : start with PHOTO_USER_MAPPING_TABLE = [%s]", photoUserMappingTableName)

	originPhotoBucketName, ok = os.LookupEnv("ORIGIN_PHOTO_BUCKET_NAME")
	if !ok {
		anlogger.Fatalf(nil, "lambda-initialization : get_presigned_url.go : env can not be empty ORIGIN_PHOTO_BUCKET_NAME")
	}
	anlogger.Debugf(nil, "lambda-initialization : get_presigned_url.go : start with ORIGIN_PHOTO_BUCKET_NAME = [%s]", originPhotoBucketName)

	awsSession, err = session.NewSession(aws.NewConfig().
		WithRegion(commons.Region).WithMaxRetries(commons.MaxRetries).
		WithLogger(aws.LoggerFunc(func(args ...interface{}) { anlogger.AwsLog(args) })).WithLogLevel(aws.LogOff))
	if err != nil {
		anlogger.Fatalf(nil, "lambda-initialization : get_presigned_url.go : error during initialization : %v", err)
	}
	anlogger.Debugf(nil, "lambda-initialization : get_presigned_url.go : aws session was successfully initialized")

	daxEndpoint, ok := os.LookupEnv("DAX_ENDPOINT")
	if !ok {
		anlogger.Fatalf(nil, "lambda-initialization : get_presigned_url.go : env can not be empty DAX_ENDPOINT")
	}
	cfg := dax.DefaultConfig()
	cfg.HostPorts = []string{daxEndpoint}
	cfg.Region = commons.Region
	daxClient, err = dax.New(cfg)
	if err != nil {
		anlogger.Fatalf(nil, "lambda-initialization : get_presigned_url.go : error initialize DAX cluster")
	}
	anlogger.Debugf(nil, "lambda-initialization : get_presigned_url.go : dax client was successfully initialized")

	clientLambda = lambda.New(awsSession)
	anlogger.Debugf(nil, "lambda-initialization : get_presigned_url.go : lambda client was successfully initialized")

	deliveryStreamName, ok = os.LookupEnv("DELIVERY_STREAM")
	if !ok {
		anlogger.Fatalf(nil, "lambda-initialization : get_presigned_url.go : env can not be empty DELIVERY_STREAM")
	}
	anlogger.Debugf(nil, "lambda-initialization : get_presigned_url.go : start with DELIVERY_STREAM = [%s]", deliveryStreamName)

	awsDeliveryStreamClient = firehose.New(awsSession)
	anlogger.Debugf(nil, "lambda-initialization : get_presigned_url.go : firehose client was successfully initialized")
}

func handler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	lc, _ := lambdacontext.FromContext(ctx)

	anlogger.Debugf(lc, "get_presigned_url.go : start handle request %v", request)

	sourceIp := request.RequestContext.Identity.SourceIP

	if commons.IsItWarmUpRequest(request.Body, anlogger, lc) {
		return events.APIGatewayProxyResponse{}, nil
	}

	appVersion, isItAndroid, ok, errStr := commons.ParseAppVersionFromHeaders(request.Headers, anlogger, lc)
	if !ok {
		anlogger.Errorf(lc, "get_presigned_url.go : return %s to client", errStr)
		return events.APIGatewayProxyResponse{StatusCode: 200, Body: errStr}, nil
	}

	reqParam, ok, errStr := parseParams(request.Body, lc)
	if !ok {
		anlogger.Errorf(lc, "get_presigned_url.go : return %s to client", errStr)
		return events.APIGatewayProxyResponse{StatusCode: 200, Body: errStr}, nil
	}

	userId, ok, _, errStr := commons.CallVerifyAccessToken(appVersion, isItAndroid, reqParam.AccessToken, internalAuthFunctionName, clientLambda, anlogger, lc)
	if !ok {
		anlogger.Errorf(lc, "get_presigned_url.go : return %s to client", errStr)
		return events.APIGatewayProxyResponse{StatusCode: 200, Body: errStr}, nil
	}

	needToRetry := true
	var photoId string
	var s3Key string
	for needToRetry {
		photoId, ok, errStr = generatePhotoId(userId, lc)
		if !ok {
			anlogger.Errorf(lc, "get_presigned_url.go : return %s to client", errStr)
			return events.APIGatewayProxyResponse{StatusCode: 200, Body: errStr}, nil
		}
		s3Key = photoId + "_photo." + reqParam.Extension
		wasCreated, retry, errStr := apimodel.CreatePhotoIdUserIdMappingUpdate(s3Key, userId, photoUserMappingTableName, daxClient, anlogger, lc)
		if !wasCreated && !needToRetry {
			anlogger.Errorf(lc, "get_presigned_url.go : return %s to client", errStr)
			return events.APIGatewayProxyResponse{StatusCode: 200, Body: errStr}, nil
		}
		needToRetry = retry
	}

	uri, ok, errStr := makePresignUrl(userId, originPhotoBucketName, s3Key, presignFunctionName, lc)
	if !ok {
		anlogger.Errorf(lc, "get_presigned_url.go : return %s to client", errStr)
		return events.APIGatewayProxyResponse{StatusCode: 200, Body: errStr}, nil
	}

	event := commons.NewUserAskUploadLinkEvent(originPhotoBucketName, s3Key, userId, sourceIp)
	commons.SendAnalyticEvent(event, userId, deliveryStreamName, awsDeliveryStreamClient, anlogger, lc)

	resp := apimodel.GetPresignUrlResp{
		Uri:           uri,
		OriginPhotoId: "origin_" + photoId,
		ClientPhotoId: reqParam.ClientPhotoId,
	}
	body, err := json.Marshal(resp)
	if err != nil {
		anlogger.Errorf(lc, "get_presigned_url.go : error while marshaling resp [%v] object for userId [%s] : %v", resp, userId, err)
		anlogger.Errorf(lc, "get_presigned_url.go : userId [%s], return %s to client", userId, commons.InternalServerError)
		return events.APIGatewayProxyResponse{StatusCode: 200, Body: commons.InternalServerError}, nil
	}

	anlogger.Infof(lc, "get_presigned_url.go : return presign url for userId [%s]", userId)
	return events.APIGatewayProxyResponse{StatusCode: 200, Body: string(body)}, nil
}

//return generated photoId, was everything ok and error string
func generatePhotoId(userId string, lc *lambdacontext.LambdaContext) (string, bool, string) {
	anlogger.Debugf(lc, "get_presigned_url.go : generate photoId for userId [%s]", userId)
	saltForPhotoId, err := uuid.NewV4()
	if err != nil {
		anlogger.Errorf(lc, "get_presigned_url.go : error while generate salt for photoId, userId [%s] : %v", userId, err)
		return "", false, commons.InternalServerError
	}
	sha := sha1.New()
	_, err = sha.Write([]byte(userId))
	if err != nil {
		anlogger.Errorf(lc, "get_presigned_url.go : error while write userId to sha algo, userId [%s] : %v", userId, err)
		return "", false, commons.InternalServerError
	}
	_, err = sha.Write([]byte(saltForPhotoId.String()))
	if err != nil {
		anlogger.Errorf(lc, "get_presigned_url.go : error while write salt to sha algo, userId [%s] : %v", userId, err)
		return "", false, commons.InternalServerError
	}
	resultPhotoId := fmt.Sprintf("%x", sha.Sum(nil))
	anlogger.Debugf(lc, "get_presigned_url.go : successfully generate photoId [%s] for userId [%s]", resultPhotoId, userId)
	return resultPhotoId, true, ""
}

func parseParams(params string, lc *lambdacontext.LambdaContext) (*apimodel.GetPresignUrlReq, bool, string) {
	anlogger.Debugf(lc, "get_presigned_url.go : parse request body %s", params)
	var req apimodel.GetPresignUrlReq
	err := json.Unmarshal([]byte(params), &req)
	if err != nil {
		anlogger.Errorf(lc, "get_presigned_url.go : error marshaling required params from the string [%s] : %v", params, err)
		return nil, false, commons.InternalServerError
	}

	if req.Extension == "" {
		anlogger.Errorf(lc, "get_presigned_url.go : wrong required param extension [%s]", req.Extension)
		return nil, false, commons.WrongRequestParamsClientError
	}

	if req.ClientPhotoId == "" {
		anlogger.Errorf(lc, "get_presigned_url.go : wrong required param clientPhotoId [%s]", req.ClientPhotoId)
		return nil, false, commons.WrongRequestParamsClientError
	}

	anlogger.Debugf(lc, "get_presigned_url.go : successfully parse request string [%s] to %v", params, req)
	return &req, true, ""
}

//return uri, ok, error string
func makePresignUrl(userId, bucket, key, functionName string, lc *lambdacontext.LambdaContext) (string, bool, string) {
	anlogger.Debugf(lc, "get_presigned_url.go : make pre-signed url for userId [%s], bucket [%s] and key [%s]",
		userId, bucket, key)

	req := apimodel.MakePresignUrlInternalReq{
		Bucket: bucket,
		Key:    key,
	}

	jsonBody, err := json.Marshal(req)
	if err != nil {
		anlogger.Errorf(lc, "get_presigned_url.go : error marshaling req %s into json, for userId [%s] : %v", req, userId, err)
		return "", false, commons.InternalServerError
	}

	resp, err := clientLambda.Invoke(&lambda.InvokeInput{FunctionName: aws.String(functionName), Payload: jsonBody})
	if err != nil {
		anlogger.Errorf(lc, "get_presigned_url.go : error invoke function [%s] with body %s, for userId [%s] : %v",
			functionName, jsonBody, userId, err)
		return "", false, commons.InternalServerError
	}

	if *resp.StatusCode != 200 {
		anlogger.Errorf(lc, "get_presigned_url.go : status code = %d, response body %s for request %s, for userId [%s]",
			*resp.StatusCode, string(resp.Payload), jsonBody, userId)
		return "", false, commons.InternalServerError
	}

	var response apimodel.MakePresignUrlInternalResp
	err = json.Unmarshal(resp.Payload, &response)
	if err != nil {
		anlogger.Errorf(lc, "get_presigned_url.go : error unmarshaling response %s into json, for userId [%s] : %v",
			string(resp.Payload), userId, err)
		return "", false, commons.InternalServerError
	}

	anlogger.Debugf(lc, "get_presigned_url.go : successfully made pre-sign url [%s], for userId [%s]", response.Uri, userId)
	return response.Uri, true, ""
}

func main() {
	basicLambda.Start(handler)
}

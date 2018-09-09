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
	"encoding/json"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/aws/aws-sdk-go/service/lambda"
	"time"
	"strings"
	"github.com/aws/aws-sdk-go/aws/awserr"
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
var userPhotoTable string

func init() {
	var env string
	var ok bool
	var papertrailAddress string
	var err error
	var awsSession *session.Session

	env, ok = os.LookupEnv("ENV")
	if !ok {
		fmt.Printf("delete_photo.go : env can not be empty ENV")
		os.Exit(1)
	}
	fmt.Printf("delete_photo.go : start with ENV = [%s]", env)

	papertrailAddress, ok = os.LookupEnv("PAPERTRAIL_LOG_ADDRESS")
	if !ok {
		fmt.Printf("delete_photo.go : env can not be empty PAPERTRAIL_LOG_ADDRESS")
		os.Exit(1)
	}
	fmt.Printf("delete_photo.go : start with PAPERTRAIL_LOG_ADDRESS = [%s]", papertrailAddress)

	anlogger, err = syslog.New(papertrailAddress, fmt.Sprintf("%s-%s", env, "create-auth"))
	if err != nil {
		fmt.Errorf("delete_photo.go : error during startup : %v", err)
		os.Exit(1)
	}
	anlogger.Debugf(nil, "delete_photo.go : logger was successfully initialized")

	internalAuthFunctionName, ok = os.LookupEnv("INTERNAL_AUTH_FUNCTION_NAME")
	if !ok {
		fmt.Printf("delete_photo.go : env can not be empty INTERNAL_AUTH_FUNCTION_NAME")
		os.Exit(1)
	}
	anlogger.Debugf(nil, "delete_photo.go : start with INTERNAL_AUTH_FUNCTION_NAME = [%s]", internalAuthFunctionName)

	presignFunctionName, ok = os.LookupEnv("PRESIGN_FUNCTION_NAME")
	if !ok {
		fmt.Printf("delete_photo.go : env can not be empty PRESIGN_FUNCTION_NAME")
		os.Exit(1)
	}
	anlogger.Debugf(nil, "delete_photo.go : start with PRESIGN_FUNCTION_NAME = [%s]", presignFunctionName)

	photoUserMappingTableName, ok = os.LookupEnv("PHOTO_USER_MAPPING_TABLE")
	if !ok {
		fmt.Printf("delete_photo.go : env can not be empty PHOTO_USER_MAPPING_TABLE")
		os.Exit(1)
	}
	anlogger.Debugf(nil, "delete_photo.go : start with PHOTO_USER_MAPPING_TABLE = [%s]", photoUserMappingTableName)

	originPhotoBucketName, ok = os.LookupEnv("ORIGIN_PHOTO_BUCKET_NAME")
	if !ok {
		fmt.Printf("delete_photo.go : env can not be empty ORIGIN_PHOTO_BUCKET_NAME")
		os.Exit(1)
	}
	anlogger.Debugf(nil, "delete_photo.go : start with ORIGIN_PHOTO_BUCKET_NAME = [%s]", originPhotoBucketName)

	userPhotoTable, ok = os.LookupEnv("USER_PHOTO_TABLE")
	if !ok {
		fmt.Printf("delete_photo.go : env can not be empty USER_PHOTO_TABLE")
		os.Exit(1)
	}
	anlogger.Debugf(nil, "delete_photo.go : start with USER_PHOTO_TABLE = [%s]", userPhotoTable)

	awsSession, err = session.NewSession(aws.NewConfig().
		WithRegion(apimodel.Region).WithMaxRetries(apimodel.MaxRetries).
		WithLogger(aws.LoggerFunc(func(args ...interface{}) { anlogger.AwsLog(args) })).WithLogLevel(aws.LogOff))
	if err != nil {
		anlogger.Fatalf(nil, "delete_photo.go : error during initialization : %v", err)
	}
	anlogger.Debugf(nil, "delete_photo.go : aws session was successfully initialized")

	awsDbClient = dynamodb.New(awsSession)
	anlogger.Debugf(nil, "delete_photo.go : dynamodb client was successfully initialized")

	clientLambda = lambda.New(awsSession)
	anlogger.Debugf(nil, "delete_photo.go : lambda client was successfully initialized")

	deliveryStreamName, ok = os.LookupEnv("DELIVERY_STREAM")
	if !ok {
		anlogger.Fatalf(nil, "delete_photo.go : env can not be empty DELIVERY_STREAM")
		os.Exit(1)
	}
	anlogger.Debugf(nil, "delete_photo.go : start with DELIVERY_STREAM = [%s]", deliveryStreamName)

	awsDeliveryStreamClient = firehose.New(awsSession)
	anlogger.Debugf(nil, "delete_photo.go : firehose client was successfully initialized")
}

func handler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	lc, _ := lambdacontext.FromContext(ctx)

	anlogger.Debugf(lc, "delete_photo.go : start handle request %v", request)

	reqParam, ok, errStr := parseParams(request.Body, lc)
	if !ok {
		anlogger.Errorf(lc, "delete_photo.go : return %s to client", errStr)
		return events.APIGatewayProxyResponse{StatusCode: 200, Body: errStr}, nil
	}

	userId, ok, errStr := apimodel.CallVerifyAccessToken(reqParam.AccessToken, internalAuthFunctionName, clientLambda, anlogger, lc)
	if !ok {
		anlogger.Errorf(lc, "delete_photo.go : return %s to client", errStr)
		return events.APIGatewayProxyResponse{StatusCode: 200, Body: errStr}, nil
	}

	photoIds, originPhotoId := getAllPhotoIds(reqParam.PhotoId, userId, lc)
	for _, val := range photoIds {
		ok, errStr := markAsDel(userId, val, lc)
		if !ok {
			anlogger.Errorf(lc, "delete_photo.go : userId [%s], return %s to client", userId, errStr)
			return events.APIGatewayProxyResponse{StatusCode: 200, Body: errStr}, nil
		}
		ok, errStr = asyncDel(userId, val, lc)
		if !ok {
			anlogger.Errorf(lc, "delete_photo.go : userId [%s], return %s to client", userId, errStr)
			return events.APIGatewayProxyResponse{StatusCode: 200, Body: errStr}, nil
		}
	}

	event := apimodel.NewUserDeletePhotoEvent(userId, originPhotoId)
	apimodel.SendAnalyticEvent(event, userId, deliveryStreamName, awsDeliveryStreamClient, anlogger, lc)

	resp := apimodel.BaseResponse{}
	body, err := json.Marshal(resp)
	if err != nil {
		anlogger.Errorf(lc, "delete_photo.go : error while marshaling resp [%v] object for userId [%s] : %v", resp, userId, err)
		anlogger.Errorf(lc, "delete_photo.go : userId [%s], return %s to client", userId, apimodel.InternalServerError)
		return events.APIGatewayProxyResponse{StatusCode: 200, Body: apimodel.InternalServerError}, nil
	}
	anlogger.Debugf(lc, "delete_photo.go : return successful resp [%s] for userId [%s]", string(body), userId)
	anlogger.Infof(lc, "delete_photo.go : successfully delete all photo based on photoId [%s] for userId [%s]", reqParam.PhotoId, userId)
	return events.APIGatewayProxyResponse{StatusCode: 200, Body: string(body)}, nil
}

func getAllPhotoIds(sourceId, userId string, lc *lambdacontext.LambdaContext) ([]string, string) {
	anlogger.Debugf(lc, "delete_photo.go : make del photo id list based on photoId [%s] for userId [%s]", sourceId, userId)
	arr := strings.Split(sourceId, "_")
	baseId := arr[1]
	allIds := make([]string, 0)
	originPhotoId := "origin_" + baseId
	allIds = append(allIds, originPhotoId)
	for key, _ := range apimodel.AllowedPhotoResolution {
		allIds = append(allIds, key+"_"+baseId)
	}
	anlogger.Debugf(lc, "delete_photo.go : successfully cretae del photo id list based on photoId [%s] for userId [%s], del list=%v", sourceId, userId, allIds)
	return allIds, originPhotoId
}

//return ok and error string
func asyncDel(userId, photoId string, lc *lambdacontext.LambdaContext) (bool, string) {
	//todo:implement
	return true, ""
}

//return ok and error string
func markAsDel(userId, photoId string, lc *lambdacontext.LambdaContext) (bool, string) {
	anlogger.Debugf(lc, "delete_photo.go : mark photoId [%s] as deleted for userId [%s]", photoId, userId)
	input :=
		&dynamodb.UpdateItemInput{
			ExpressionAttributeNames: map[string]*string{
				"#deletedAt": aws.String(apimodel.PhotoDeletedAt),
			},
			ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
				":deletedAtV": {
					S: aws.String(time.Now().UTC().Format("2006-01-02-15-04-05.000")),
				},
			},
			Key: map[string]*dynamodb.AttributeValue{
				apimodel.UserIdColumnName: {
					S: aws.String(userId),
				},
				apimodel.PhotoIdColumnName: {
					S: aws.String(photoId),
				},
			},
			TableName:           aws.String(userPhotoTable),
			ConditionExpression: aws.String(fmt.Sprintf("attribute_exists(%s) AND attribute_exists(%s)", apimodel.UserIdColumnName, apimodel.PhotoIdColumnName)),
			UpdateExpression:    aws.String("SET #deletedAt = :deletedAtV"),
		}
	_, err := awsDbClient.UpdateItem(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case dynamodb.ErrCodeConditionalCheckFailedException:
				anlogger.Debugf(lc, "delete_photo.go : photoId [%s] not exist for userId [%s]", photoId, userId)
				return true, ""
			}
		}
		anlogger.Errorf(lc, "delete_photo.go : error while delete photoId [%s] for userId [%s] : %v", photoId, userId, err)
		return false, apimodel.InternalServerError
	}
	anlogger.Debugf(lc, "delete_photo.go : successfully delete photoId [%s] for userId [%s]", photoId, userId)
	return true, ""
}

func parseParams(params string, lc *lambdacontext.LambdaContext) (*apimodel.DeletePhotoReq, bool, string) {
	anlogger.Debugf(lc, "delete_photo.go : parse request body %s", params)
	var req apimodel.DeletePhotoReq
	err := json.Unmarshal([]byte(params), &req)
	if err != nil {
		anlogger.Errorf(lc, "delete_photo.go : error marshaling required params from the string [%s] : %v", params, err)
		return nil, false, apimodel.InternalServerError
	}

	if req.PhotoId == "" {
		anlogger.Errorf(lc, "delete_photo.go : wrong required param photoId [%s]", req.PhotoId)
		return nil, false, apimodel.WrongRequestParamsClientError
	}

	anlogger.Debugf(lc, "delete_photo.go : successfully parse request string [%s] to %v", params, req)
	return &req, true, ""
}

func main() {
	basicLambda.Start(handler)
}

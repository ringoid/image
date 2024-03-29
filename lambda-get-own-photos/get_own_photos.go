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
	"encoding/json"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/aws/aws-sdk-go/service/lambda"
	"sort"
	"strings"
	"strconv"
	"github.com/ringoid/commons"
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
var userPhotoTable string
var env string

var userIdStatusEnabledMap map[string]bool

func init() {
	var ok bool
	var papertrailAddress string
	var err error
	var awsSession *session.Session

	env, ok = os.LookupEnv("ENV")
	if !ok {
		fmt.Printf("lambda-initialization : get_own_photos.go : env can not be empty ENV\n")
		os.Exit(1)
	}
	fmt.Printf("lambda-initialization : get_own_photos.go : start with ENV = [%s]\n", env)

	papertrailAddress, ok = os.LookupEnv("PAPERTRAIL_LOG_ADDRESS")
	if !ok {
		fmt.Printf("lambda-initialization : get_own_photos.go : env can not be empty PAPERTRAIL_LOG_ADDRESS\n")
		os.Exit(1)
	}
	fmt.Printf("lambda-initialization : get_own_photos.go : start with PAPERTRAIL_LOG_ADDRESS = [%s]\n", papertrailAddress)

	anlogger, err = commons.New(papertrailAddress, fmt.Sprintf("%s-%s", env, "get-own-photos-image"), apimodel.IsDebugLogEnabled)
	if err != nil {
		fmt.Errorf("lambda-initialization : get_own_photos.go : error during startup : %v\n", err)
		os.Exit(1)
	}
	anlogger.Debugf(nil, "lambda-initialization : get_own_photos.go : logger was successfully initialized")

	internalAuthFunctionName, ok = os.LookupEnv("INTERNAL_AUTH_FUNCTION_NAME")
	if !ok {
		anlogger.Fatalf(nil, "lambda-initialization : get_own_photos.go : env can not be empty INTERNAL_AUTH_FUNCTION_NAME")
	}
	anlogger.Debugf(nil, "lambda-initialization : get_own_photos.go : start with INTERNAL_AUTH_FUNCTION_NAME = [%s]", internalAuthFunctionName)

	presignFunctionName, ok = os.LookupEnv("PRESIGN_FUNCTION_NAME")
	if !ok {
		anlogger.Fatalf(nil, "lambda-initialization : get_own_photos.go : env can not be empty PRESIGN_FUNCTION_NAME")
	}
	anlogger.Debugf(nil, "lambda-initialization : get_own_photos.go : start with PRESIGN_FUNCTION_NAME = [%s]", presignFunctionName)

	photoUserMappingTableName, ok = os.LookupEnv("PHOTO_USER_MAPPING_TABLE")
	if !ok {
		anlogger.Fatalf(nil, "lambda-initialization : get_own_photos.go : env can not be empty PHOTO_USER_MAPPING_TABLE")
	}
	anlogger.Debugf(nil, "lambda-initialization : get_own_photos.go : start with PHOTO_USER_MAPPING_TABLE = [%s]", photoUserMappingTableName)

	originPhotoBucketName, ok = os.LookupEnv("ORIGIN_PHOTO_BUCKET_NAME")
	if !ok {
		anlogger.Fatalf(nil, "lambda-initialization : get_own_photos.go : env can not be empty ORIGIN_PHOTO_BUCKET_NAME")
	}
	anlogger.Debugf(nil, "lambda-initialization : get_own_photos.go : start with ORIGIN_PHOTO_BUCKET_NAME = [%s]", originPhotoBucketName)

	userPhotoTable, ok = os.LookupEnv("USER_PHOTO_TABLE")
	if !ok {
		anlogger.Fatalf(nil, "lambda-initialization : get_own_photos.go : env can not be empty USER_PHOTO_TABLE")
	}
	anlogger.Debugf(nil, "lambda-initialization : get_own_photos.go : start with USER_PHOTO_TABLE = [%s]", userPhotoTable)

	awsSession, err = session.NewSession(aws.NewConfig().
		WithRegion(commons.Region).WithMaxRetries(commons.MaxRetries).
		WithLogger(aws.LoggerFunc(func(args ...interface{}) { anlogger.AwsLog(args) })).WithLogLevel(aws.LogOff))
	if err != nil {
		anlogger.Fatalf(nil, "lambda-initialization : get_own_photos.go : error during initialization : %v", err)
	}
	anlogger.Debugf(nil, "lambda-initialization : get_own_photos.go : aws session was successfully initialized")

	awsDbClient = dynamodb.New(awsSession)
	anlogger.Debugf(nil, "lambda-initialization : get_own_photos.go : dynamodb client was successfully initialized")

	clientLambda = lambda.New(awsSession)
	anlogger.Debugf(nil, "lambda-initialization : get_own_photos.go : lambda client was successfully initialized")

	deliveryStreamName, ok = os.LookupEnv("DELIVERY_STREAM")
	if !ok {
		anlogger.Fatalf(nil, "lambda-initialization : get_own_photos.go : env can not be empty DELIVERY_STREAM")
	}
	anlogger.Debugf(nil, "lambda-initialization : get_own_photos.go : start with DELIVERY_STREAM = [%s]", deliveryStreamName)

	awsDeliveryStreamClient = firehose.New(awsSession)
	anlogger.Debugf(nil, "lambda-initialization : get_own_photos.go : firehose client was successfully initialized")

	userIdStatusEnabledMap = make(map[string]bool)
	//Kirill
	userIdStatusEnabledMap["d0b285a7d39f07e528dfba085e07a6135ddde188"] = true
	userIdStatusEnabledMap["ea6fa85e8afcf574d50c59b1d6cd1f2217fb718c"] = true
	userIdStatusEnabledMap["b9094fec646aa6296d0d3b3238801f92af34083a"] = true
	userIdStatusEnabledMap["54047644372b264ee02a1ac4e47cc6d02fc517bd"] = true
	//Iam
	userIdStatusEnabledMap["f2c1e8abc72645c23cf2f89c8f0e7cb4fd7d9adc"] = true
	//Victor
	userIdStatusEnabledMap["c86a29c241f8a0dadf3cff31b4c831bbfe3f2633"] = true
	//Maxim
	userIdStatusEnabledMap["f966276704b50ec1d472e34bbd184d89082bcdfb"] = true

}

func handler(ctx context.Context, request events.ALBTargetGroupRequest) (events.ALBTargetGroupResponse, error) {
	lc, _ := lambdacontext.FromContext(ctx)

	userAgent := request.Headers["user-agent"]
	if strings.HasPrefix(userAgent, "ELB-HealthChecker") {
		return commons.NewServiceResponse("{}"), nil
	}

	if request.HTTPMethod != "GET" {
		return commons.NewWrongHttpMethodServiceResponse(), nil
	}
	sourceIp := request.Headers["x-forwarded-for"]

	anlogger.Debugf(lc, "get_own_photos.go : start handle request %v", request)

	appVersion, isItAndroid, ok, errStr := commons.ParseAppVersionFromHeaders(request.Headers, anlogger, lc)
	if !ok {
		anlogger.Errorf(lc, "get_own_photos.go : return %s to client", errStr)
		return commons.NewServiceResponse(errStr), nil
	}

	accessToken, okA := request.QueryStringParameters["accessToken"]
	resolution, okR := request.QueryStringParameters["resolution"]

	if !okA || !okR {
		errStr := commons.WrongRequestParamsClientError
		anlogger.Errorf(lc, "get_own_photos.go : one or both of required params (accessToken || resolution) is empty", errStr)
		return commons.NewServiceResponse(errStr), nil
	}

	if !commons.AllowedPhotoResolution[resolution] {
		anlogger.Warnf(lc, "get_own_photos.go : resolution [%s] is not supported, so use [%s] resolution", resolution, commons.BiggestDefaultPhotoResolution)
		resolution = commons.BiggestDefaultPhotoResolution
	}

	userId, ok, _, errStr := commons.CallVerifyAccessToken(appVersion, isItAndroid, accessToken, internalAuthFunctionName, clientLambda, anlogger, lc)
	if !ok {
		anlogger.Errorf(lc, "get_own_photos.go : return %s to client", errStr)
		return commons.NewServiceResponse(errStr), nil
	}

	photos, ok, errStr := getOwnPhotos(userId, resolution, lc)
	if !ok {
		anlogger.Errorf(lc, "get_own_photos.go : userId [%s], return %s to client", userId, errStr)
		return commons.NewServiceResponse(errStr), nil
	}

	metaMap, ok, errStr := getMetaInfs(userId, lc)
	if !ok {
		anlogger.Errorf(lc, "get_own_photos.go : userId [%s], return %s to client", userId, errStr)
		return commons.NewServiceResponse(errStr), nil
	}

	photos = fillMetaPhotoInf(photos, metaMap)
	photos = sortOwnPhotos(photos)

	resp := apimodel.GetOwnPhotosResp{}
	ownPhotos := make([]apimodel.OwnPhoto, 0)
	for _, value := range photos {
		eachPhoto := apimodel.OwnPhoto{
			PhotoId:       value.PhotoId,
			PhotoUri:      value.PhotoSourceUri,
			Likes:         value.Likes,
			OriginPhotoId: value.OriginPhotoId,
			Blocked:       false,
		}
		//check moderation
		if value.HiddenInModeration {
			eachPhoto.Blocked = true
			eachPhoto.PhotoUri = ""
		}

		ownPhotos = append(ownPhotos, eachPhoto)
	}
	resp.Photos = ownPhotos

	event := commons.NewGetOwnPhotosEvent(userId, sourceIp, len(resp.Photos))
	commons.SendAnalyticEvent(event, userId, deliveryStreamName, awsDeliveryStreamClient, anlogger, lc)

	resp.LastOnlineText = "Online"
	resp.LastOnlineFlag = "online"
	resp.DistanceText = "unknown"

	body, err := json.Marshal(resp)
	if err != nil {
		anlogger.Errorf(lc, "get_own_photos.go : error while marshaling resp [%v] object for userId [%s] : %v", resp, userId, err)
		anlogger.Errorf(lc, "get_own_photos.go : userId [%s], return %s to client", userId, commons.InternalServerError)
		return commons.NewServiceResponse(commons.InternalServerError), nil
	}
	anlogger.Debugf(lc, "get_own_photos.go : return successful resp [%s] for userId [%s]", string(body), userId)

	anlogger.Infof(lc, "get_own_photos.go : return [%d] own photos to the user with userId [%s]", len(resp.Photos), userId)
	return commons.NewServiceResponse(string(body)), nil
}

func fillMetaPhotoInf(source []*apimodel.UserPhoto, metaMap map[string]*apimodel.UserPhotoMetaInf) []*apimodel.UserPhoto {
	for _, val := range source {
		photoId := val.OriginPhotoId
		if meta, ok := metaMap[photoId]; ok {
			val.Likes = meta.Likes
		}
	}
	return source
}

func sortOwnPhotos(source []*apimodel.UserPhoto) []*apimodel.UserPhoto {
	sort.SliceStable(source, func(i, j int) bool {
		return source[i].UpdatedAt > source[j].UpdatedAt
	})
	return source
}

//return own photos, ok and error string
//todo:keep in mind that we should use ExclusiveStartKey later, if somebody will have > 100K photos
func getOwnPhotos(userId, resolution string, lc *lambdacontext.LambdaContext) ([]*apimodel.UserPhoto, bool, string) {
	anlogger.Debugf(lc, "get_own_photos.go : get all own photos for userId [%s] with resolution [%s]", userId, resolution)
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
		FilterExpression:       aws.String(fmt.Sprintf("attribute_not_exists(%s) OR attribute_exists(%s)", commons.PhotoDeletedAtColumnName, commons.PhotoHiddenAtColumnName)),
		ConsistentRead:         aws.Bool(true),
		KeyConditionExpression: aws.String("#userId = :userIdV AND begins_with(#photoId, :photoIdV)"),
		TableName:              aws.String(userPhotoTable),
	}
	result, err := awsDbClient.Query(input)
	if err != nil {
		anlogger.Errorf(lc, "get_own_photos.go : error while query all own photos userId [%s] with resolution [%s] : %v", userId, resolution, err)
		return make([]*apimodel.UserPhoto, 0), false, commons.InternalServerError
	}

	if *result.Count == 0 {
		anlogger.Debugf(lc, "get_own_photos.go : there is no photo for userId [%s] with resolution [%s]", userId, resolution)
		return make([]*apimodel.UserPhoto, 0), true, ""
	}

	items := make([]*apimodel.UserPhoto, 0)
	for _, v := range result.Items {
		originPhotoId := strings.Replace(*v[commons.PhotoIdColumnName].S, resolution, "origin", 1)
		hiddenInModeration := false
		if val, ok := v[commons.PhotoHiddenAtColumnName]; ok {
			if val.S != nil {
				hiddenInModeration = true
			}
		}
		items = append(items, &apimodel.UserPhoto{
			UserId:             *v[commons.UserIdColumnName].S,
			PhotoId:            *v[commons.PhotoIdColumnName].S,
			PhotoSourceUri:     *v[commons.PhotoSourceUriColumnName].S,
			PhotoType:          *v[commons.PhotoTypeColumnName].S,
			Bucket:             *v[commons.PhotoBucketColumnName].S,
			Key:                *v[commons.PhotoKeyColumnName].S,
			UpdatedAt:          *v[commons.UpdatedTimeColumnName].S,
			OriginPhotoId:      originPhotoId,
			HiddenInModeration: hiddenInModeration,
		})
	}
	anlogger.Debugf(lc, "get_own_photos.go : successfully fetch [%v] photos for userId [%s] and resolution [%s], result=%v",
		*result.Count, userId, resolution, items)
	return items, true, ""
}

//return photo's meta infs, ok and error string
//todo:keep in mind that we should use ExclusiveStartKey later, if somebody will have > 100K photos
func getMetaInfs(userId string, lc *lambdacontext.LambdaContext) (map[string]*apimodel.UserPhotoMetaInf, bool, string) {
	anlogger.Debugf(lc, "get_own_photos.go : get all photo's meta infs for userId [%s]", userId)
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
		FilterExpression:       aws.String(fmt.Sprintf("attribute_not_exists(%s) OR attribute_exists(%s)", commons.PhotoDeletedAtColumnName, commons.PhotoHiddenAtColumnName)),
		ConsistentRead:         aws.Bool(true),
		KeyConditionExpression: aws.String("#userId = :userIdV"),
		TableName:              aws.String(userPhotoTable),
	}
	result, err := awsDbClient.Query(input)
	if err != nil {
		anlogger.Errorf(lc, "get_own_photos.go : error while query all photo's meta infs for userId [%s] : %v", userId, err)
		return make(map[string]*apimodel.UserPhotoMetaInf, 0), false, commons.InternalServerError
	}

	if *result.Count == 0 {
		anlogger.Debugf(lc, "get_own_photos.go : there is no photo's meta info for userId [%s]", userId)
		return make(map[string]*apimodel.UserPhotoMetaInf, 0), true, ""
	}

	anlogger.Debugf(lc, "get_own_photos.go : there is [%d] photo's meta info for userId [%s]", *result.Count, userId)

	items := make(map[string]*apimodel.UserPhotoMetaInf, 0)
	for _, v := range result.Items {
		photoId := *v[commons.PhotoIdColumnName].S
		likes := 0
		if _, likeExist := v[commons.PhotoLikesColumnName]; likeExist {
			likes, err = strconv.Atoi(*v[commons.PhotoLikesColumnName].N)
			if err != nil {
				anlogger.Errorf(lc, "get_own_photos.go : error while convert likes from photo meta inf to int, photoId [%s] for userId [%s] : %v", photoId, userId, err)
				return make(map[string]*apimodel.UserPhotoMetaInf, 0), false, commons.InternalServerError
			}
		}
		items[photoId] = &apimodel.UserPhotoMetaInf{
			UserId:        *v[commons.UserIdColumnName].S,
			OriginPhotoId: photoId,
			Likes:         likes,
		}
	}

	anlogger.Debugf(lc, "get_own_photos.go : successfully fetch [%v] photo meta inf for userId [%s], result=%v",
		*result.Count, userId, items)

	return items, true, ""
}

func main() {
	basicLambda.Start(handler)
}

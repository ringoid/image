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
	"sort"
	"strings"
	"strconv"
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

	anlogger, err = syslog.New(papertrailAddress, fmt.Sprintf("%s-%s", env, "get-own-photos-image"))
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
		fmt.Printf("lambda-initialization : get_own_photos.go : env can not be empty PRESIGN_FUNCTION_NAME")
		os.Exit(1)
	}
	anlogger.Debugf(nil, "lambda-initialization : get_own_photos.go : start with PRESIGN_FUNCTION_NAME = [%s]", presignFunctionName)

	photoUserMappingTableName, ok = os.LookupEnv("PHOTO_USER_MAPPING_TABLE")
	if !ok {
		fmt.Printf("lambda-initialization : get_own_photos.go : env can not be empty PHOTO_USER_MAPPING_TABLE")
		os.Exit(1)
	}
	anlogger.Debugf(nil, "lambda-initialization : get_own_photos.go : start with PHOTO_USER_MAPPING_TABLE = [%s]", photoUserMappingTableName)

	originPhotoBucketName, ok = os.LookupEnv("ORIGIN_PHOTO_BUCKET_NAME")
	if !ok {
		fmt.Printf("lambda-initialization : get_own_photos.go : env can not be empty ORIGIN_PHOTO_BUCKET_NAME")
		os.Exit(1)
	}
	anlogger.Debugf(nil, "lambda-initialization : get_own_photos.go : start with ORIGIN_PHOTO_BUCKET_NAME = [%s]", originPhotoBucketName)

	userPhotoTable, ok = os.LookupEnv("USER_PHOTO_TABLE")
	if !ok {
		fmt.Printf("lambda-initialization : get_own_photos.go : env can not be empty USER_PHOTO_TABLE")
		os.Exit(1)
	}
	anlogger.Debugf(nil, "lambda-initialization : get_own_photos.go : start with USER_PHOTO_TABLE = [%s]", userPhotoTable)

	awsSession, err = session.NewSession(aws.NewConfig().
		WithRegion(apimodel.Region).WithMaxRetries(apimodel.MaxRetries).
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
		os.Exit(1)
	}
	anlogger.Debugf(nil, "lambda-initialization : get_own_photos.go : start with DELIVERY_STREAM = [%s]", deliveryStreamName)

	awsDeliveryStreamClient = firehose.New(awsSession)
	anlogger.Debugf(nil, "lambda-initialization : get_own_photos.go : firehose client was successfully initialized")
}

func handler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	lc, _ := lambdacontext.FromContext(ctx)

	anlogger.Debugf(lc, "get_own_photos.go : start handle request %v", request)

	if apimodel.IsItWarmUpRequest(request.Body, anlogger, lc) {
		return events.APIGatewayProxyResponse{}, nil
	}

	appVersion, isItAndroid, ok, errStr := apimodel.ParseAppVersionFromHeaders(request.Headers, anlogger, lc)
	if !ok {
		anlogger.Errorf(lc, "delete_photo.go : return %s to client", errStr)
		return events.APIGatewayProxyResponse{StatusCode: 200, Body: errStr}, nil
	}

	accessToken := request.QueryStringParameters["accessToken"]
	resolution := request.QueryStringParameters["resolution"]

	if !apimodel.AllowedPhotoResolution[resolution] {
		errStr := apimodel.WrongRequestParamsClientError
		anlogger.Errorf(lc, "get_own_photos : resolution [%s] is not supported", resolution)
		anlogger.Errorf(lc, "get_own_photos.go : return %s to client", errStr)
		return events.APIGatewayProxyResponse{StatusCode: 200, Body: errStr}, nil
	}

	userId, ok, _, errStr := apimodel.CallVerifyAccessToken(appVersion, isItAndroid, accessToken, internalAuthFunctionName, clientLambda, anlogger, lc)
	if !ok {
		anlogger.Errorf(lc, "get_own_photos.go : return %s to client", errStr)
		return events.APIGatewayProxyResponse{StatusCode: 200, Body: errStr}, nil
	}

	photos, ok, errStr := getOwnPhotos(userId, resolution, lc)
	if !ok {
		anlogger.Errorf(lc, "get_own_photos.go : return %s to client", errStr)
		return events.APIGatewayProxyResponse{StatusCode: 200, Body: errStr}, nil
	}

	metaMap, ok, errStr := getMetaInfs(userId, lc)
	if !ok {
		anlogger.Errorf(lc, "get_own_photos.go : return %s to client", errStr)
		return events.APIGatewayProxyResponse{StatusCode: 200, Body: errStr}, nil
	}

	photos = fillMetaPhotoInf(photos, metaMap)
	photos = sortOwnPhotos(photos)

	resp := apimodel.GetOwnPhotosResp{}
	ownPhotos := make([]apimodel.OwnPhoto, 0)
	for _, value := range photos {
		ownPhotos = append(ownPhotos, apimodel.OwnPhoto{
			PhotoId:       value.PhotoId,
			PhotoUri:      value.PhotoSourceUri,
			Likes:         value.Likes,
			OriginPhotoId: value.OriginPhotoId,
		})
	}
	resp.Photos = ownPhotos

	body, err := json.Marshal(resp)
	if err != nil {
		anlogger.Errorf(lc, "get_own_photos.go : error while marshaling resp [%v] object for userId [%s] : %v", resp, userId, err)
		anlogger.Errorf(lc, "get_own_photos.go : userId [%s], return %s to client", userId, apimodel.InternalServerError)
		return events.APIGatewayProxyResponse{StatusCode: 200, Body: apimodel.InternalServerError}, nil
	}
	anlogger.Debugf(lc, "get_own_photos.go : return successful resp [%s] for userId [%s]", string(body), userId)
	return events.APIGatewayProxyResponse{StatusCode: 200, Body: string(body)}, nil
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
		if source[i].Likes == source[j].Likes {
			return source[i].UpdatedAt > source[j].UpdatedAt
		}
		return source[i].Likes > source[j].Likes
	})
	return source
}

//return own photos, ok and error string
//todo:keep in mind that we should use ExclusiveStartKey later, if somebody will have > 100K photos
func getOwnPhotos(userId, resolution string, lc *lambdacontext.LambdaContext) ([]*apimodel.UserPhoto, bool, string) {
	anlogger.Debugf(lc, "get_own_photos.go : get all own photos for userId [%s] with resolution [%s]", userId, resolution)
	input := &dynamodb.QueryInput{
		ExpressionAttributeNames: map[string]*string{
			"#userId":  aws.String(apimodel.UserIdColumnName),
			"#photoId": aws.String(apimodel.PhotoIdColumnName),
		},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":userIdV": {
				S: aws.String(userId),
			},
			":photoIdV": {
				S: aws.String(resolution),
			},
		},
		FilterExpression:       aws.String(fmt.Sprintf("attribute_not_exists(%s)", apimodel.PhotoDeletedAtColumnName)),
		ConsistentRead:         aws.Bool(true),
		KeyConditionExpression: aws.String("#userId = :userIdV AND begins_with(#photoId, :photoIdV)"),
		TableName:              aws.String(userPhotoTable),
	}
	result, err := awsDbClient.Query(input)
	if err != nil {
		anlogger.Errorf(lc, "get_own_photos.go : error while query all own photos userId [%s] with resolution [%s] : %v", userId, resolution, err)
		return make([]*apimodel.UserPhoto, 0), false, apimodel.InternalServerError
	}

	if *result.Count == 0 {
		anlogger.Debugf(lc, "get_own_photos.go : there is no photo for userId [%s] with resolution [%s]", userId, resolution)
		return make([]*apimodel.UserPhoto, 0), true, ""
	}

	items := make([]*apimodel.UserPhoto, 0)
	for _, v := range result.Items {
		originPhotoId := strings.Replace(*v[apimodel.PhotoIdColumnName].S, resolution, "origin", 1)
		items = append(items, &apimodel.UserPhoto{
			UserId:         *v[apimodel.UserIdColumnName].S,
			PhotoId:        *v[apimodel.PhotoIdColumnName].S,
			PhotoSourceUri: *v[apimodel.PhotoSourceUriColumnName].S,
			PhotoType:      *v[apimodel.PhotoTypeColumnName].S,
			Bucket:         *v[apimodel.PhotoBucketColumnName].S,
			Key:            *v[apimodel.PhotoKeyColumnName].S,
			UpdatedAt:      *v[apimodel.UpdatedTimeColumnName].S,
			OriginPhotoId:  originPhotoId,
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
	metaInfPartitionKey := userId + apimodel.PhotoPrimaryKeyMetaPostfix
	input := &dynamodb.QueryInput{
		ExpressionAttributeNames: map[string]*string{
			"#userId": aws.String(apimodel.UserIdColumnName),
		},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":userIdV": {
				S: aws.String(metaInfPartitionKey),
			},
		},
		ConsistentRead:         aws.Bool(true),
		KeyConditionExpression: aws.String("#userId = :userIdV"),
		TableName:              aws.String(userPhotoTable),
	}
	result, err := awsDbClient.Query(input)
	if err != nil {
		anlogger.Errorf(lc, "get_own_photos.go : error while query all photo's meta infs for userId [%s] : %v", userId, err)
		return make(map[string]*apimodel.UserPhotoMetaInf, 0), false, apimodel.InternalServerError
	}

	if *result.Count == 0 {
		anlogger.Debugf(lc, "get_own_photos.go : there is no photo's meta info for userId [%s]", userId)
		return make(map[string]*apimodel.UserPhotoMetaInf, 0), true, ""
	}

	items := make(map[string]*apimodel.UserPhotoMetaInf, 0)
	for _, v := range result.Items {
		photoId := *v[apimodel.PhotoIdColumnName].S

		likes, err := strconv.Atoi(*v[apimodel.PhotoLikesColumnName].N)
		if err != nil {
			anlogger.Errorf(lc, "get_own_photos.go : error while convert likes from photo meta inf to int, photoId [%s] for userId [%s] : %v", photoId, userId, err)
			return make(map[string]*apimodel.UserPhotoMetaInf, 0), false, apimodel.InternalServerError
		}

		items[photoId] = &apimodel.UserPhotoMetaInf{
			UserId:        *v[apimodel.UserIdColumnName].S,
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

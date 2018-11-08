package main

import (
	"context"
	basicLambda "github.com/aws/aws-lambda-go/lambda"
	"../sys_log"
	"../apimodel"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/aws"
	"os"
	"fmt"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-lambda-go/lambdacontext"
)

var anlogger *syslog.Logger
var awsDbClient *dynamodb.DynamoDB
var userPhotoTable string

const defaultBatchSize = 100 //100 is a max

func init() {
	var env string
	var ok bool
	var papertrailAddress string
	var err error
	var awsSession *session.Session

	env, ok = os.LookupEnv("ENV")
	if !ok {
		fmt.Printf("lambda-initialization : get_images.go : env can not be empty ENV\n")
		os.Exit(1)
	}
	fmt.Printf("lambda-initialization : get_images.go : start with ENV = [%s]\n", env)

	papertrailAddress, ok = os.LookupEnv("PAPERTRAIL_LOG_ADDRESS")
	if !ok {
		fmt.Printf("lambda-initialization : get_images.go : env can not be empty PAPERTRAIL_LOG_ADDRESS\n")
		os.Exit(1)
	}
	fmt.Printf("lambda-initialization : get_images.go : start with PAPERTRAIL_LOG_ADDRESS = [%s]\n", papertrailAddress)

	anlogger, err = syslog.New(papertrailAddress, fmt.Sprintf("%s-%s", env, "internal-get-images-image"))
	if err != nil {
		fmt.Errorf("lambda-initialization : get_images.go : error during startup : %v\n", err)
		os.Exit(1)
	}
	anlogger.Debugf(nil, "lambda-initialization : get_images.go : logger was successfully initialized")

	userPhotoTable, ok = os.LookupEnv("USER_PHOTO_TABLE")
	if !ok {
		fmt.Printf("lambda-initialization : get_images.go : env can not be empty USER_PHOTO_TABLE")
		os.Exit(1)
	}
	anlogger.Debugf(nil, "lambda-initialization : get_images.go : start with USER_PHOTO_TABLE = [%s]", userPhotoTable)

	awsSession, err = session.NewSession(aws.NewConfig().
		WithRegion(apimodel.Region).WithMaxRetries(apimodel.MaxRetries).
		WithLogger(aws.LoggerFunc(func(args ...interface{}) { anlogger.AwsLog(args) })).WithLogLevel(aws.LogOff))
	if err != nil {
		anlogger.Fatalf(nil, "lambda-initialization : get_images.go : error during initialization : %v", err)
	}
	anlogger.Debugf(nil, "lambda-initialization : get_images.go : aws session was successfully initialized")

	awsDbClient = dynamodb.New(awsSession)
	anlogger.Debugf(nil, "lambda-initialization : get_images.go : dynamodb client was successfully initialized")

}

func handler(ctx context.Context, request apimodel.GetNewFacesResp) (apimodel.FacesWithUrlResp, error) {
	lc, _ := lambdacontext.FromContext(ctx)

	anlogger.Debugf(lc, "get_images.go : start handle request, isItWarmUpRequest [%v],  profiles %v", request.WarmUpRequest, request.Profiles)

	if request.WarmUpRequest {
		return apimodel.FacesWithUrlResp{}, nil
	}

	respChan := make(chan map[string]string)
	batchCounter := 0
	userIdPhotos := make([]map[string]string, 0)
	for _, eachProfile := range request.Profiles {
		targetUserId := eachProfile.UserId
		for _, eachPhoto := range eachProfile.Photos {
			targetPhotoId := eachPhoto.PhotoId
			eachMap := make(map[string]string)
			eachMap[targetUserId] = targetPhotoId
			userIdPhotos = append(userIdPhotos, eachMap)
			if len(userIdPhotos) >= defaultBatchSize {
				go photos(userIdPhotos, respChan, lc)
				batchCounter++
				userIdPhotos = make([]map[string]string, 0)
			}
		}
	}
	if len(userIdPhotos) > 0 {
		go photos(userIdPhotos, respChan, lc)
		batchCounter++
	}

	finalMap := make(map[string]string)
	for i := 0; i < batchCounter; i++ {
		resMap := <-respChan
		for k, v := range resMap {
			finalMap[k] = v
		}
	}

	resp := apimodel.FacesWithUrlResp{
		UserIdPhotoIdKeyUrlMap: finalMap,
	}

	anlogger.Debugf(lc, "get_images.go : return successful resp %v", resp)
	return resp, nil
}

//as an argument function receives list with maps where each key is userId and value is photoId
//return map where each key is userId_photoId and value is photo url, ok and error string
func photos(userIdPhotos []map[string]string, respChan chan<- map[string]string, lc *lambdacontext.LambdaContext) {
	anlogger.Debugf(lc, "get_images.go : make batch request to fetch photos, len is %d", len(userIdPhotos))
	keys := make([]map[string]*dynamodb.AttributeValue, 0)
	for _, paramMap := range userIdPhotos {
		eachMap := make(map[string]*dynamodb.AttributeValue)
		for k, v := range paramMap {
			eachMap[apimodel.UserIdColumnName] = &dynamodb.AttributeValue{
				S: aws.String(k),
			}
			eachMap[apimodel.PhotoIdColumnName] = &dynamodb.AttributeValue{
				S: aws.String(v),
			}
		}
		keys = append(keys, eachMap)
	}
	keysAndAttributes := &dynamodb.KeysAndAttributes{
		ConsistentRead: aws.Bool(true),
		Keys:           keys,
	}
	requestItems := make(map[string]*dynamodb.KeysAndAttributes)
	requestItems[userPhotoTable] = keysAndAttributes

	input := &dynamodb.BatchGetItemInput{
		RequestItems: requestItems,
	}

	result, err := awsDbClient.BatchGetItem(input)
	if err != nil {
		anlogger.Errorf(lc, "get_images.go : error while making batch request to fetch photos : %v", err)
		respChan <- make(map[string]string)
		return
	}

	if len(result.UnprocessedKeys) != 0 {
		//todo: later we can handle this case and retry operation
		anlogger.Warnf(lc, "get_images.go : error while making batch request to fetch photos, there are UnprocessedKeys %v", result.UnprocessedKeys)
	}

	resultMap := make(map[string]string)
	for _, attributeList := range result.Responses {
		for _, eachAttr := range attributeList {
			targetUserId := *eachAttr[apimodel.UserIdColumnName].S
			targetPhotoId := *eachAttr[apimodel.PhotoIdColumnName].S
			_, wasPhotoDeleted := eachAttr[apimodel.PhotoDeletedAtColumnName]
			_, wasHidden := eachAttr[apimodel.PhotoHiddenAtColumnName]
			if wasPhotoDeleted || wasHidden {
				anlogger.Debugf(lc, "get_images.go : photo with userId [%s] and photoId [%s] is deleted or hidden, so exclude it from response", targetUserId, targetPhotoId)
				continue
			}
			targetPhotoUriAttr, ok := eachAttr[apimodel.PhotoSourceUriColumnName]
			if !ok {
				anlogger.Debugf(lc, "get_images.go : photo with userId [%s] and photoId [%s] don't have uri, so exclude it from response", targetUserId, targetPhotoId)
				continue
			}
			resultMap[targetUserId+"_"+targetPhotoId] = *targetPhotoUriAttr.S
		}
	}
	anlogger.Debugf(lc, "get_images.go : successfully fetch photos [%d] with batch request", len(resultMap))
	respChan <- resultMap
}

func main() {
	basicLambda.Start(handler)
}

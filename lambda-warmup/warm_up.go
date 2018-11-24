package main

import (
	"context"
	basicLambda "github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"os"
	"fmt"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/aws/aws-sdk-go/service/lambda"
	"strings"
	"github.com/ringoid/commons"
)

var anlogger *commons.Logger
var clientLambda *lambda.Lambda
var allLambdaNames string

func init() {
	var env string
	var ok bool
	var papertrailAddress string
	var err error
	var awsSession *session.Session

	env, ok = os.LookupEnv("ENV")
	if !ok {
		fmt.Printf("lambda-initialization : warm_up_image.go : env can not be empty ENV\n")
		os.Exit(1)
	}
	fmt.Printf("lambda-initialization : warm_up_image.go : start with ENV = [%s]\n", env)

	papertrailAddress, ok = os.LookupEnv("PAPERTRAIL_LOG_ADDRESS")
	if !ok {
		fmt.Printf("lambda-initialization : warm_up_image.go : env can not be empty PAPERTRAIL_LOG_ADDRESS\n")
		os.Exit(1)
	}
	fmt.Printf("lambda-initialization : warm_up_image.go : start with PAPERTRAIL_LOG_ADDRESS = [%s]\n", papertrailAddress)

	anlogger, err = commons.New(papertrailAddress, fmt.Sprintf("%s-%s", env, "warm-up-image"))
	if err != nil {
		fmt.Errorf("lambda-initialization : warm_up_image.go : error during startup : %v\n", err)
		os.Exit(1)
	}
	anlogger.Debugf(nil, "lambda-initialization : warm_up_image.go : logger was successfully initialized")

	allLambdaNames, ok = os.LookupEnv("NEED_WARM_UP_LAMBDA_NAMES")
	if !ok {
		anlogger.Fatalf(nil, "lambda-initialization : warm_up_image.go : env can not be empty NEED_WARM_UP_LAMBDA_NAMES")
	}
	anlogger.Debugf(nil, "lambda-initialization : warm_up_image.go : start with NEED_WARM_UP_LAMBDA_NAMES = [%s]", allLambdaNames)

	awsSession, err = session.NewSession(aws.NewConfig().
		WithRegion(commons.Region).WithMaxRetries(commons.MaxRetries).
		WithLogger(aws.LoggerFunc(func(args ...interface{}) { anlogger.AwsLog(args) })).WithLogLevel(aws.LogOff))
	if err != nil {
		anlogger.Fatalf(nil, "lambda-initialization : warm_up_image.go : error during initialization : %v", err)
	}
	anlogger.Debugf(nil, "lambda-initialization : warm_up_image.go : aws session was successfully initialized")

	clientLambda = lambda.New(awsSession)
	anlogger.Debugf(nil, "lambda-initialization : warm_up_image.go : lambda client was successfully initialized")
}

func handler(ctx context.Context, request events.CloudWatchEvent) error {
	lc, _ := lambdacontext.FromContext(ctx)
	names := strings.Split(allLambdaNames, ",")
	for _, n := range names {
		commons.WarmUpLambda(n, clientLambda, anlogger, lc)
	}
	return nil
}

func main() {
	basicLambda.Start(handler)
}

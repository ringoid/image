package apimodel

import (
	"../sys_log"
	"github.com/aws/aws-sdk-go/aws"
	"encoding/json"
	"github.com/aws/aws-sdk-go/service/firehose"
	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/aws/aws-sdk-go/service/kinesis"
	"github.com/aws/aws-sdk-go/service/s3"
)

//return userId, ok, error string
func CallVerifyAccessToken(accessToken, functionName string, clientLambda *lambda.Lambda, anlogger *syslog.Logger, lc *lambdacontext.LambdaContext) (string, bool, string) {
	req := InternalGetUserIdReq{
		AccessToken: accessToken,
	}
	jsonBody, err := json.Marshal(req)
	if err != nil {
		anlogger.Errorf(lc, "common_action.go : error marshaling req %s into json : %v", req, err)
		return "", false, InternalServerError
	}

	resp, err := clientLambda.Invoke(&lambda.InvokeInput{FunctionName: aws.String(functionName), Payload: jsonBody})
	if err != nil {
		anlogger.Errorf(lc, "common_action.go : error invoke function [%s] with body %s : %v", functionName, jsonBody, err)
		return "", false, InternalServerError
	}

	if *resp.StatusCode != 200 {
		anlogger.Errorf(lc, "common_action.go : status code = %d, response body %s for request %s", *resp.StatusCode, string(resp.Payload), jsonBody)
		return "", false, InternalServerError
	}

	var response InternalGetUserIdResp
	err = json.Unmarshal(resp.Payload, &response)
	if err != nil {
		anlogger.Errorf(lc, "common_action.go : error unmarshaling response %s into json : %v", string(resp.Payload), err)
		return "", false, InternalServerError
	}

	if response.ErrorCode != "" {
		anlogger.Errorf(lc, "common_action.go : error response from function [%s], response=%v", functionName, response)
		switch response.ErrorCode {
		case "InvalidAccessTokenClientError":
			return "", false, InvalidAccessTokenClientError
		default:
			return "", false, InternalServerError
		}
	}

	anlogger.Debugf(lc, "common_action.go : successfully validate accessToken, userId [%s]", response.UserId)
	return response.UserId, true, ""
}

func SendAnalyticEvent(event interface{}, userId, deliveryStreamName string, awsDeliveryStreamClient *firehose.Firehose,
	anlogger *syslog.Logger, lc *lambdacontext.LambdaContext) {
	anlogger.Debugf(lc, "common_action.go : send analytics event [%v] for userId [%s]", event, userId)
	data, err := json.Marshal(event)
	if err != nil {
		anlogger.Errorf(lc, "common_action.go : error marshaling analytics event [%v] for userId [%s] : %v", event, userId, err)
		return
	}
	newLine := "\n"
	data = append(data, newLine...)
	_, err = awsDeliveryStreamClient.PutRecord(&firehose.PutRecordInput{
		DeliveryStreamName: aws.String(deliveryStreamName),
		Record: &firehose.Record{
			Data: data,
		},
	})

	if err != nil {
		anlogger.Errorf(lc, "common_action.go : error sending analytics event [%v] for userId [%s] : %v", event, userId, err)
	}

	anlogger.Debugf(lc, "common_action.go : successfully send analytics event [%v] for userId [%s]", event, userId)
}

//ok and error string
func SendCommonEvent(event interface{}, userId, commonStreamName string, awsKinesisClient *kinesis.Kinesis,
	anlogger *syslog.Logger, lc *lambdacontext.LambdaContext) (bool, string) {
	anlogger.Debugf(lc, "common_action.go : send common event [%v] for userId [%s]", event, userId)
	data, err := json.Marshal(event)
	if err != nil {
		anlogger.Errorf(lc, "common_action.go : error marshaling common event [%v] for userId [%s] : %v", event, userId, err)
		return false, InternalServerError
	}
	input := &kinesis.PutRecordInput{
		StreamName:   aws.String(commonStreamName),
		PartitionKey: aws.String(userId),
		Data:         []byte(data),
	}
	_, err = awsKinesisClient.PutRecord(input)
	if err != nil {
		anlogger.Errorf(lc, "common_action.go : error putting common event into stream, event [%v] for userId [%s] : %v", event, userId, err)
		return false, InternalServerError
	}
	anlogger.Debugf(lc, "common_action.go : successfully send common event [%v] for userId [%s]", event, userId)
	return true, ""
}

//return ok and error string
func SendAsyncTask(task interface{}, asyncTaskQueue, userId string,
	awsSqsClient *sqs.SQS, anlogger *syslog.Logger, lc *lambdacontext.LambdaContext) (bool, string) {
	anlogger.Debugf(lc, "common_action.go : send async task %v for userId [%s]", task, userId)
	body, err := json.Marshal(task)
	if err != nil {
		anlogger.Errorf(lc, "common_action.go : error marshal task %v for userId [%s]: %v", task, userId, err)
		return false, InternalServerError
	}
	input := &sqs.SendMessageInput{
		QueueUrl:    aws.String(asyncTaskQueue),
		MessageBody: aws.String(string(body)),
	}
	_, err = awsSqsClient.SendMessage(input)
	if err != nil {
		anlogger.Errorf(lc, "common_action.go : error sending async task %v to the queue for userId [%s] : %v", task, userId, err)
		return false, InternalServerError
	}
	return true, ""
}

//return ok and error string
func DeleteFromS3(bucket, key, userId string, awsS3Client *s3.S3, lc *lambdacontext.LambdaContext, anlogger *syslog.Logger) (bool, string) {
	anlogger.Debugf(lc, "common_action.go : delete from s3 bucket [%s] with key [%s] for userId [%s]",
		bucket, key, userId)

	input := &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}

	_, err := awsS3Client.DeleteObject(input)
	if err != nil {
		anlogger.Errorf(lc, "common_action.go : error delete from s3 bucket [%s] with key [%s] for userId [%s] : %v",
			bucket, key, userId, err)
		return false, InternalServerError
	}

	anlogger.Debugf(lc, "common_action.go : successfully delete from s3 bucket [%s] with key [%s] for userId [%s]",
		bucket, key, userId)
	return true, ""
}

package main

import (
	"fmt"
	"github.com/aws/aws-lambda-go/lambdacontext"
	"../apimodel"
	"github.com/aws/aws-sdk-go/service/s3"
	"encoding/json"
	"errors"
	"github.com/ringoid/commons"
)

func removeS3Object(body []byte, awsS3Client *s3.S3, lc *lambdacontext.LambdaContext, anlogger *commons.Logger) error {
	var rTask apimodel.RemoveS3ObjectAsyncTask
	err := json.Unmarshal([]byte(body), &rTask)
	if err != nil {
		anlogger.Errorf(lc, "remove_s3_object.go : error unmarshal body [%s] to RemoveS3ObjectAsyncTask: %v", body, err)
		return errors.New(fmt.Sprintf("error unmarshal body %s : %v", body, err))
	}
	ok, errStr := commons.DeleteFromS3(rTask.Bucket, rTask.Key, "admin", awsS3Client, lc, anlogger)
	if !ok {
		return errors.New(errStr)
	}
	return nil
}

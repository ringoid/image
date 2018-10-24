package main

import (
	"github.com/anthonynsimon/bild/imgio"
	"github.com/aws/aws-lambda-go/lambdacontext"
	"../sys_log"
	"../apimodel"
	"bytes"
	"fmt"
	"encoding/json"
	"errors"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/anthonynsimon/bild/transform"
	"image"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

const defaultJPEGQuality = 80

func resizePhoto(body []byte, downloader *s3manager.Downloader, uploader *s3manager.Uploader, awsDbClient *dynamodb.DynamoDB, lc *lambdacontext.LambdaContext, anlogger *syslog.Logger) error {
	anlogger.Debugf(lc, "resize_photo.go : resize photo by request body [%s]", body)
	var rTask apimodel.ResizePhotoAsyncTask
	err := json.Unmarshal([]byte(body), &rTask)
	if err != nil {
		anlogger.Errorf(lc, "resize_photo.go : error unmarshal body [%s] to ResizePhotoAsyncTask: %v", body, err)
		return errors.New(fmt.Sprintf("error unmarshal body %s : %v", body, err))
	}

	sourceImage, ok, errStr := getImage(rTask.SourceBucket, rTask.SourceKey, rTask.UserId, downloader, lc, anlogger)
	if !ok {
		return errors.New(errStr)
	}
	img, _, err := image.Decode(bytes.NewReader(sourceImage))
	if err != nil {
		anlogger.Errorf(lc, "resize_photo.go : error decode image file from bucket [%s], key [%s] for userId [%s] : %v",
			rTask.SourceBucket, rTask.SourceKey, rTask.UserId, err)
		return errors.New(apimodel.InternalServerError)
	}

	width := rTask.TargetWidth
	height := rTask.TargetHeight

	resized := transform.Resize(img, width, height, transform.Linear)
	result := bytes.Buffer{}
	err = imgio.JPEGEncoder(defaultJPEGQuality)(&result, resized)
	if err != nil {
		anlogger.Errorf(lc, "resize_photo.go : error encode image file from bucket [%s], key [%s] with target width [%d] and target height [%d] for userId [%s] : %v",
			rTask.SourceBucket, rTask.SourceKey, width, height, rTask.UserId, err)
		return errors.New(apimodel.InternalServerError)
	}
	link := fmt.Sprintf("https://s3-eu-west-1.amazonaws.com/%s/%s", rTask.TargetBucket, rTask.TargetKey)
	userPhoto := &apimodel.UserPhoto{
		UserId:         rTask.UserId,
		PhotoId:        rTask.PhotoId,
		PhotoType:      rTask.PhotoType,
		Bucket:         rTask.TargetBucket,
		Key:            rTask.TargetKey,
		Size:           int64(len(result.Bytes())),
		PhotoSourceUri: link,
	}

	ok, errStr = apimodel.SavePhoto(userPhoto, rTask.TableName, awsDbClient, anlogger, lc)
	//ok, errStr = savePhoto(userPhoto, rTask.TableName, awsDbClient, lc, anlogger)
	if !ok && len(errStr) != 0 {
		return errors.New(errStr)
	} else if !ok && len(errStr) == 0 {
		return nil
	}

	ok, errStr = uploadImage(result.Bytes(), rTask.TargetBucket, rTask.TargetKey, rTask.UserId, uploader, lc, anlogger)
	if !ok {
		return errors.New(errStr)
	}

	anlogger.Debugf(lc, "resize_photo.go : successfully resize photo by request %v for userId [%s]", rTask, rTask.UserId)
	return nil
}

//return image, ok and error string
func getImage(bucket, key, userId string, downloader *s3manager.Downloader, lc *lambdacontext.LambdaContext, anlogger *syslog.Logger) ([]byte, bool, string) {
	anlogger.Debugf(lc, "resize_photo.go : get image from bucket [%s] with a key [%s] for userId [%s]", bucket, key, userId)

	buff := &aws.WriteAtBuffer{}
	_, err := downloader.Download(buff, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		anlogger.Errorf(lc, "resize_photo.go : error downloading image from bucket [%s] with a key [%s] for userId [%s] : %v",
			bucket, key, userId, err)
		return nil, false, apimodel.InternalServerError
	}
	anlogger.Debugf(lc, "resize_photo.go : successfully got image from bucket [%s] with a key [%s] for userId [%s]", bucket, key, userId)
	return buff.Bytes(), true, ""
}

//return ok and error string
func uploadImage(source []byte, bucket, key, userId string, uploader *s3manager.Uploader, lc *lambdacontext.LambdaContext, anlogger *syslog.Logger) (bool, string) {
	anlogger.Debugf(lc, "resize_photo.go : upload image to bucket [%s] with a key [%s] for userId [%s]", bucket, key, userId)
	_, err := uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   bytes.NewReader(source),
		ACL:    aws.String("public-read"),
	})
	if err != nil {
		anlogger.Errorf(lc, "resize_photo.go : error upload image to bucket [%s] with a key [%s] for userId [%s] : %v", bucket, key, userId, err)
		return false, apimodel.InternalServerError
	}
	anlogger.Debugf(lc, "resize_photo.go : successfully uploaded image to bucket [%s] with a key [%s] for userId [%s]", bucket, key, userId)
	return true, ""
}

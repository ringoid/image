package main

import (
	"github.com/anthonynsimon/bild/imgio"
	"github.com/aws/aws-lambda-go/lambdacontext"
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
	"github.com/ringoid/commons"
	"github.com/aws/aws-sdk-go/service/kinesis"
)

func resizePhoto(body []byte, downloader *s3manager.Downloader, uploader *s3manager.Uploader, awsDbClient *dynamodb.DynamoDB,
	commonStreamName string, awsKinesisClient *kinesis.Kinesis, lc *lambdacontext.LambdaContext, anlogger *commons.Logger) error {
	anlogger.Debugf(lc, "resize_photo.go : resize photo by request body [%s]", body)
	var rTask apimodel.ResizePhotoAsyncTask
	err := json.Unmarshal([]byte(body), &rTask)
	if err != nil {
		anlogger.Errorf(lc, "resize_photo.go : error unmarshal body [%s] to ResizePhotoAsyncTask: %v", body, err)
		return errors.New(fmt.Sprintf("error unmarshal body %s : %v", body, err))
	}

	buff, err := getImage(rTask.SourceBucket, rTask.SourceKey, rTask.UserId, downloader, lc, anlogger)
	if err != nil {
		isExist, ok, errStr := isPhotoExist(rTask.UserId, rTask.OriginPhotoId, rTask.TableName, lc, anlogger)
		if !ok {
			return errors.New(errStr)
		}
		if isExist {
			anlogger.Errorf(lc, "resize_photo.go : error downloading image from bucket [%s] with a key [%s] for userId [%s]: %v",
				rTask.SourceBucket, rTask.SourceKey, rTask.UserId, err)
			return errors.New(commons.InternalServerError)
		}
		anlogger.Infof(lc, "resize_photo.go : don't need resize photo with originId [%s] and resizedId [%s] for userId [%s] coz photo already deleted or hidden",
			rTask.OriginPhotoId, rTask.PhotoId, rTask.UserId)
		return nil
	}

	img, _, err := image.Decode(bytes.NewReader(buff.Bytes()))
	if err != nil {
		anlogger.Errorf(lc, "resize_photo.go : error decode image file from bucket [%s], key [%s] for userId [%s] : %v",
			rTask.SourceBucket, rTask.SourceKey, rTask.UserId, err)
		return errors.New(commons.InternalServerError)
	}

	width := rTask.TargetWidth
	height := rTask.TargetHeight
	resolution := fmt.Sprintf("%vx%v", width, height)
	if rTask.PhotoType == commons.ThumbnailPhotoType {
		resolution = resolution + "_" + commons.ThumbnailPhotoType
	}
	resized := transform.Resize(img, width, height, transform.Linear)
	result := bytes.Buffer{}
	err = imgio.JPEGEncoder(rTask.PhotoQuality)(&result, resized)
	if err != nil {
		anlogger.Errorf(lc, "resize_photo.go : error encode image file from bucket [%s], key [%s] with target width [%d] and target height [%d] for userId [%s] : %v",
			rTask.SourceBucket, rTask.SourceKey, width, height, rTask.UserId, err)
		return errors.New(commons.InternalServerError)
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

	ok, errStr := apimodel.SavePhoto(userPhoto, rTask.TableName, awsDbClient, anlogger, lc)
	if !ok && len(errStr) != 0 {
		return errors.New(errStr)
	} else if !ok && len(errStr) == 0 {
		anlogger.Infof(lc, "resize_photo.go : photo with originId [%s] and resizedId [%s] for userId [%s] was deleted during resize, so don't save resized photo in DB",
			rTask.OriginPhotoId, rTask.PhotoId, rTask.UserId)
		return nil
	}

	ok, errStr = uploadImage(result.Bytes(), rTask.TargetBucket, rTask.TargetKey, rTask.UserId, uploader, lc, anlogger)
	if !ok {
		return errors.New(errStr)
	}

	event := commons.NewPhotoResizeEvent(rTask.UserId, rTask.OriginPhotoId, rTask.PhotoId, resolution, link)
	ok, errStr = commons.SendCommonEvent(event, rTask.UserId, commonStreamName, rTask.UserId, awsKinesisClient, anlogger, lc)
	if !ok {
		return errors.New(errStr)
	}

	anlogger.Infof(lc, "resize_photo.go : successfully resize photo with originId [%s] and resizedId [%s] for userId [%s]",
		rTask.OriginPhotoId, rTask.PhotoId, rTask.UserId)
	return nil
}

//return image, ok and error string
func getImage(bucket, key, userId string, downloader *s3manager.Downloader, lc *lambdacontext.LambdaContext, anlogger *commons.Logger) (*aws.WriteAtBuffer, error) {
	anlogger.Debugf(lc, "resize_photo.go : get image from bucket [%s] with a key [%s] for userId [%s]", bucket, key, userId)

	buff := &aws.WriteAtBuffer{}
	_, err := downloader.Download(buff, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	return buff, err
}

//return ok and error string
func uploadImage(source []byte, bucket, key, userId string, uploader *s3manager.Uploader, lc *lambdacontext.LambdaContext, anlogger *commons.Logger) (bool, string) {
	anlogger.Debugf(lc, "resize_photo.go : upload image to bucket [%s] with a key [%s] for userId [%s]", bucket, key, userId)
	_, err := uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   bytes.NewReader(source),
		ACL:    aws.String("public-read"),
	})
	if err != nil {
		anlogger.Errorf(lc, "resize_photo.go : error upload image to bucket [%s] with a key [%s] for userId [%s] : %v", bucket, key, userId, err)
		return false, commons.InternalServerError
	}
	anlogger.Debugf(lc, "resize_photo.go : successfully uploaded image to bucket [%s] with a key [%s] for userId [%s]", bucket, key, userId)
	return true, ""
}

func isPhotoExist(userId, originPhotoId, tableName string, lc *lambdacontext.LambdaContext, anlogger *commons.Logger) (bool, bool, string) {
	anlogger.Debugf(lc, "resize_photo.go : check that photo with origin photoId [%s] for userId [%s] is still exist", originPhotoId, userId)
	input := &dynamodb.GetItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			commons.UserIdColumnName: {
				S: aws.String(userId),
			},
			commons.PhotoIdColumnName: {
				S: aws.String(originPhotoId),
			},
		},
		ConsistentRead: aws.Bool(true),
		TableName:      aws.String(tableName),
	}

	result, err := awsDbClient.GetItem(input)
	if err != nil {
		anlogger.Errorf(lc, "resize_photo.go : error while check that photo with origin photoId [%s] for userId [%s] is still exist : %v", originPhotoId, userId, err)
		return false, false, commons.InternalServerError
	}

	if len(result.Item) == 0 {
		anlogger.Debugf(lc, "resize_photo.go : there is no photo with origin photoId [%s] for userId [%s]", originPhotoId, userId)
		return false, true, ""
	}

	if _, wasPhotoDeleted := result.Item[commons.PhotoDeletedAtColumnName]; wasPhotoDeleted {
		anlogger.Debugf(lc, "resize_photo.go : photo with origin photoId [%s] for userId [%s] was deleted already", originPhotoId, userId)
		return false, true, ""
	}

	if _, wasHidden := result.Item[commons.PhotoHiddenAtColumnName]; wasHidden {
		anlogger.Debugf(lc, "resize_photo.go : photo with origin photoId [%s] for userId [%s] was hidden already", originPhotoId, userId)
		return false, true, ""
	}
	anlogger.Debugf(lc, "resize_photo.go : successfully check that photo with origin photoId [%s] for userId [%s] still exist", originPhotoId, userId)
	return true, true, ""
}

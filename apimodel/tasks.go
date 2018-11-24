package apimodel

import "fmt"

const (
	ImageRemovePhotoTaskType    = "IMAGE_REMOVE_PHOTO"
	ImageResizePhotoTaskType    = "IMAGE_RESIZE_PHOTO"
	ImageRemoveS3ObjectTaskType = "IMAGE_REMOVE_S3_OBJECT"
)

type AsyncTask struct {
	TaskType string `json:"taskType"`
}

func (task AsyncTask) String() string {
	return fmt.Sprintf("%#v", task)
}

type RemovePhotoAsyncTask struct {
	TaskType  string `json:"taskType"`
	UserId    string `json:"userId"`
	PhotoId   string `json:"photoId"`
	TableName string `json:"tableName"`
}

func (task RemovePhotoAsyncTask) String() string {
	return fmt.Sprintf("%#v", task)
}

func NewRemovePhotoAsyncTask(userId, photoId, tableName string) *RemovePhotoAsyncTask {
	return &RemovePhotoAsyncTask{
		TaskType:  ImageRemovePhotoTaskType,
		UserId:    userId,
		PhotoId:   photoId,
		TableName: tableName,
	}
}

type RemoveS3ObjectAsyncTask struct {
	TaskType string `json:"taskType"`
	Bucket   string `json:"bucket"`
	Key      string `json:"key"`
}

func (task RemoveS3ObjectAsyncTask) String() string {
	return fmt.Sprintf("%#v", task)
}

func NewRemoveS3ObjectAsyncTask(bucket, key string) *RemoveS3ObjectAsyncTask {
	return &RemoveS3ObjectAsyncTask{
		TaskType: ImageRemoveS3ObjectTaskType,
		Bucket:   bucket,
		Key:      key,
	}
}

type ResizePhotoAsyncTask struct {
	TaskType     string `json:"taskType"`
	UserId       string `json:"userId"`
	PhotoId      string `json:"photoId"`
	PhotoType    string `json:"photoType"`
	SourceBucket string `json:"sourceBucket"`
	SourceKey    string `json:"sourceKey"`
	TargetWidth  int    `json:"targetWidth"`
	TargetHeight int    `json:"targetHeight"`
	TargetBucket string `json:"targetBucket"`
	TargetKey    string `json:"targetKey"`
	TableName    string `json:"tableName"`
}

func (task ResizePhotoAsyncTask) String() string {
	return fmt.Sprintf("%#v", task)
}

func NewResizePhotoAsyncTask(userId, photoId, photoType, sourceBucket, sourceKey, targetBucket, targetKey, tableName string, targetWidth, targetHeight int) *ResizePhotoAsyncTask {
	return &ResizePhotoAsyncTask{
		TaskType:     ImageResizePhotoTaskType,
		UserId:       userId,
		PhotoId:      photoId,
		PhotoType:    photoType,
		SourceBucket: sourceBucket,
		SourceKey:    sourceKey,
		TargetWidth:  targetWidth,
		TargetHeight: targetHeight,
		TargetBucket: targetBucket,
		TargetKey:    targetKey,
		TableName:    tableName,
	}
}

package apimodel

import "fmt"

const (
	RemovePhotoTaskType = "REMOVE_PHOTO"
	ResizePhotoTaskType = "RESIZE_PHOTO"
)

type AsyncTask struct {
	TaskType string `json:"taskType"`
}

func (task AsyncTask) String() string {
	return fmt.Sprintf("[AsyncTask={taskType=%s}]", task)
}

type RemovePhotoAsyncTask struct {
	TaskType  string `json:"taskType"`
	UserId    string `json:"userId"`
	PhotoId   string `json:"photoId"`
	TableName string `json:"tableName"`
}

func (task RemovePhotoAsyncTask) String() string {
	return fmt.Sprintf("[RemovePhotoAsyncTask={taskType=%s, userId=%s, photoId=%s, tableName=%s}]",
		task.TaskType, task.UserId, task.PhotoId, task.TableName)
}

func NewRemovePhotoAsyncTask(userId, photoId, tableName string) *RemovePhotoAsyncTask {
	return &RemovePhotoAsyncTask{
		TaskType:  RemovePhotoTaskType,
		UserId:    userId,
		PhotoId:   photoId,
		TableName: tableName,
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
	return fmt.Sprintf("[ResizePhotoAsyncTask={taskType=%s, userId=%s, photoId=%s, photoType=%s, sourceBucket=%s, sourceKey=%s,"+
		"targetWidth=%d, targetHeight=%d, targetBucket=%s, targetKey=%s, tableName=%s}]", task.TaskType,
		task.UserId, task.PhotoId, task.PhotoType, task.SourceBucket, task.SourceKey, task.TargetWidth, task.TargetHeight,
		task.TargetBucket, task.TargetKey, task.TableName)
}

func NewResizePhotoAsyncTask(userId, photoId, photoType, sourceBucket, sourceKey, targetBucket, targetKey, tableName string, targetWidth, targetHeight int) *ResizePhotoAsyncTask {
	return &ResizePhotoAsyncTask{
		TaskType:     ResizePhotoTaskType,
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

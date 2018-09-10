package apimodel

import "fmt"

const (
	RemovePhotoTaskType = "REMOVE_PHOTO"
)

type AsyncTask struct {
	TaskType string `json:"taskType"`
}

func (task AsyncTask) String() string {
	return fmt.Sprintf("[AsyncTask={taskType=%s}]", task)
}

type RemovePhotoAsyncTask struct {
	AsyncTask
	UserId    string `json:"userId"`
	PhotoId   string `json:"photoId"`
	TableName string `json:"tableName"`
}

func (task RemovePhotoAsyncTask) String() string {
	return fmt.Sprintf("[%v, RemovePhotoAsyncTask={userId=%s, photoId=%s, tableName=%s}]",
		task.AsyncTask, task.UserId, task.PhotoId, task.TableName)
}

func NewRemovePhotoAsyncTask(userId, photoId, tableName string) *RemovePhotoAsyncTask {
	return &RemovePhotoAsyncTask{
		AsyncTask: AsyncTask{TaskType: RemovePhotoTaskType},
		UserId:    userId,
		PhotoId:   photoId,
		TableName: tableName,
	}
}

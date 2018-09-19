package apimodel

import (
	"time"
	"fmt"
)

type UserAskUploadPhotoLinkEvent struct {
	UserId    string `json:"userId"`
	Bucket    string `json:"bucket"`
	PhotoKey  string `json:"photoKey"`
	UnixTime  int64  `json:"unixTime"`
	EventType string `json:"eventType"`
}

func (event UserAskUploadPhotoLinkEvent) String() string {
	return fmt.Sprintf("[UserAskUploadPhotoLinkEvent={userId=%s, bucket=%s, photoKey=%s, unixTime=%v, eventType=%s}]", event.UserId, event.Bucket, event.PhotoKey, event.UnixTime, event.EventType)
}

func NewUserAskUploadLinkEvent(bucket, photoKey, userId string) *UserAskUploadPhotoLinkEvent {
	return &UserAskUploadPhotoLinkEvent{
		UserId:    userId,
		Bucket:    bucket,
		PhotoKey:  photoKey,
		UnixTime:  time.Now().Unix(),
		EventType: "IMAGE_USER_ASK_UPLOAD_PHOTO_LINK",
	}
}

type UserUploadedPhotoEvent struct {
	UserId    string `json:"userId"`
	Bucket    string `json:"bucket"`
	PhotoKey  string `json:"photoKey"`
	PhotoId   string `json:"photoId"`
	PhotoType string `json:"photoType"`
	Size      int64  `json:"size"`
	UnixTime  int64  `json:"unixTime"`
	EventType string `json:"eventType"`
}

func (event UserUploadedPhotoEvent) String() string {
	return fmt.Sprintf("[UserUploadedPhotoEvent={userId=%s, bucket=%s, photoKey=%s, photoId=%s, photoType=%s, size=%v, unixTime=%v, eventType=%s}]", event.UserId, event.Bucket, event.PhotoKey, event.PhotoId, event.PhotoType, event.Size, event.UnixTime, event.EventType)
}

func NewUserUploadedPhotoEvent(photo UserPhoto) *UserUploadedPhotoEvent {
	return &UserUploadedPhotoEvent{
		UserId:    photo.UserId,
		Bucket:    photo.Bucket,
		PhotoKey:  photo.Key,
		PhotoId:   photo.PhotoId,
		PhotoType: photo.PhotoType,
		Size:      photo.Size,
		UnixTime:  time.Now().Unix(),
		EventType: "IMAGE_USER_UPLOAD_PHOTO",
	}
}

type UserDeletePhotoEvent struct {
	UserId    string `json:"userId"`
	PhotoId   string `json:"photoId"`
	UnixTime  int64  `json:"unixTime"`
	EventType string `json:"eventType"`
}

func (event UserDeletePhotoEvent) String() string {
	return fmt.Sprintf("[UserDeletePhotoEvent={userId=%s, photoId=%s, unixTime=%v, eventType=%s}]",
		event.UserId, event.PhotoId, event.UnixTime, event.EventType)
}

func NewUserDeletePhotoEvent(userId, photoId string) *UserDeletePhotoEvent {
	return &UserDeletePhotoEvent{
		UserId:    userId,
		PhotoId:   photoId,
		UnixTime:  time.Now().Unix(),
		EventType: "IMAGE_USER_DELETE_PHOTO",
	}
}

type RemoveTooLargeObjectEvent struct {
	UserId    string `json:"userId"`
	Bucket    string `json:"bucket"`
	Key       string `json:"key"`
	Size      int64  `json:"size"`
	UnixTime  int64  `json:"unixTime"`
	EventType string `json:"eventType"`
}

func (event RemoveTooLargeObjectEvent) String() string {
	return fmt.Sprintf("[RemoveTooLargeObjectEvent={userId=%s, bucket=%s, key=%s, size=%v, unixTime=%v, eventType=%s}]",
		event.UserId, event.Bucket, event.Key, event.Size, event.UnixTime, event.EventType)
}

func NewRemoveTooLargeObjectEvent(userId, bucket, key string, size int64) *RemoveTooLargeObjectEvent {
	return &RemoveTooLargeObjectEvent{
		UserId:    userId,
		Bucket:    bucket,
		Key:       key,
		Size:      size,
		UnixTime:  time.Now().Unix(),
		EventType: "IMAGE_REMOVE_TO_BIG_S3_OBJECT",
	}
}

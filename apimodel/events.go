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
		EventType: "IMAGE_USER_UPLOADED_PHOTO",
	}
}

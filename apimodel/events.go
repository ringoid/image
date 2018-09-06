package apimodel

import (
	"time"
	"fmt"
)

type UserAcceptTermsEvent struct {
	UserId                     string `json:"userId"`
	Locale                     string `json:"locale"`
	SourceIp                   string `json:"sourceIp"`
	ClientValidationFail       bool   `json:"clientValidationFail"`
	UnixTime                   int64  `json:"unixTime"`
	EventType                  string `json:"eventType"`
	DateTimeTermsAndConditions int64  `json:"dtTC"`
	DateTimePrivacyNotes       int64  `json:"dtPN"`
	DateTimeLegalAge           int64  `json:"dtLA"`
	CustomerId                 string `json:"customerId"`
}

func (event UserAcceptTermsEvent) String() string {
	return fmt.Sprintf("[UserAcceptTermsEvent={userId=%s, locale=%s, sourceIp=%s, clientValidationFail=%v, customerId=%s, unixTime=%v, dtTC=%v, dtPN=%v, dtLA=%v, eventType=%s}]",
		event.UserId, event.Locale, event.SourceIp, event.ClientValidationFail, event.CustomerId, event.UnixTime, event.DateTimeTermsAndConditions, event.DateTimePrivacyNotes, event.DateTimeLegalAge, event.EventType)
}

func NewUserAcceptTermsEvent(req *StartReq, sourceIp, userId, customerId string) *UserAcceptTermsEvent {
	return &UserAcceptTermsEvent{
		UserId: userId,
		Locale: req.Locale,
		//gdpr?
		SourceIp:   sourceIp,
		CustomerId: customerId,

		ClientValidationFail:       req.ClientValidationFail,
		UnixTime:                   time.Now().Unix(),
		DateTimeLegalAge:           req.DateTimeLegalAge,
		DateTimePrivacyNotes:       req.DateTimePrivacyNotes,
		DateTimeTermsAndConditions: req.DateTimeTermsAndConditions,

		EventType: "AUTH_USER_ACCEPT_TERMS",
	}
}

type UserVerificationCompleteEvent struct {
	UserId    string `json:"userId"`
	UnixTime  int64  `json:"unixTime"`
	EventType string `json:"eventType"`
}

func (event UserVerificationCompleteEvent) String() string {
	return fmt.Sprintf("[UserVerificationCompleteEvent={userId=%s, unixTime=%v, eventType=%v}]", event.UserId, event.UnixTime, event.EventType)
}

func NewUserVerificationCompleteEvent(userId string) *UserVerificationCompleteEvent {
	return &UserVerificationCompleteEvent{
		UserId:    userId,
		UnixTime:  time.Now().Unix(),
		EventType: "AUTH_USER_COMPLETE_VERIFICATION",
	}
}

type UserProfileCreatedEvent struct {
	UserId      string `json:"userId"`
	Sex         string `json:"sex"`
	YearOfBirth int    `json:"yearOfBirth"`
	UnixTime    int64  `json:"unixTime"`
	EventType   string `json:"eventType"`
}

func (event UserProfileCreatedEvent) String() string {
	return fmt.Sprintf("[UserProfileCreatedEvent={userId=%s, sex=%s, yearOfBirth=%v, unixTime=%v, eventType=%s}]",
		event.UserId, event.Sex, event.YearOfBirth, event.UnixTime, event.EventType)
}

func NewUserProfileCreatedEvent(userId string, req *CreateReq) *UserProfileCreatedEvent {
	return &UserProfileCreatedEvent{
		UserId:      userId,
		Sex:         req.Sex,
		YearOfBirth: req.YearOfBirth,
		UnixTime:    time.Now().Unix(),
		EventType:   "AUTH_USER_PROFILE_CREATED",
	}
}

type UserSettingsUpdatedEvent struct {
	UserId              string `json:"userId"`
	WhoCanSeePhoto      string `json:"whoCanSeePhoto"`      //OPPOSITE (default) || INCOGNITO || ONLY_ME
	SafeDistanceInMeter int    `json:"safeDistanceInMeter"` // 0 (default for men) || 10 (default for women)
	PushMessages        bool   `json:"pushMessages"`        // true (default for men) || false (default for women)
	PushMatches         bool   `json:"pushMatches"`         // true (default)
	PushLikes           string `json:"pushLikes"`           //EVERY (default for men) || 10_NEW (default for women) || 100_NEW
	InAppMessages       bool   `json:"inAppMessages"`       //true (default for everybody)
	InAppMatches        bool   `json:"inAppMatches"`        //true (default for everybody)
	InAppLikes          string `json:"inAppLikes"`          //EVERY (default for everybody) || 10_NEW (default for women) || 100_NEW || NONE
	UnixTime            int64  `json:"unixTime"`
	EventType           string `json:"eventType"`
}

func (event UserSettingsUpdatedEvent) String() string {
	return fmt.Sprintf("[UserSettingsUpdatedEvent={userId=%s, whoCanSeePhoto=%s, safeDistanceInMeter=%d, pushMessages=%v, pushMatches=%v, pushLikes=%v, inAppMessages=%v, inAppMatches=%v, inAppLikes=%v, unixTime=%v, eventType=%s}]",
		event.UserId, event.WhoCanSeePhoto, event.SafeDistanceInMeter, event.PushMessages, event.PushMatches, event.PushLikes, event.InAppMessages, event.InAppMatches, event.InAppLikes, event.UnixTime, event.EventType)
}

func NewUserSettingsUpdatedEvent(settings *UserSettings) *UserSettingsUpdatedEvent {
	return &UserSettingsUpdatedEvent{
		UserId:              settings.UserId,
		WhoCanSeePhoto:      settings.WhoCanSeePhoto,
		SafeDistanceInMeter: settings.SafeDistanceInMeter,
		PushMessages:        settings.PushMessages,
		PushMatches:         settings.PushMatches,
		PushLikes:           settings.PushLikes,
		InAppMessages:       settings.InAppMessages,
		InAppMatches:        settings.InAppMatches,
		InAppLikes:          settings.InAppLikes,
		UnixTime:            time.Now().Unix(),
		EventType:           "AUTH_USER_SETTINGS_UPDATED",
	}
}

type UserLogoutEvent struct {
	UserId    string `json:"userId"`
	UnixTime  int64  `json:"unixTime"`
	EventType string `json:"eventType"`
}

func (event UserLogoutEvent) String() string {
	return fmt.Sprintf("[UserLogoutEvent={userId=%s, unixTime=%v, eventType=%s}]", event.UserId, event.UnixTime, event.EventType)
}

func NewUserLogoutEvent(userId string) *UserLogoutEvent {
	return &UserLogoutEvent{
		UserId:    userId,
		UnixTime:  time.Now().Unix(),
		EventType: "AUTH_USER_LOGOUT",
	}
}

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

package apimodel

import (
	"fmt"
)

const (
	Region     = "eu-west-1"
	MaxRetries = 3

	TwilioApiKeyName    = "twilio-api-key"
	TwilioSecretKeyBase = "%s/Twilio/Api/Key"

	SecretWordKeyName = "secret-word-key"
	SecretWordKeyBase = "%s/SecretWord"

	SessionGSIName = "sessionGSI"

	PhoneColumnName     = "phone"
	UserIdColumnName    = "user_id"
	SessionIdColumnName = "session_id"

	CountryCodeColumnName      = "country_code"
	PhoneNumberColumnName      = "phone_number"
	TokenUpdatedTimeColumnName = "token_updated_at"

	SessionTokenColumnName = "session_token"
	SexColumnName          = "sex"

	YearOfBirthColumnName = "year_of_birth"
	ProfileCreatedAt      = "profile_created_at"
	CustomerIdColumnName  = "customer_id"

	UpdatedTimeColumnName = "updated_at"

	PhotoIdColumnName        = "photo_id"
	PhotoSourceUriColumnName = "photo_uri"
	PhotoTypeColumnName      = "photo_type"
	PhotoBucketColumnName    = "photo_bucket"
	PhotoKeyColumnName       = "photo_key"
	PhotoSizeColumnName      = "photo_size"
	PhotoDeletedAt           = "deleted_at"

	WhoCanSeePhotoColumnName      = "who_can_see_photo"
	SafeDistanceInMeterColumnName = "safe_distance_in_meter"
	PushMessagesColumnName        = "push_messages"
	PushMatchesColumnName         = "push_matches"
	PushLikesColumnName           = "push_likes"
	InAppMessagesColumnName       = "in_app_messages"
	InAppMatchesColumnName        = "in_app_matches"
	InAppLikesColumnName          = "in_app_likes"

	AccessTokenUserIdClaim       = "userId"
	AccessTokenSessionTokenClaim = "sessionToken"

	InternalServerError           = `{"errorCode":"InternalServerError","errorMessage":"Internal Server Error"}`
	WrongRequestParamsClientError = `{"errorCode":"WrongParamsClientError","errorMessage":"Wrong request params"}`
	PhoneNumberClientError        = `{"errorCode":"PhoneNumberClientError","errorMessage":"Phone number is invalid"}`
	CountryCallingCodeClientError = `{"errorCode":"CountryCallingCodeClientError","errorMessage":"Country code is invalid"}`

	WrongSessionIdClientError        = `{"errorCode":"WrongSessionIdClientError","errorMessage":"Session id is invalid"}`
	NoPendingVerificationClientError = `{"errorCode":"NoPendingVerificationClientError","errorMessage":"No pending verifications found"}`
	WrongVerificationCodeClientError = `{"errorCode":"WrongVerificationCodeClientError","errorMessage":"Wrong verification code"}`

	WrongYearOfBirthClientError   = `{"errorCode":"WrongYearOfBirthClientError","errorMessage":"Wrong year of birth"}`
	WrongSexClientError           = `{"errorCode":"WrongSexClientError","errorMessage":"Wrong sex"}`
	InvalidAccessTokenClientError = `{"errorCode":"InvalidAccessTokenClientError","errorMessage":"Invalid access token"}`
)

var AllowedPhotoResolution map[string]bool

type BaseResponse struct {
	ErrorCode    string `json:"errorCode"`
	ErrorMessage string `json:"errorMessage"`
}

func (resp BaseResponse) String() string {
	return fmt.Sprintf("[BaseResponse={errorCode=%s, errorMessage=%s}", resp.ErrorCode, resp.ErrorMessage)
}

func init() {
	AllowedPhotoResolution = make(map[string]bool)
	AllowedPhotoResolution["640x480"] = true
}

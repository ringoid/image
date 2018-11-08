package apimodel

import (
	"fmt"
)

type WarmUpRequest struct {
	WarmUpRequest bool `json:"warmUpRequest"`
}

func (req WarmUpRequest) String() string {
	return fmt.Sprintf("%#v", req)
}

type InternalGetUserIdReq struct {
	WarmUpRequest bool   `json:"warmUpRequest"`
	AccessToken   string `json:"accessToken"`
	BuildNum      int    `json:"buildNum"`
	IsItAndroid   bool   `json:"isItAndroid"`
}

func (req InternalGetUserIdReq) String() string {
	return fmt.Sprintf("%#v", req)
}

type InternalGetUserIdResp struct {
	BaseResponse
	UserId         string `json:"userId"`
	IsUserReported bool   `json:"isUserReported"`
}

func (resp InternalGetUserIdResp) String() string {
	return fmt.Sprintf("%#v", resp)
}

type GetPresignUrlReq struct {
	WarmUpRequest bool   `json:"warmUpRequest"`
	AccessToken   string `json:"accessToken"`
	Extension     string `json:"extension"`
	ClientPhotoId string `json:"clientPhotoId"`
}

func (req GetPresignUrlReq) String() string {
	return fmt.Sprintf("%#v", req)
}

type GetPresignUrlResp struct {
	BaseResponse
	Uri           string `json:"uri"`
	OriginPhotoId string `json:"originPhotoId"`
	ClientPhotoId string `json:"clientPhotoId"`
}

func (resp GetPresignUrlResp) GoString() string {
	return fmt.Sprintf("%#v", resp)
}

type MakePresignUrlInternalReq struct {
	WarmUpRequest bool   `json:"warmUpRequest"`
	Bucket        string `json:"bucket"`
	Key           string `json:"key"`
}

func (req MakePresignUrlInternalReq) String() string {
	return fmt.Sprintf("%#v", req)
}

type MakePresignUrlInternalResp struct {
	Uri string `json:"uri"`
}

func (resp MakePresignUrlInternalResp) String() string {
	return fmt.Sprintf("%#v", resp)
}

type GetOwnPhotosResp struct {
	BaseResponse
	Photos []OwnPhoto `json:"photos"`
}

func (resp GetOwnPhotosResp) String() string {
	return fmt.Sprintf("%#v", resp)
}

type OwnPhoto struct {
	PhotoId       string `json:"photoId"`
	PhotoUri      string `json:"photoUri"`
	Likes         int    `json:"likes"`
	OriginPhotoId string `json:"originPhotoId"`
}

func (obj OwnPhoto) String() string {
	return fmt.Sprintf("%#v", obj)
}

type DeletePhotoReq struct {
	WarmUpRequest bool   `json:"warmUpRequest"`
	AccessToken   string `json:"accessToken"`
	PhotoId       string `json:"photoId"`
}

func (req DeletePhotoReq) String() string {
	return fmt.Sprintf("%#v", req)
}

type Profile struct {
	UserId string  `json:"userId"`
	Photos []Photo `json:"photos"`
}

func (p Profile) String() string {
	return fmt.Sprintf("%#v", p)
}

type Photo struct {
	PhotoId  string `json:"photoId"`
	PhotoUri string `json:"photoUri"`
}

func (p Photo) String() string {
	return fmt.Sprintf("%#v", p)
}

type GetNewFacesResp struct {
	BaseResponse
	WarmUpRequest bool      `json:"warmUpRequest"`
	Profiles      []Profile `json:"profiles"`
}

func (resp GetNewFacesResp) String() string {
	return fmt.Sprintf("%#v", resp)
}

type FacesWithUrlResp struct {
	//contains userId_photoId like a key and photoUrl like a value
	UserIdPhotoIdKeyUrlMap map[string]string `json:"urlPhotoMap"`
}

func (resp FacesWithUrlResp) String() string {
	return fmt.Sprintf("%#v", resp)
}

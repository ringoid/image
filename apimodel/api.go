package apimodel

import (
	"fmt"
)

type WarmUpRequest struct {
	WarmUpRequest bool `json:"warmUpRequest"`
}

func (req WarmUpRequest) String() string {
	return fmt.Sprintf("[WarmUpRequest={warmUpRequest=%s}]", req.WarmUpRequest)
}

type InternalGetUserIdReq struct {
	WarmUpRequest bool `json:"warmUpRequest"`
	AccessToken string `json:"accessToken"`
}

func (req InternalGetUserIdReq) String() string {
	return fmt.Sprintf("[InternalGetUserIdReq={accessToken=%s}]", req.AccessToken)
}

type InternalGetUserIdResp struct {
	BaseResponse
	UserId string `json:"userId"`
}

func (resp InternalGetUserIdResp) String() string {
	return fmt.Sprintf("[%v, InternalGetUserIdResp={userId=%s}]",
		resp.BaseResponse, resp.UserId)
}

type GetPresignUrlReq struct {
	WarmUpRequest bool `json:"warmUpRequest"`
	AccessToken   string `json:"accessToken"`
	Extension     string `json:"extension"`
	ClientPhotoId string `json:"clientPhotoId"`
}

func (req GetPresignUrlReq) String() string {
	return fmt.Sprintf("[GetPresignUrlReq={accessToken=%s, extension=%s, clientPhotoId=%s}]", req.AccessToken, req.Extension, req.ClientPhotoId)
}

type GetPresignUrlResp struct {
	BaseResponse
	Uri           string `json:"uri"`
	OriginPhotoId string `json:"originPhotoId"`
	ClientPhotoId string `json:"clientPhotoId"`
}

func (resp GetPresignUrlResp) GoString() string {
	return fmt.Sprintf("[%v, GetPresignUrlResp={uri=%s, originPhotoId=%s, clientId=%s}]", resp.BaseResponse, resp.Uri, resp.OriginPhotoId, resp.ClientPhotoId)
}

type MakePresignUrlInternalReq struct {
	WarmUpRequest bool `json:"warmUpRequest"`
	Bucket string `json:"bucket"`
	Key    string `json:"key"`
}

func (req MakePresignUrlInternalReq) String() string {
	return fmt.Sprintf("[MakePresignUrlInternalReq={bucket=%s, key=%s}]", req.Bucket, req.Key)
}

type MakePresignUrlInternalResp struct {
	Uri string `json:"uri"`
}

func (resp MakePresignUrlInternalResp) String() string {
	return fmt.Sprintf("[MakePresignUrlInternalResp={uri=%s}]", resp.Uri)
}

type GetOwnPhotosResp struct {
	BaseResponse
	Photos []OwnPhoto `json:"photos"`
}

func (resp GetOwnPhotosResp) String() string {
	return fmt.Sprintf("[%v, GetOwnPhotosResp={photos=%v}]", resp.BaseResponse, resp.Photos)
}

type OwnPhoto struct {
	PhotoId       string `json:"photoId"`
	PhotoUri      string `json:"photoUri"`
	Likes         int    `json:"likes"`
	OriginPhotoId string `json:"originPhotoId"`
}

func (obj OwnPhoto) String() string {
	return fmt.Sprintf("[OwnPhoto={photoId=%s, photoUri=%s, likes=%d, originPhotoId=%s}]", obj.PhotoId, obj.PhotoUri, obj.Likes, obj.OriginPhotoId)
}

type DeletePhotoReq struct {
	WarmUpRequest bool `json:"warmUpRequest"`
	AccessToken string `json:"accessToken"`
	PhotoId     string `json:"photoId"`
}

func (req DeletePhotoReq) String() string {
	return fmt.Sprintf("[DeletePhotoReq={accessToken=%s, photoId=%s}]", req.AccessToken, req.PhotoId)
}

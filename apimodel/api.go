package apimodel

import (
	"fmt"
)

type InternalGetUserIdReq struct {
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
	AccessToken string `json:"accessToken"`
	Extension   string `json:"extension"`
}

func (req GetPresignUrlReq) String() string {
	return fmt.Sprintf("[GetPresignUrlReq={accessToken=%s, extension=%s}]", req.AccessToken, req.Extension)
}

type GetPresignUrlResp struct {
	BaseResponse
	Uri string `json:"uri"`
}

func (resp GetPresignUrlResp) GoString() string {
	return fmt.Sprintf("[%v, GetPresignUrlResp={uri=%s}]", resp.BaseResponse, resp.Uri)
}

type MakePresignUrlInternalReq struct {
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

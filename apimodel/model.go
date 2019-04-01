package apimodel

import "fmt"

type UserPhoto struct {
	UserId             string
	PhotoId            string
	PhotoSourceUri     string //only for public photo
	PhotoType          string //origin/resized_640x48/..
	Bucket             string
	Key                string
	Size               int64
	UpdatedAt          string
	Likes              int
	OriginPhotoId      string
	HiddenInModeration bool
}

func (model UserPhoto) String() string {
	return fmt.Sprintf("%#v", model)
}

type UserPhotoMetaInf struct {
	UserId        string
	OriginPhotoId string
	Likes         int
}

func (model UserPhotoMetaInf) String() string {
	return fmt.Sprintf("%#v", model)
}

package apimodel

import "fmt"

type UserPhoto struct {
	UserId         string
	PhotoId        string
	PhotoSourceUri string //only for public photo
	PhotoType      string //origin/resized_640x48/..
	Bucket         string
	Key            string
	Size           int64
}

func (model UserPhoto) String() string {
	return fmt.Sprintf("[UserPhoto={userId=%s, photoId=%s, photoSourceUri=%s, photoType=%s, bucket=%s, key=%s, size=%v}]",
		model.UserId, model.PhotoId, model.PhotoSourceUri, model.PhotoType, model.Bucket, model.Key, model.Size)
}

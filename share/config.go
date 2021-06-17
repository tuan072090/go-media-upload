package share

import (
	"os"
	"strconv"
)

var PORT = os.Getenv("PORT")
var MEDIA_URL = os.Getenv("MEDIA_URL")
var MAX_UPLOAD_SIZE = os.Getenv("MAX_UPLOAD_SIZE")

func GetMediaUrl() string {
	if MEDIA_URL != ""{
		return MEDIA_URL
	}

	return "localhost"
}

func GetMaxUploadSize() int64 {

	if MAX_UPLOAD_SIZE != "" {
		i, err := strconv.Atoi(MAX_UPLOAD_SIZE)
		if err != nil {
			return 10 * 1024 * 1024 // 10 mb
		}
		return int64(i)
	}
	return 10 * 1024 * 1024 // 10 mb
}

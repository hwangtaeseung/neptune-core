package awssdk

import (
	"os"
)
type S3Url struct {
	Region                  string `json:"region"`
	InputBucket             string `json:"input_bucket"`
	TempBucket              string `json:"temp_bucket"`
	OutputBucket            string `json:"output_bucket"`
	SystemSettingsBucket    string `json:"setting_bucket"`
	Key                     string `json:"key"`
	MediaId 				string `json:"media_id"`
	EncodingProfileForVideo string `json:"encoding_profile_for_video"`
	AudioType               string `json:"audio_type"`
	EncodingProfileForAudio string `json:"encoding_profile_for_audio"`
}

func GetS3UrlFromEnv() *S3Url {
	return &S3Url{
		Region:                  os.Getenv("AWS_S3_REGION"),
		InputBucket:             os.Getenv("AWS_S3_INPUT_BUCKET"),
		TempBucket:              os.Getenv("AWS_S3_TEMP_BUCKET"),
		OutputBucket:            os.Getenv("AWS_S3_OUTPUT_BUCKET"),
		SystemSettingsBucket:    os.Getenv("AWS_S3_SYSTEM_SETTINGS_BUCKET"),
		Key:                     os.Getenv("AWS_S3_KEY"),
		MediaId:                 os.Getenv("MEDIA_ID"),
		EncodingProfileForVideo: os.Getenv("ENCODING_PROFILE_VIDEO"),
		AudioType:               os.Getenv("AUDIO_TYPE"),
		EncodingProfileForAudio: os.Getenv("ENCODING_PROFILE_AUDIO"),
	}
}
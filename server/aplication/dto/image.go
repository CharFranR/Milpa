package dto

type ImageDataDTO struct {
	Content     []byte
	ContentType string
	Filepath    string
	Filename    string
}

type UploadedImageDTO struct {
	Path string `json:"path"`
}

package cloudinary

import (
	"context"
	"mime/multipart"
)

// Usecase defines the interface for Cloudinary operations
type Usecase interface {
	Upload(ctx context.Context, fileHeader *multipart.FileHeader, from string) (File, error)
}

package cloudinary

import (
	"context"
	"fmt"
	"mime/multipart"

	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
	"github.com/google/uuid"
	"github.com/nguyentantai21042004/smap-api/pkg/util"
)

func (c *implCloudinary) Upload(ctx context.Context, fileHeader *multipart.FileHeader, from string) (File, error) {
	file, err := fileHeader.Open()
	if err != nil {
		return File{}, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	pubID := from + "/" + uuid.New().String()
	uploadResult, err := c.cld.Upload.Upload(ctx, file, uploader.UploadParams{
		PublicID:       pubID,
		UseFilename:    util.ToPointer(true),
		UniqueFilename: util.ToPointer(true),
	})
	if err != nil {
		return File{}, err
	}
	return File{
		URL:      uploadResult.SecureURL,
		PublicID: uploadResult.PublicID,
		Width:    uploadResult.Width,
		Height:   uploadResult.Height,
		Format:   uploadResult.Format,
		Bytes:    uploadResult.Bytes,
		Type:     uploadResult.ResourceType,
	}, nil
}

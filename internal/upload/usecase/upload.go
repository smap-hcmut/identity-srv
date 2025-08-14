package usecase

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/nguyentantai21042004/smap-api/internal/models"
	"github.com/nguyentantai21042004/smap-api/internal/upload"
	"github.com/nguyentantai21042004/smap-api/pkg/minio"
	"github.com/nguyentantai21042004/smap-api/pkg/postgres"
)

func (uc *usecase) Create(ctx context.Context, sc models.Scope, ip upload.CreateInput) (upload.UploadOutput, error) {
	// Validate file
	if ip.FileHeader == nil {
		return upload.UploadOutput{}, upload.ErrInvalidFile
	}

	// Validate file size (max 10MB)
	if ip.FileHeader.Size > 10*1024*1024 {
		return upload.UploadOutput{}, upload.ErrFileTooLarge
	}

	// Validate bucket
	if ip.BucketName == "" {
		return upload.UploadOutput{}, upload.ErrInvalidBucket
	}

	// Generate unique object name
	ext := filepath.Ext(ip.FileHeader.Filename)
	objectName := fmt.Sprintf("%s/%s%s", time.Now().Format("2006/01/02"), postgres.NewUUID(), ext)

	// Upload to MinIO
	file, err := ip.FileHeader.Open()
	if err != nil {
		uc.l.Errorf(ctx, "internal.upload.usecase.Create.ip.FileHeader.Open: %v", err)
		return upload.UploadOutput{}, err
	}
	defer file.Close()

	// Create upload request
	uploadReq := &minio.UploadRequest{
		BucketName:   ip.BucketName,
		ObjectName:   objectName,
		OriginalName: ip.FileHeader.Filename,
		Reader:       file,
		Size:         ip.FileHeader.Size,
		ContentType:  ip.FileHeader.Header.Get("Content-Type"),
	}

	// Upload to MinIO
	fileInfo, err := uc.minio.UploadFile(ctx, uploadReq)
	if err != nil {
		uc.l.Errorf(ctx, "internal.upload.usecase.Create.uc.minio.UploadFile: %v", err)
		return upload.UploadOutput{}, err
	}

	// Create upload record
	etag := fileInfo.ETag
	url := fileInfo.URL
	uploadModel := models.Upload{
		ID:            postgres.NewUUID(),
		BucketName:    fileInfo.BucketName,
		ObjectName:    fileInfo.ObjectName,
		OriginalName:  fileInfo.OriginalName,
		Size:          fileInfo.Size,
		ContentType:   fileInfo.ContentType,
		Etag:          &etag,
		URL:           &url,
		Source:        upload.MinIO,
		CreatedUserID: sc.UserID,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	// Save to database
	createdUpload, err := uc.repo.Create(ctx, sc, upload.CreateOptions{Upload: uploadModel})
	if err != nil {
		uc.l.Errorf(ctx, "internal.upload.usecase.Create.uc.repo.Create: %v", err)
		return upload.UploadOutput{}, err
	}

	return upload.UploadOutput{Upload: createdUpload}, nil
}

func (uc *usecase) Detail(ctx context.Context, sc models.Scope, ID string) (upload.UploadOutput, error) {
	uploadModel, err := uc.repo.Detail(ctx, sc, ID)
	if err != nil {
		uc.l.Errorf(ctx, "internal.upload.usecase.Detail.uc.repo.Detail: %v", err)
		return upload.UploadOutput{}, err
	}

	return upload.UploadOutput{Upload: uploadModel}, nil
}

func (uc *usecase) Get(ctx context.Context, sc models.Scope, ip upload.GetInput) (upload.GetOutput, error) {
	uploads, paginator, err := uc.repo.Get(ctx, sc, upload.GetOptions{
		Filter:   ip.Filter,
		PagQuery: ip.PagQuery,
	})
	if err != nil {
		uc.l.Errorf(ctx, "internal.upload.usecase.Get.uc.repo.Get: %v", err)
		return upload.GetOutput{}, err
	}

	return upload.GetOutput{
		Uploads:   uploads,
		Paginator: paginator,
	}, nil
}

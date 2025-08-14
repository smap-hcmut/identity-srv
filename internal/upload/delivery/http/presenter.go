package http

import (
	"encoding/json"
	"mime/multipart"

	"github.com/nguyentantai21042004/smap-api/internal/upload"
	"github.com/nguyentantai21042004/smap-api/pkg/paginator"
)

type createReq struct {
	BucketName string                `form:"bucket_name" binding:"required"`
	FileHeader *multipart.FileHeader `form:"-"` // Will be set manually
}

func (req createReq) toInput() upload.CreateInput {
	return upload.CreateInput{
		FileHeader: req.FileHeader,
		BucketName: req.BucketName,
	}
}

type getReq struct {
	Page          int    `form:"page" binding:"omitempty,min=1"`
	Limit         int64  `form:"limit" binding:"omitempty,min=1,max=100"`
	BucketName    string `form:"bucket_name"`
	OriginalName  string `form:"original_name"`
	Source        string `form:"source"`
	CreatedUserID string `form:"created_user_id"`
}

func (req getReq) toInput() upload.GetInput {
	filter := upload.Filter{}
	if req.BucketName != "" {
		filter.BucketName = &req.BucketName
	}
	if req.OriginalName != "" {
		filter.OriginalName = &req.OriginalName
	}
	if req.Source != "" {
		filter.Source = &req.Source
	}
	if req.CreatedUserID != "" {
		filter.CreatedUserID = &req.CreatedUserID
	}

	return upload.GetInput{
		Filter: filter,
		PagQuery: paginator.PaginateQuery{
			Page:  req.Page,
			Limit: req.Limit,
		},
	}
}

type getUploadResp struct {
	Data []uploadItem                `json:"data"`
	Meta paginator.PaginatorResponse `json:"meta"`
}

type uploadItem struct {
	ID            string `json:"id"`
	BucketName    string `json:"bucket_name"`
	ObjectName    string `json:"object_name"`
	OriginalName  string `json:"original_name"`
	Size          int64  `json:"size"`
	ContentType   string `json:"content_type"`
	Etag          string `json:"etag,omitempty"`
	Metadata      string `json:"metadata,omitempty"`
	URL           string `json:"url,omitempty"`
	Source        string `json:"source"`
	PublicID      string `json:"public_id,omitempty"`
	CreatedUserID string `json:"created_user_id"`
	CreatedAt     string `json:"created_at"`
	UpdatedAt     string `json:"updated_at"`
}

func (h handler) newGetResp(o upload.GetOutput) getUploadResp {
	items := make([]uploadItem, len(o.Uploads))
	for i, upload := range o.Uploads {
		etag := ""
		if upload.Etag != nil {
			etag = *upload.Etag
		}

		metadata := ""
		if upload.Metadata != nil {
			// Convert metadata to JSON string
			if metadataBytes, err := json.Marshal(upload.Metadata); err == nil {
				metadata = string(metadataBytes)
			}
		}

		url := ""
		if upload.URL != nil {
			url = *upload.URL
		}

		publicID := ""
		if upload.PublicID != nil {
			publicID = *upload.PublicID
		}

		items[i] = uploadItem{
			ID:            upload.ID,
			BucketName:    upload.BucketName,
			ObjectName:    upload.ObjectName,
			OriginalName:  upload.OriginalName,
			Size:          upload.Size,
			ContentType:   upload.ContentType,
			Etag:          etag,
			Metadata:      metadata,
			URL:           url,
			Source:        upload.Source,
			PublicID:      publicID,
			CreatedUserID: upload.CreatedUserID,
			CreatedAt:     upload.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt:     upload.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
	}

	return getUploadResp{
		Data: items,
		Meta: paginator.PaginatorResponse{
			Total:      o.Paginator.Total,
			TotalPages: o.Paginator.TotalPages(),
		},
	}
}

func (h handler) newItem(o upload.UploadOutput) uploadItem {
	etag := ""
	if o.Upload.Etag != nil {
		etag = *o.Upload.Etag
	}

	metadata := ""
	if o.Upload.Metadata != nil {
		// Convert metadata to JSON string
		if metadataBytes, err := json.Marshal(o.Upload.Metadata); err == nil {
			metadata = string(metadataBytes)
		}
	}

	url := ""
	if o.Upload.URL != nil {
		url = *o.Upload.URL
	}

	publicID := ""
	if o.Upload.PublicID != nil {
		publicID = *o.Upload.PublicID
	}

	return uploadItem{
		ID:            o.Upload.ID,
		BucketName:    o.Upload.BucketName,
		ObjectName:    o.Upload.ObjectName,
		OriginalName:  o.Upload.OriginalName,
		Size:          o.Upload.Size,
		ContentType:   o.Upload.ContentType,
		Etag:          etag,
		Metadata:      metadata,
		URL:           url,
		Source:        o.Upload.Source,
		PublicID:      publicID,
		CreatedUserID: o.Upload.CreatedUserID,
		CreatedAt:     o.Upload.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:     o.Upload.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

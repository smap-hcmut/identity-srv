package models

import (
	"encoding/json"
	"time"

	"github.com/nguyentantai21042004/smap-api/internal/dbmodels"
)

type Upload struct {
	ID            string                 `json:"id"`
	BucketName    string                 `json:"bucket_name"`
	ObjectName    string                 `json:"object_name"`
	OriginalName  string                 `json:"original_name"`
	Size          int64                  `json:"size"`
	ContentType   string                 `json:"content_type"`
	Etag          *string                `json:"etag,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
	URL           *string                `json:"url,omitempty"`
	Source        string                 `json:"source"`
	PublicID      *string                `json:"public_id,omitempty"`
	CreatedUserID string                 `json:"created_user_id"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
	DeletedAt     *time.Time             `json:"deleted_at,omitempty"`
}

func NewUpload(dbUpload dbmodels.Upload) Upload {
	var metadata map[string]interface{}
	if dbUpload.Metadata.Valid {
		_ = json.Unmarshal(dbUpload.Metadata.JSON, &metadata)
	}

	return Upload{
		ID:            dbUpload.ID,
		BucketName:    dbUpload.BucketName,
		ObjectName:    dbUpload.ObjectName,
		OriginalName:  dbUpload.OriginalName,
		Size:          dbUpload.Size,
		ContentType:   dbUpload.ContentType,
		Etag:          dbUpload.Etag.Ptr(),
		Metadata:      metadata,
		URL:           dbUpload.URL.Ptr(),
		Source:        dbUpload.Source,
		PublicID:      dbUpload.PublicID.Ptr(),
		CreatedUserID: dbUpload.CreatedUserID,
		CreatedAt:     dbUpload.CreatedAt,
		UpdatedAt:     dbUpload.UpdatedAt,
		DeletedAt:     dbUpload.DeletedAt.Ptr(),
	}
}

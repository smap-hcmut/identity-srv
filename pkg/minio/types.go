package minio

import (
	"io"
	"time"
)

type FileInfo struct {
	ID           string            `json:"id"`
	BucketName   string            `json:"bucket_name"`
	ObjectName   string            `json:"object_name"`
	OriginalName string            `json:"original_name"`
	Size         int64             `json:"size"`
	ContentType  string            `json:"content_type"`
	ETag         string            `json:"etag"`
	LastModified time.Time         `json:"last_modified"`
	Metadata     map[string]string `json:"metadata"`
	URL          string            `json:"url,omitempty"`
}

type UploadRequest struct {
	BucketName   string            `json:"bucket_name"`
	ObjectName   string            `json:"object_name"`
	OriginalName string            `json:"original_name"`
	Reader       io.Reader         `json:"-"`
	Size         int64             `json:"size"`
	ContentType  string            `json:"content_type"`
	Metadata     map[string]string `json:"metadata"`
}

type DownloadRequest struct {
	BucketName  string     `json:"bucket_name"`
	ObjectName  string     `json:"object_name"`
	Range       *ByteRange `json:"range,omitempty"`
	Disposition string     `json:"disposition"` // "auto", "inline", "attachment"
}

type ByteRange struct {
	Start int64 `json:"start"`
	End   int64 `json:"end"`
}

type ListRequest struct {
	BucketName string `json:"bucket_name"`
	Prefix     string `json:"prefix"`
	Recursive  bool   `json:"recursive"`
	MaxKeys    int    `json:"max_keys"`
}

type ListResponse struct {
	Files       []*FileInfo `json:"files"`
	IsTruncated bool        `json:"is_truncated"`
	NextMarker  string      `json:"next_marker,omitempty"`
	TotalCount  int         `json:"total_count"`
}

type PresignedURLRequest struct {
	BucketName string            `json:"bucket_name"`
	ObjectName string            `json:"object_name"`
	Method     string            `json:"method"`
	Expiry     time.Duration     `json:"expiry"`
	Headers    map[string]string `json:"headers"`
}

type PresignedURLResponse struct {
	URL       string            `json:"url"`
	ExpiresAt time.Time         `json:"expires_at"`
	Headers   map[string]string `json:"headers,omitempty"`
	Method    string            `json:"method"`
}

type DownloadHeaders struct {
	ContentType        string
	ContentDisposition string
	ContentLength      string
	LastModified       string
	ETag               string
	CacheControl       string
	AcceptRanges       string
	ContentRange       string
}

type BucketInfo struct {
	Name         string    `json:"name"`
	CreationDate time.Time `json:"creation_date"`
	Region       string    `json:"region"`
}

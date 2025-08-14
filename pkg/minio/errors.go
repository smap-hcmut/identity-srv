package minio

import "fmt"

const (
	ErrCodeConnection     = "CONNECTION_ERROR"
	ErrCodeBucketNotFound = "BUCKET_NOT_FOUND"
	ErrCodeObjectNotFound = "OBJECT_NOT_FOUND"
	ErrCodePermission     = "PERMISSION_DENIED"
	ErrCodeInvalidInput   = "INVALID_INPUT"
)

type StorageError struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	Operation string `json:"operation"`
	Cause     error  `json:"-"`
}

func (e *StorageError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s", e.Message, e.Cause.Error())
	}
	return e.Message
}

func (e *StorageError) Unwrap() error {
	return e.Cause
}

func NewConnectionError(err error) *StorageError {
	return &StorageError{Code: ErrCodeConnection, Message: "Storage connection failed", Cause: err}
}

func NewBucketNotFoundError(bucketName string) *StorageError {
	return &StorageError{Code: ErrCodeBucketNotFound, Message: "Bucket not found: " + bucketName}
}

func NewObjectNotFoundError(objectName string) *StorageError {
	return &StorageError{Code: ErrCodeObjectNotFound, Message: "Object not found: " + objectName}
}

func NewInvalidInputError(message string) *StorageError {
	return &StorageError{Code: ErrCodeInvalidInput, Message: message}
}

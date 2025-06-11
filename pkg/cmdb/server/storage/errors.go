package storage

import (
	"fmt"
)

const (
	ErrCodeKeyNotFound int = iota + 1
	ErrCodeKeyExists
	ErrCodeInvalidObj
	ErrCodeResourceReferenced
	ErrCodeReferencedNotExist
)

var errCodeToMessage = map[int]string{
	ErrCodeKeyNotFound:        "key not found",
	ErrCodeKeyExists:          "key exists",
	ErrCodeInvalidObj:         "invalid object",
	ErrCodeResourceReferenced: "resource has been referenced",
	ErrCodeReferencedNotExist: "resource reference targert not exist",
}

func NewKeyNotFoundError(key string, rv int64) *StorageError {
	return &StorageError{
		Code:            ErrCodeKeyNotFound,
		Key:             key,
		ResourceVersion: rv,
	}
}

func NewKeyExistsError(key string, rv int64) *StorageError {
	return &StorageError{
		Code:            ErrCodeKeyExists,
		Key:             key,
		ResourceVersion: rv,
	}
}

func NewInvalidObjError(key, msg string) *StorageError {
	return &StorageError{
		Code:               ErrCodeInvalidObj,
		Key:                key,
		AdditionalErrorMsg: msg,
	}
}

func NewReferencedNotExist(key, msg string) *StorageError {
	return &StorageError{
		Code:               ErrCodeReferencedNotExist,
		Key:                key,
		AdditionalErrorMsg: msg,
	}
}

func NewResourceReferencedError(key, msg string) *StorageError {
	return &StorageError{
		Code:               ErrCodeResourceReferenced,
		Key:                key,
		AdditionalErrorMsg: msg,
	}
}

type StorageError struct {
	Code               int
	Key                string
	ResourceVersion    int64
	AdditionalErrorMsg string
}

func (e *StorageError) Error() string {
	return fmt.Sprintf("StorageError: %s, Code: %d, Key: %s, ResourceVersion: %d, AdditionalErrorMsg: %s",
		errCodeToMessage[e.Code], e.Code, e.Key, e.ResourceVersion, e.AdditionalErrorMsg)
}

// IsNotFound returns true if and only if err is "key" not found error.
func IsNotFound(err error) bool {
	return isErrCode(err, ErrCodeKeyNotFound)
}

// IsExist returns true if and only if err is "key" already exists error.
func IsExist(err error) bool {
	return isErrCode(err, ErrCodeKeyExists)
}

// IsInvalidObj returns true if and only if err is invalid error
func IsInvalidObj(err error) bool {
	return isErrCode(err, ErrCodeInvalidObj)
}

// IsResourceReferenced returns true if resource referenced by others
func IsResourceReferenced(err error) bool {
	return isErrCode(err, ErrCodeResourceReferenced)
}

// IsReferencedNotExist returns true if resource reference target not exist
func IsReferencedNotExist(err error) bool {
	return isErrCode(err, ErrCodeReferencedNotExist)
}

func isErrCode(err error, code int) bool {
	if err == nil {
		return false
	}
	if e, ok := err.(*StorageError); ok {
		return e.Code == code
	}
	return false
}

// InternalError is generated when an error occurs in the storage package, i.e.,
// not from the underlying storage backend (e.g., etcd).
type InternalError struct {
	Reason string
}

func (e InternalError) Error() string {
	return e.Reason
}

// IsInternalError returns true if and only if err is an InternalError.
func IsInternalError(err error) bool {
	_, ok := err.(InternalError)
	return ok
}

func NewInternalError(reason string) InternalError {
	return InternalError{reason}
}

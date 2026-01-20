// Package storage 提供数据存储接口
package storage

import "errors"

var (
	// ErrNotFound 资源未找到
	ErrNotFound = errors.New("resource not found")
	
	// ErrAlreadyExists 资源已存在
	ErrAlreadyExists = errors.New("resource already exists")
	
	// ErrInvalidData 无效数据
	ErrInvalidData = errors.New("invalid data")
)

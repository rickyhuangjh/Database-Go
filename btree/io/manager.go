package io

import (
	"btree/utils"
	"os"
)

const dataFilePath = "data.bin"

type StorageManager interface {
	Read(pageIdx int) ([]byte, error)
	Write(pageIdx int, data []byte) error
}

type storageManager struct {
	file *os.File
	pageSize int
}


func NewStorageManager() *storageManager {
	file, err := os.OpenFile(dataFilePath, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		panic(err)
	}
	return &storageManager{file: file, pageSize: utils.PageSize}
}

func (s *storageManager) Read(pageIdx int) ([]byte, error) {
	offset := pageIdx * s.pageSize
	buf := make([]byte, s.pageSize)
	_, err := s.file.ReadAt(buf, int64(offset))
	if err != nil {
		return nil, err
	}
	return buf, nil
}

func (s *storageManager) Write(pageIdx int, data []byte) error {
	offset := pageIdx * s.pageSize
	_, err := s.file.WriteAt(data, int64(offset))
	return err
}








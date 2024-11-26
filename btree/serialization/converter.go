package serialization

import (
	"btree/node"
	"btree/utils"
	"os"
)

const dataFilePath = "data.bin"

type NodeStorageManager interface {
	Read(pageIdx int) (node.BTreeNodeSerialized, error)
	Write(pageIdx int, data node.BTreeNodeSerialized) error
}

type NodeStorageManagerImpl struct {
	file *os.File
	pageSize int
}


func NewNodeStorageManager() *NodeStorageManagerImpl {
	file, err := os.OpenFile(dataFilePath, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		panic(err)
	}
	return &NodeStorageManagerImpl{file: file, pageSize: utils.PageSize}
}

func (s *NodeStorageManagerImpl) Read(pageIdx int) ([]byte, error) {
	offset := pageIdx * s.pageSize
	buf := make([]byte, s.pageSize)
	_, err := s.file.ReadAt(buf, int64(offset))
	if err != nil {
		return nil, err
	}
	return buf, nil
}

func (s *NodeStorageManagerImpl) Write(pageIdx int, data []byte) error {
	offset := pageIdx * s.pageSize
	_, err := s.file.WriteAt(data, int64(offset))
	return err
}





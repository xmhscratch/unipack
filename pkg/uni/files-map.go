package uni

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
)

func (m *FilesMap) Exist(ino uint64) bool {
	_, ok := m.TFilesMap[ino]
	return ok
}

func (m *FilesMap) Get(ino uint64) *FileInode {
	if m.Exist(ino) {
		return m.TFilesMap[ino]
	}
	return nil
}

func (m *FilesMap) Set(ino uint64, fi *FileInode) {
	// fmt.Println(ino, fi.FilePath)
	m.TFilesMap[ino] = fi
}

func (m *FilesMap) GetSaveName() string {
	return "unipack." + strconv.FormatUint(m.uint64, 10)
}

func (m *FilesMap) Populate() (isCreated bool, err error) {
	savePath := filepath.Join(os.TempDir(), m.GetSaveName())

	saveFile, err := os.OpenFile(savePath, os.O_CREATE|os.O_RDONLY, 0644)
	if err != nil {
		panic(err)
	}
	defer saveFile.Close()

	fmt.Println("open:" + saveFile.Name())

	var buf *bytes.Buffer = bytes.NewBuffer(make([]byte, 0))
	_, err = io.CopyBuffer(buf, saveFile, make([]byte, 1024))
	if err != nil {
		panic(err)
	}

	if buf.Len() != 0 {
		dec := gob.NewDecoder(buf)
		err = dec.Decode(&m)
		if err != nil {
			panic(err)
		}
		// fmt.Printf("%+v\n", m)
	} else {
		isCreated = true
	}
	return isCreated, err
}

func (m *FilesMap) Save() error {
	savePath := filepath.Join(os.TempDir(), m.GetSaveName())

	saveFile, err := os.OpenFile(savePath, os.O_TRUNC|os.O_RDWR, 0644)
	if err != nil {
		panic(err)
	}
	defer saveFile.Close()

	fmt.Println("save:" + saveFile.Name())

	var buf *bytes.Buffer = bytes.NewBuffer(make([]byte, 0))
	enc := gob.NewEncoder(buf)
	err = enc.Encode(m)
	if err != nil {
		panic(err)
	}

	_, err = io.CopyBuffer(saveFile, buf, make([]byte, 1024))
	return err
}

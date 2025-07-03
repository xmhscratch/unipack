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

func (m TFilesMap) Exist(ino uint64) bool {
	_, ok := m[ino]
	return ok
}

func (m TFilesMap) Get(ino uint64) *FileInode {
	if m.Exist(ino) {
		return m[ino]
	}
	return nil
}

func (m TFilesMap) Set(ino uint64, fi *FileInode) {
	m[ino] = fi
}

// ===============================================

func (m *FilesMap) Exist(ino uint64) bool {
	return m.TFilesMap.Exist(ino)
}

func (m *FilesMap) Get(ino uint64) *FileInode {
	return m.TFilesMap.Get(ino)
}

func (m *FilesMap) Set(ino uint64, fi *FileInode) {
	m.TFilesMap.Set(ino, fi)
}

func (m *FilesMap) GetSaveName() string {
	return "unipack." + strconv.FormatUint(m.uint64, 10)
}

func (m *FilesMap) Load() (isNew bool, err error) {
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
		// fmt.Printf("%+v\n", string(m.TFilesMap[9223372036854775808].Content))
	} else {
		isNew = true
	}
	return isNew, err
}

func (m *FilesMap) Save() error {
	savePath := filepath.Join(os.TempDir(), m.GetSaveName())

	saveFile, err := os.OpenFile(savePath, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0644)
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

	// fmt.Printf("%+v\n", string(m.TFilesMap[9223372036854775808].Content))

	_, err = io.CopyBuffer(saveFile, buf, make([]byte, 1024))
	return err
}

// map[0:0xc000078e00 9223372036854775808:0xc000078dc0]
// map[0:0xc0005ee740 9223372036854775808:0xc0005ee700]

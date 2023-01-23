package main

import (
	"io/fs"
	"net/http"
	"path/filepath"
)

type FileSystemNoDirList struct {
	fs            http.FileSystem
	IndexFilename string
}

func CreateFileSystemNoDirList(dir http.Dir, indexFilename string) FileSystemNoDirList {
	return FileSystemNoDirList{
		fs:            http.FileSystem(dir),
		IndexFilename: indexFilename,
	}
}

func (fsNoDir FileSystemNoDirList) Open(path string) (http.File, error) {
	var err error
	var f http.File

	f, err = fsNoDir.fs.Open(path)
	if err != nil {
		return nil, err
	}

	var s fs.FileInfo
	s, err = f.Stat()
	if err != nil {
		return nil, err
	}
	if s.IsDir() {
		index := filepath.Join(path, fsNoDir.IndexFilename)
		if _, err := fsNoDir.fs.Open(index); err != nil {
			closeErr := f.Close()
			if closeErr != nil {
				return nil, closeErr
			}

			return nil, err
		}
	}

	return f, nil
}

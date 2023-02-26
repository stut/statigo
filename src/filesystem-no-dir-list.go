package main

import (
	"io/fs"
	"net/http"
	"path/filepath"
	"strings"
)

type FileSystemNoDirList struct {
	fs            http.FileSystem
	IndexFilename string
	ServeHidden   bool
}

func CreateFileSystemNoDirList(dir http.Dir, indexFilename string, serveHidden bool) FileSystemNoDirList {
	return FileSystemNoDirList{
		fs:            http.FileSystem(dir),
		IndexFilename: indexFilename,
		ServeHidden:   serveHidden,
	}
}

func (fsNoDir FileSystemNoDirList) Open(path string) (http.File, error) {
	var err error
	var f http.File

	if !fsNoDir.ServeHidden && strings.Contains(path, "/.") {
		return nil, fs.ErrNotExist
	}

	f, err = fsNoDir.fs.Open(path)
	if err != nil {
		return nil, fs.ErrNotExist
	}

	var s fs.FileInfo
	s, err = f.Stat()
	if err != nil {
		return nil, fs.ErrNotExist
	}
	if s.IsDir() {
		index := filepath.Join(path, fsNoDir.IndexFilename)
		if _, err := fsNoDir.fs.Open(index); err != nil {
			closeErr := f.Close()
			if closeErr != nil {
				return nil, fs.ErrNotExist
			}

			return nil, fs.ErrNotExist
		}
	}

	return f, nil
}

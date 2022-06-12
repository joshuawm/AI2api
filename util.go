package main

import (
	"encoding/json"
	"errors"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

//just utility
func SplitPathIntoPaths(path string) []string {
	var Separator string
	if runtime.GOOS == "windows" {
		Separator = "\\"
	} else {
		Separator = "/"
	}
	var absolutePath bool = false
	if path[0:1] == Separator {
		absolutePath = true
	}
	Paths := strings.Split(path, Separator)
	if absolutePath {
		Paths[0] = Separator + Paths[0]
	}
	return Paths
}

func getFilesList(rw http.ResponseWriter, r *http.Request) {
	folderName := r.URL.Query().Get("folder")
	fEn, err := os.ReadDir(folderName)
	if err != nil {
		internalErrorResponse(err, &rw)
		return
	}
	temp := []string{}
	for _, v := range fEn {
		if v.IsDir() {
			continue
		} else {
			temp = append(temp, v.Name())
		}
	}
	by, _ := json.Marshal(Response{Status: true, Content: temp, Err: ""})
	rw.Write(by)
	return
}

func download(rw http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")
	if path == "" {
		回去大礼包("没写path路径，回去重写", &rw)
		return
	}
	var paths []string

	json.Unmarshal([]byte(path), &paths)
	p := filepath.Join(paths...)
	content, err := os.ReadFile(p)
	if err != nil {
		internalErrorResponse(err, &rw)
		return
	}

	mType := mime.TypeByExtension(filepath.Ext(p))
	rw.Header().Set("Content-Type", mType)
	rw.Header().Set("Content-Disposition", `attachment; filename="`+filepath.Base(p)+`"`)
	rw.Write(content)
}

func fileNameWithoutExtSliceNotation(fileName string) string {
	return fileName[:len(fileName)-len(filepath.Ext(fileName))]
}

func rawSizeString2Size(rawString string) ([2]uint, error) {
	// var size [2]uint = nil
	size := [2]uint{0, 0}
	// rawSizeString:= r.URL.Query().Get("size")
	if rawString != "" {
		sizeSplited := strings.Split(rawString, ",")
		if len(sizeSplited) == 2 {

			tempi, err := strconv.ParseUint(sizeSplited[0], 0, 32)
			if err != nil {
				return size, err
			}
			size[0] = uint(tempi)
			tempj, err := strconv.ParseUint(sizeSplited[0], 0, 32)
			if err != nil {
				return size, err
			}
			size[1] = uint(tempj)
			return size, nil
		} else {
			return nil, errors.New("length != 2")
		}

	}
}

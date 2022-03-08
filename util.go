package main

import (
	"encoding/json"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
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

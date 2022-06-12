package main

import (
	"log"
	"net/http"
)

func TestImageCrop(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(10 << 10)
	for _, file := range r.MultipartForm.File {
		f, e := file[0].Open()
		if e != nil {
			log.Panic(e)
		}
		CropImageMan(f, [2]uint{100, 100}, fileNameWithoutExtSliceNotation(file[0].Filename))
	}
}

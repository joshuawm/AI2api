package main

import (
	"encoding/json"
	"net/http"
	"os"
)

func getAllCfg(rw http.ResponseWriter, r *http.Request) {
	fEn, err := os.ReadDir("AerialDetection/configs/DOTA")
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

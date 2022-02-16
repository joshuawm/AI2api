package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

type TrainConfig struct {
	// your train files. like  trqain.py yolo/train.py
	TrainFile  string   `json:"trainFile"`
	TrainExtra []string `json:"trainExtra"`
}

var (
	cons     map[string]*websocket.Conn = make(map[string]*websocket.Conn)
	upgrader websocket.Upgrader         = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     func(r *http.Request) bool { return true },
	}
)

func train(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	trainName := vars["name"]
	if trainName == "" {
		回去大礼包("train name 不存在", &w)
		return
	}
	//get
	trainConfig := config.Trains[trainName]

	var config []string
	var configString string

	if configString == "" {
		configString = r.FormValue("config")
	}
	if configString == "" {
		config = []string{}
	}

	// get all varaille
	if configString != "" {
		err := json.Unmarshal([]byte(configString), &config)
		if err != nil {
			internalErrorResponse(err, &w)
			return
		}
	} else {
		config = []string{}
	}

	con := cons[trainName]
	if con == nil {
		回去大礼包("没有ws连接", &w)
		return
	}

	go func(trainFile TrainConfig, extraConfig []string, con *websocket.Conn) {
		commandS := append([]string{trainFile.TrainFile}, extraConfig...)
		//redirect stderr and stdout to a file and make it run
		// commandS = append(commandS, " >- 1>&2 | tail -f -")

		//combine stdout and stderr together
		//https://stackoverflow.com/questions/35994907/go-combining-cmd-stdoutpipe-and-cmd-stderrpipe
		cmd := exec.Command("python", commandS...)
		stdout, err := cmd.StdoutPipe()
		cmd.Stderr = cmd.Stdout
		if err != nil {
			con.WriteMessage(websocket.TextMessage, []byte("训练失败"))
			con.WriteMessage(websocket.TextMessage, []byte(err.Error()))
			return
		}
		if e := cmd.Start(); e != nil {
			con.WriteMessage(websocket.TextMessage, []byte("训练失败"))
			con.WriteMessage(websocket.TextMessage, []byte(e.Error()))
			return
		}
		read := bufio.NewReader(stdout)
		line, _, err := read.ReadLine()
		// for {
		// 	line, _, err := read.ReadLine()
		// 	if err != nil {
		// 		fmt.Print("err")
		// 		log.Fatal(err)
		// 	}
		// 	fmt.Print(string(line))
		// }
		for err == nil {
			con.WriteMessage(websocket.TextMessage, line)
			fmt.Print(string(line))
			line, _, err = read.ReadLine()
		}
		con.WriteMessage(websocket.TextMessage, []byte("train is ended"))
	}(trainConfig, config, con)
	t, _ := json.Marshal(Response{Status: 1, Content: []string{"success! train is starting"}, Err: ""})
	w.Write([]byte(t))
}

func ws(w http.ResponseWriter, r *http.Request) {
	name := mux.Vars(r)["room"]
	if name == "" {
		return
	}
	con, err := upgrader.Upgrade(w, r, nil)
	cons[name] = con
	if err != nil {
		return
	}
	for {
		msgType, msg, err := con.ReadMessage()
		if err != nil {
			continue
		}
		if msgType == websocket.TextMessage {
			fmt.Print(string(msg))
		}
	}

}

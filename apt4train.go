package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"github.com/gorilla/websocket"
	"github.com/gorilla/mux"
)

type TrainConfig struct{
	// your train files
	TrainFile string  `json:"trainFile"`


}

var (
	con websocket.Conn
	upgrader websocket.Upgrader = websocket.Upgrader{
		ReadBufferSize: 1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {return true},
	}
	
)

func train(w http.ResponseWriter,r *http.Request){
	vars := mux.Vars(r)
	trainName := vars["name"]
	if trainName==""{
		回去大礼包("train name 不存在",&w)
		return
	}
	trainConfig := config.Trains[trainName]

	var config []string
	var configString string
	configString=r.URL.Query().Get("config")
	if configString==""{
		configString=r.FormValue("config")
	}
	// if configString==""{
	// 	回去大礼包("config string不存在",&w)
	// 	return
	// }
	if configString!=""{
		err :=json.Unmarshal([]byte(configString),&config)
		if err!=nil{
			internalErrorResponse(err,&w)
			return
		}
	}else{
		config=[]string{}
	}

	
	go func (trainFile TrainConfig,extraConfig []string){
		commandS := append([]string{trainFile.TrainFile},config...)
		cmd:=exec.Command("python",commandS...)
		stdout,err:=cmd.StdoutPipe()
		if(err!=nil){
			con.WriteMessage(websocket.TextMessage,[]byte("训练失败"))
			con.WriteMessage(websocket.TextMessage,[]byte(err.Error()))
			return 
		}
		if e:=cmd.Start();e!=nil{
			con.WriteMessage(websocket.TextMessage,[]byte("训练失败"))
			con.WriteMessage(websocket.TextMessage,[]byte(e.Error()))	
			return	
		}
		read :=bufio.NewReader(stdout)
		line,_,err:=read.ReadLine()
		for err!=nil{
			con.WriteMessage(websocket.TextMessage,line)
			line,_,err=read.ReadLine()
		}
		con.WriteMessage(websocket.TextMessage,[]byte("train is ended"))
		return
	}(trainConfig,config)

	
	
}

func ws(w http.ResponseWriter,r *http.Request){
	con,err  := upgrader.Upgrade(w,r,nil)
	if err !=nil{
		return
	}
	for{
		msgType , msg ,err:=con.ReadMessage()
		if(err!=nil){
			continue
		}
		if(msgType==websocket.TextMessage){
			fmt.Print(string(msg))
		}
	}

}
package main

import (
	// "bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/gorilla/mux"
)

type Config struct{
	Paths map[string]path `json:"paths"`
	Trains map[string]TrainConfig `json:"trains"`
}

type path struct{
	Path string `json:"path"`
	//model的主目录
	ModelPath string `json:"modelPath"`
	//where you save all posted data
	SavePath string `json:"savePath"`
	//OutputPath
	OutPath string `json:"outPath"`
	//save as current data filename
	Time bool `json:"time"`
	//file keys to retrive files from request
	Files []string `json:"files"`
	//to determine whether to active the config file mode
	ConfigFiles bool `json:"configFiles"`
	//the same
	ConfigFile string `json:"configFile"`
	ConfigContent []string `json:"configContent"`
	RunCommand []string `json:"commands"`
}

var config Config

type Response struct{
	Status int8 `json:"status"`
	Content []string `json:"content"`
	Err string `json:"err"`
}

type Env struct{
	//model的主目录
	ModelPath string 
	//where you save all posted data
	SavePath string 	
	//OutputPath
	OutPath string 	
	OutputPath string
	SaveFullPath string
	FolderName string
	ConfigFullFile string
}

func configWriter(templateFilePath string,targetFilePath string,content []string) (bool,error){
	temContent,err :=os.ReadFile(templateFilePath)
	if err!=nil{
		fmt.Print(err.Error())
		return false,err
	}
	file,err :=os.OpenFile(targetFilePath,os.O_CREATE|os.O_WRONLY,0777)
	if err!=nil {
		return false,err
	}
	file.Write(temContent)
	for _,v:=range content{
		file.WriteString(v+"\n")
	}
	return true,nil
}

func predict(w http.ResponseWriter,r *http.Request){ //Method:POST
	vars:=mux.Vars(r)
	modelName := vars["name"]
	if modelName==""{
		//妈呀，真的可以用中文做函数名
		回去大礼包("没有传入模型ID",&w)
		return
	}
	path,ok := config.Paths[modelName]
	if !ok {
		回去大礼包("传入的模型ID不存在",&w)
		return
	}
	//start !
	//check whther the time is enabled
	var folderName string
	if path.Time {
		t := time.Now()
		tim := t.Format("2020-10-01 12:23:12")
		tim = strings.Replace(tim," ","",-1)
		tim = strings.Replace(tim,":","",-1)
		folderName = strings.Replace(tim,"-","",-1)
	}else{
		folderName= r.URL.Query().Get("name")
		if folderName == ""{
			回去大礼包("time设置为false，但是没有传入保存文件名字",&w)
			return	
		}
	}
	folderPath:=filepath.Join(path.ModelPath,path.SavePath,folderName)
	//原先就存在这个文件夹，可能里面有内容，先删除
	if _,ee := os.Stat(folderPath);errors.Is(ee,os.ErrExist){
		os.RemoveAll(folderPath)
	}
	os.MkdirAll(folderPath,os.ModePerm)

	//env
	env := Env{
		ModelPath: path.ModelPath,
		SavePath: path.SavePath,
		OutPath: path.OutPath,

		OutputPath: filepath.Join(path.ModelPath,path.OutPath,folderName),
		SaveFullPath: filepath.Join(path.ModelPath,path.SavePath,folderName),
		FolderName: folderName,
	}

	//save all files
	r.ParseMultipartForm(10<<10)
	filesKey:=path.Files
	for _,key := range filesKey{
		Files,ok := r.MultipartForm.File[key]
		if !ok{
			回去大礼包("在配置文件中写的Key在实际中并未找到",&w)
			return
		}
		for _,v := range Files{
			f,err:=v.Open()
			if err!=nil {
			}
			ff,err:=os.OpenFile(filepath.Join(env.SaveFullPath,v.Filename),os.O_CREATE|os.O_WRONLY,0777)
			io.Copy(ff,f)
			ff.Close()
			f.Close()
		}
	}

	//check config writer
	if (path.ConfigFiles){
		contents:=path.ConfigContent
		if len(contents)==0 {
			回去大礼包("定义了config文件，但是没有书写内容",&w)
			return
		}
		//content to 
		readyContent,err:=  stringSubstituting(contents,env)
		if err!=nil {
			internalErrorResponse(err,&w)
			return
		}
		//filepath.Base to get filename
		env.ConfigFullFile=filepath.Join(env.SaveFullPath,filepath.Base(path.ConfigFile))
		ok,err:=configWriter(path.ConfigFile,env.ConfigFullFile,readyContent)
		if !ok {
			internalErrorResponse(err,&w)
			return
		}
	}
	
	//run command
	//string substituting
	//https://www.socketloop.com/tutorials/golang-interpolating-or-substituting-variables-in-string-examples
	readyCommand ,err:= stringSubstituting(path.RunCommand,env)
	if err!=nil{
		internalErrorResponse(err,&w)
		return
	}
	comhandler :=exec.Command("python3",readyCommand...)
	logOut,err :=comhandler.CombinedOutput()
	if err!=nil{
		internalErrorResponse(err,&w)
		return
	}
	var logFullPath string = filepath.Join(env.SaveFullPath,"log.txt")
	logF,err := os.Create(logFullPath)
	if err!=nil{
		internalErrorResponse(err,&w)
		return
	}
	logF.Write(logOut)
	logF.Close()

	//display all output files
	fileEntryy, err := os.ReadDir(env.OutputPath)
	if err!=nil {
		internalErrorResponse(err,&w)
		return
	}
	outFiles := []string{}
	for _,fEntry := range fileEntryy{
		if  fEntry.IsDir(){
			

		}else{
			outFiles = append(outFiles, fEntry.Name())
		}
	}

}

func filesRecursive(dirPath string,root []string)[]string{
	if len(root)==0{
		//this is the first time, so initialize
		root = filepath.SplitList(dirPath)
	}
	result := []string{}
	fEntery,err := os.ReadDir(dirPath)
	if err!=nil {
		return []string{}
	}
	for _,fe := range fEntery{
		if fe.IsDir()  {
			root=append(root, fe.Name())
			
			resultTemp:=filesRecursive(filepath.Join(root...),root)
			result=append(result, resultTemp...)
		}else{
			t := append(root,fe.Name())
			newCon ,err := json.Marshal(t)
			if err!=nil {
				fmt.Print(" %s 在序列化出错，已跳过",dirPath)
			}else{
				result=append(result, string(newCon))
			}
		}
	}
	return result
}

func stringSubstituting(contentArr []string,substitution interface{})([]string,error){
	temp := template.New("sysEnv")
	var delim string = "[delim]"
	readystr := strings.Join(contentArr,delim)
	templatehandler ,err  := temp.Parse(readystr)
	if err!=nil {
		return []string{},err
	}
	var resulttemp string =""
	resultWriter := bytes.NewBufferString(resulttemp)
	err = templatehandler.Execute(resultWriter,substitution)
	if err!=nil {
		return []string{},err
	}
	result := strings.Split(resultWriter.String(),delim)
	return result,nil
}

func internalErrorResponse(err error,w *http.ResponseWriter){
	wpu := *w
	wpu.WriteHeader(http.StatusInternalServerError)
	wpu.Write([]byte(err.Error()))
}

func 回去大礼包(why string,w *http.ResponseWriter){
	wPupet := *w
	c:=Response{Status: 0,Err: why}
	content ,err := json.Marshal(c)
	if err!=nil{	
		wPupet.WriteHeader(http.StatusInternalServerError)
	}else{
		wPupet.WriteHeader(http.StatusBadRequest)
		wPupet.Write(content)
	}
}

func main(){
	f,err:=os.ReadFile("config.json")
	if err!=nil {
		fmt.Print(err.Error())
		os.Exit(0)
	}
	err = json.Unmarshal(f,&config)
	if err!=nil{
		fmt.Print(err.Error())
		os.Exit(0)
	}
	fmt.Print(config)
	r:=mux.NewRouter()
	r.HandleFunc("/predict/{name}",predict)
	r.HandleFunc("/train/{name}",train)
	r.HandleFunc("/ws/{room}",ws)
	http.Handle("/",r)
	http.ListenAndServe(":8080",nil)

}
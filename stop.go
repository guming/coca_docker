package main

import (
	"syscall"
	"strconv"
	"fmt"
	"encoding/json"
	"github.com/coca_docker/container"
	"io/ioutil"
	log "github.com/sirupsen/logrus"
)

func stopContainer (containerName string) {
	//kill pid
	pid,err:=getContainerPidByName(containerName)
	if err!=nil{
		log.Errorf("get pid error %s and err is %v",containerName,err)
		return
	}
	npid,err:=strconv.Atoi(pid)
	if err!=nil{
		log.Errorf("conv pid error %v",err)
		return
	}
	if err:=syscall.Kill(npid,syscall.SIGTERM);err!=nil{
		log.Errorf("kill pid error %v",err)
		return
	}

	//update config.json
	cinfo,err:=getContainerInfoByName(containerName)
	if err != nil {
		log.Errorf("json marshal %s error %v", containerName, err)
		return
	}

	cinfo.Pid=" "
	cinfo.Status=container.STOP
	newContentBytes, err := json.Marshal(cinfo)
	dirURL := fmt.Sprintf(container.DefaultInfoLocation, containerName)
	configFilePath := dirURL + container.ConfigName
	if err:=ioutil.WriteFile(configFilePath,newContentBytes,0622);err!=nil{
		log.Errorf("write json %s error %v", containerName, err)
		return
	}
}

func getContainerInfoByName(containerName string) (*container.ContainerInfo, error) {
	configFileDir := fmt.Sprintf(container.DefaultInfoLocation, containerName)
	configFileDir = configFileDir + container.ConfigName
	content,err:=ioutil.ReadFile(configFileDir)
	if err != nil {
		log.Errorf("read file %s error %v", configFileDir, err)
		return nil, err
	}
	var containerInfo container.ContainerInfo
	if err := json.Unmarshal(content, &containerInfo); err != nil {
		log.Errorf("json unmarshal error %v", err)
		return nil, err
	}

	return &containerInfo, nil
}
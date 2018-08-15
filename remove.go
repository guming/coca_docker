package main

import (
	log "github.com/sirupsen/logrus"
	"github.com/coca_docker/container"
	"os"
	"fmt"
)
func removeContainer (containerName string) {
	cinfo,err:=getContainerInfoByName(containerName)
	if err != nil {
		log.Errorf("json marshal %s error %v", containerName, err)
		return
	}
	if cinfo.Status!=container.STOP {
		log.Errorf("container still running %v", containerName, err)
		return
	}
	dirURL := fmt.Sprintf(container.DefaultInfoLocation, containerName)
	if err:=os.RemoveAll(dirURL);err!=nil{
		log.Errorf("remove dirURL %s error %v", dirURL, err)
		return
	}
	container.DeleteWorkSpace(containerName,cinfo.Volume)
}

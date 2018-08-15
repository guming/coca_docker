package main

import (
	"os/exec"
	log "github.com/sirupsen/logrus"
	"fmt"
	"github.com/coca_docker/container"
)

func commitContainer(imageName string,containerName string){

	mntURL:=fmt.Sprintf(container.MntURL, containerName)
	imageTar:=container.RootURL+"/"+imageName+".tar"
	log.Infof("image tar is %s",imageTar)

	if _,err:=exec.Command("tar","-czf",imageTar,"-C",mntURL,".").CombinedOutput();err!=nil{
		log.Errorf("commit container err %v",err)
	}

}
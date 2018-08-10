package main

import (
	"os/exec"
	log "github.com/sirupsen/logrus"
)

func commitContainer(imageName string){

	mntURL:="/root/mnt"
	imageTar:=mntURL+imageName+".tar"
	log.Infof("image tar is %s",imageTar)

	if _,err:=exec.Command("tar","-czf",imageTar,"-C",mntURL,".").CombinedOutput();err!=nil{
		log.Errorf("commit container err %v",err)
	}

}
package main

import (
	"github.com/coca_docker/container"
	log "github.com/sirupsen/logrus"
	"os"
)

func Run(command string,tty bool){
	process:=container.NewParentProcess(tty,command)
	err:=process.Start()
	if err!=nil{
		log.Errorf("container start err %",err.Error())
	}
	process.Wait()
	os.Exit(-1)
}

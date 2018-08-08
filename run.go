package main

import (
	"github.com/coca_docker/container"
	log "github.com/sirupsen/logrus"
	"os"
	"github.com/coca_docker/cgroup/subsystems"
	"strings"
)

func Run(command []string,tty bool,config *subsystems.ResourceConfig){
	process,writepip:=container.NewParentProcess(tty)
	err:=process.Start()
	if err!=nil{
		log.Errorf("container start err %v",err)
	}
	//cgroup
	//cgroupManager:=cgroup.NewCgroupManager("coca-docker")
	//defer cgroupManager.Destory()
	//cgroupManager.Set(config)
	//cgroupManager.Apply(process.Process.Pid)
	sendCommandToChild(writepip,command)
	process.Wait()
	os.Exit(-1)
}

func sendCommandToChild(pip *os.File,command []string){
	cmd:=strings.Join(command," ")
	log.Infof("cmd string is %s",cmd)
	pip.WriteString(cmd)
	pip.Close()
}

package main

import (
	"fmt"
	"encoding/json"
	"github.com/coca_docker/container"
	"io/ioutil"
	log "github.com/sirupsen/logrus"
	"strings"
	"os/exec"
	"os"
)

func execContainer (containerName string,command []string){

	pid,err:=getContainerPidByName(containerName)
	if err!=nil{
		log.Errorf("get cid error %v",err)
		return
	}

	commandstr:=strings.Join(command," ")
	log.Infof("container pid %s", pid)
	log.Infof("command %s", commandstr)


	cmd:=exec.Command("/proc/self/exe","exec")
	cmd.Stderr=os.Stderr
	cmd.Stdout=os.Stdout
	cmd.Stdin=os.Stdin

	os.Setenv(ENV_EXEC_CMD,commandstr)
	os.Setenv(ENV_EXEC_PID,pid)

	if err:=cmd.Run();err!=nil {
		log.Errorf("exec failed %v",err)
	}
}


func getContainerPidByName(containerName string) (string, error) {
	dirURL := fmt.Sprintf(container.DefaultInfoLocation, containerName)
	configFilePath := dirURL + container.ConfigName
	contentBytes, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		return "", err
	}
	var containerInfo container.ContainerInfo
	if err := json.Unmarshal(contentBytes, &containerInfo); err != nil {
		return "", err
	}
	return containerInfo.Pid, nil
}
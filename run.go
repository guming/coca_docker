package main

import (
	"github.com/coca_docker/container"
	log "github.com/sirupsen/logrus"
	"os"
	"github.com/coca_docker/cgroup/subsystems"
	"strings"
	"math/rand"
	"time"
	"encoding/json"
	"fmt"
	"strconv"
	"github.com/coca_docker/cgroup"
	"github.com/coca_docker/network"
)

func Run(command []string,tty bool,config *subsystems.ResourceConfig,volume string,
	detach bool,containerName string,imageName string,envSlice []string,nw string,portmapping []string){
	containerID := randStringBytes(10)
	if containerName == "" {
		containerName = containerID
	}
	process,writepip:=container.NewParentProcess(tty,volume,containerName,imageName,envSlice)
	err:=process.Start()
	if err!=nil{
		log.Errorf("container start err %v",err)
	}
	//record container into config.json
	containerName,err=recordContainerInfo(process.Process.Pid,containerName,command,containerID,imageName,volume)
	if err!=nil{
		log.Errorf("recordContainerInfo error %v",err)
		return
	}

	//cgroup
	cgroupManager:=cgroup.NewCgroupManager("coca-docker")
	defer cgroupManager.Destory()
	cgroupManager.Set(config)
	cgroupManager.Apply(process.Process.Pid)

	if nw!=""{
		network.Init()
		cinfo:=&container.ContainerInfo{
			Pid:strconv.Itoa(process.Process.Pid),
			Name:containerName,
			PortMapping:portmapping,
			Id:containerID,
		}
		if err:=network.Connect(nw,cinfo);err!=nil {
			log.Errorf("run network connect error %v",err)
			return
		}
	}

	sendCommandToChild(writepip,command)

	if tty {
		process.Wait()
		container.DeleteWorkSpace(containerName,volume)
		deleteContainerInfo(containerName)
	}
	//os.Exit(-1)
}

func sendCommandToChild(pip *os.File,command []string){
	cmd:=strings.Join(command," ")
	log.Infof("cmd string is %s",cmd)
	pip.WriteString(cmd)
	pip.Close()
}

func recordContainerInfo(containerPID int,cname string,command []string,containerId string,
	imageName string,volume string) (string,error){
	//cid:=randStringBytes(10)
	creatTime:=time.Now().Format("2016-01-02 08:05:45")
	containerInfo:=&container.ContainerInfo{
		Pid:strconv.Itoa(containerPID),
		Id:containerId,
		Name:cname,
		CreatedTime:creatTime,
		Status:container.RUNNING,
		Command:strings.Join(command," "),
		ImageName:imageName,
		Volume:volume,
	}
	jsonBytes,err:=json.Marshal(containerInfo)
	if err!=nil{
		log.Errorf("json record error %v",err)
		return "",err
	}
	jsonvalue:=string(jsonBytes)
	confdir:=fmt.Sprintf(container.DefaultInfoLocation,cname)

	if err:=os.MkdirAll(confdir,0622);err!=nil{
		log.Errorf("confdir mkfir error %v",err)
		return "",err
	}
	file,err:=os.Create(confdir+"/"+container.ConfigName)
	if err!=nil{
		log.Errorf("config.json create error %v",err)
		return "",err
	}
	defer file.Close()
	if _,err:=file.WriteString(jsonvalue);err!=nil{
		log.Errorf("write json error %v",err)
		return "",err
	}
	return cname,nil

}

func randStringBytes(n int) string {
	letterstr:="1234567890"
	rand.Seed(time.Now().UnixNano())
	b:=make([]byte,n)
	for i := range b{
		b[i]=letterstr[rand.Intn(len(letterstr))]
	}
	return string(b)
}

func deleteContainerInfo(containerId string) {
	dirURL := fmt.Sprintf(container.DefaultInfoLocation, containerId)
	if err := os.RemoveAll(dirURL); err != nil {
		log.Errorf("Remove dir %s error %v", dirURL, err)
	}
}
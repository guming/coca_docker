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
)

func Run(command []string,tty bool,config *subsystems.ResourceConfig,volume string,detach bool,name string){
	process,writepip:=container.NewParentProcess(tty,volume)
	err:=process.Start()
	if err!=nil{
		log.Errorf("container start err %v",err)
	}
	//cgroup
	cgroupManager:=cgroup.NewCgroupManager("coca-docker")
	defer cgroupManager.Destory()
	cgroupManager.Set(config)
	cgroupManager.Apply(process.Process.Pid)
	sendCommandToChild(writepip,command)
	//record container into config.json
	containerName,err:=recordContainerInfo(process.Process.Pid,name,command)
	if err!=nil{
		log.Errorf("recordContainerInfo error %v",err)
		return
	}
	if tty {
		process.Wait()
		//mntURL := "/root/mnt/"
		//rootURL := "/root/"
		//container.DeleteWorkSpace(rootURL,mntURL,volume)
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

func recordContainerInfo(containerPID int,cname string,command []string) (string,error){
	cid:=randStringBytes(10)
	creatTime:=time.Now().Format("2000-01-01 01:01:45")
	if cname==""{
		cname=cid
	}
	containerInfo:=&container.ContainerInfo{
		Pid:strconv.Itoa(containerPID),
		Id:cid,
		Name:cname,
		CreatedTime:creatTime,
		Status:container.RUNNING,
		Command:strings.Join(command," "),
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
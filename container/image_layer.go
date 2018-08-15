package container

import (
	"os"
	"os/exec"
	log "github.com/sirupsen/logrus"
	"fmt"
)


func CreateReadOnlyLayer(imageName string) {
	busyboxURL := RootURL + "/" +imageName +"/"
	busyboxTarURL := RootURL + "/" +imageName + ".tar"
	exist, err := PathExists(busyboxURL)
	if err != nil {
		log.Infof("fail to judge whether dir %s exists. %v", busyboxURL, err)
	}
	if exist == false {
		if err := os.MkdirAll(busyboxURL, 0777); err != nil {
			log.Errorf("mkdir dir %s error. %v", busyboxURL, err)
		}
		if _, err := exec.Command("tar", "-xvf", busyboxTarURL, "-C", busyboxURL).CombinedOutput(); err != nil {
			log.Errorf("untar dir %s error %v", busyboxURL, err)
		}
	}
}

func CreateWriteLayer(containerName string) {
	writeURL := fmt.Sprintf(WriteLayer,containerName)
	if err := os.MkdirAll(writeURL, 0777); err != nil {
		log.Errorf("mkdir dir %s error. %v", writeURL, err)
	}
}

func CreateMountPoint(imageName string, containerName string) {
	mnt :=fmt.Sprintf(MntURL,containerName)
	if err := os.MkdirAll(mnt, 0777); err != nil {
		log.Errorf("mkdir dir %s error. %v", mnt, err)
	}
	wdir:=fmt.Sprintf(WriteLayer,containerName)
	dirs := "dirs=" + wdir + ":" + RootURL + "/" +imageName
	//left first rw,the second is ro
	//centos 7 aufs xfs err
	//run:mount -t tmpfs -o size=200M tmpfs /tmp
	cmd := exec.Command("mount", "-t", "aufs", "-o", dirs, "none", fmt.Sprintf(MntURL,containerName))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Errorf("%v", err)
	}
}



func DeleteMountPoint(volume string, containerName string){
	//first umount volume
	mnt :=fmt.Sprintf(MntURL,containerName)
	log.Infof("volume:%s",volume)
	if volume!=""{
		cmd := exec.Command("umount", mnt+"/"+volume)
		cmd.Stdout=os.Stdout
		cmd.Stderr=os.Stderr
		if err := cmd.Run(); err != nil {
			log.Errorf("%v",err)
		}
	}

	cmd := exec.Command("umount", mnt)
	cmd.Stdout=os.Stdout
	cmd.Stderr=os.Stderr
	if err := cmd.Run(); err != nil {
		log.Errorf("%v",err)
	}
	if err := os.RemoveAll(mnt); err != nil {
		log.Errorf("remove dir %s error %v", mnt, err)
	}
}

func DeleteWriteLayer(containerName string) {
	writeURL := fmt.Sprintf(WriteLayer,containerName)
	if err := os.RemoveAll(writeURL); err != nil {
		log.Errorf("remove dir %s error %v", writeURL, err)
	}
}

func MountVolume(volumeDirs []string,containerName string){

	parentVolume:=volumeDirs[0]
	if err:=os.MkdirAll(parentVolume,0777);err!=nil{
		log.Warnf("mkdir "+parentVolume+" error",err)
	}
	containerDir:=volumeDirs[1]
	containerVolume:=fmt.Sprintf(MntURL,containerName)+"/"+containerDir
	if err:=os.MkdirAll(containerVolume,0777);err!=nil{
		log.Warnf("mkdir "+containerVolume+" error",err)
	}
	dirs:="dirs="+parentVolume
	log.Infof("containerVolume:%s,dirs:%s",containerVolume,dirs)
	cmd:=exec.Command("mount","-t","aufs","-o",dirs,"none",containerVolume)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Errorf("mount volume %v", err)
	}
}


func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}


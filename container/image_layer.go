package container

import (
	"os"
	"os/exec"
	log "github.com/sirupsen/logrus"
)


func CreateReadOnlyLayer(rootURL string) {
	busyboxURL := rootURL + "busybox/"
	busyboxTarURL := rootURL + "busybox.tar"
	exist, err := PathExists(busyboxURL)
	if err != nil {
		log.Infof("fail to judge whether dir %s exists. %v", busyboxURL, err)
	}
	if exist == false {
		if err := os.Mkdir(busyboxURL, 0777); err != nil {
			log.Errorf("mkdir dir %s error. %v", busyboxURL, err)
		}
		if _, err := exec.Command("tar", "-xvf", busyboxTarURL, "-C", busyboxURL).CombinedOutput(); err != nil {
			log.Errorf("untar dir %s error %v", busyboxURL, err)
		}
	}
}

func CreateWriteLayer(rootURL string) {
	writeURL := rootURL + "writeLayer/"
	if err := os.Mkdir(writeURL, 0777); err != nil {
		log.Errorf("mkdir dir %s error. %v", writeURL, err)
	}
}

func CreateMountPoint(rootURL string, mntURL string) {
	if err := os.Mkdir(mntURL, 0777); err != nil {
		log.Errorf("mkdir dir %s error. %v", mntURL, err)
	}
	dirs := "dirs=" + rootURL + "writeLayer:" + rootURL + "busybox"
	//left first rw,the second is ro
	//centos 7 aufs xfs err
	//run:mount -t tmpfs -o size=200M tmpfs /tmp
	cmd := exec.Command("mount", "-t", "aufs", "-o", dirs, "none", mntURL)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Errorf("%v", err)
	}
}



func DeleteMountPoint(volume string, mntURL string){
	//first umount volume
	log.Infof("volume:%s",volume)
	if volume!=""{
		cmd := exec.Command("umount", mntURL+volume)
		cmd.Stdout=os.Stdout
		cmd.Stderr=os.Stderr
		if err := cmd.Run(); err != nil {
			log.Errorf("%v",err)
		}
	}

	cmd := exec.Command("umount", mntURL)
	cmd.Stdout=os.Stdout
	cmd.Stderr=os.Stderr
	if err := cmd.Run(); err != nil {
		log.Errorf("%v",err)
	}
	if err := os.RemoveAll(mntURL); err != nil {
		log.Errorf("remove dir %s error %v", mntURL, err)
	}
}

func DeleteWriteLayer(rootURL string) {
	writeURL := rootURL + "writeLayer/"
	if err := os.RemoveAll(writeURL); err != nil {
		log.Errorf("remove dir %s error %v", writeURL, err)
	}
}

func MountVolume(volumeDirs []string,mntURL string){

	parentVolume:=volumeDirs[0]
	if err:=os.MkdirAll(parentVolume,0777);err!=nil{
		log.Warnf("mkdir "+parentVolume+" error",err)
	}
	containerDir:=volumeDirs[1]
	containerVolume:=mntURL+containerDir
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


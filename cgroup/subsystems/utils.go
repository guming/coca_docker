package subsystems

import (
	"os"
	log "github.com/sirupsen/logrus"
	"bufio"
	"strings"
	"path"
	"fmt"
)

func findMountPoint(subsystem string) string {

	mfile,err:=os.Open("/proc/self/mountinfo")
	if err!=nil{
		log.Error(err.Error())
		return ""
	}

	mscanner:=bufio.NewScanner(mfile)
	for mscanner.Scan(){
		txt:=mscanner.Text()
		fileds:=strings.Split(txt," ")
		for _,opt:=range strings.Split(fileds[len(fileds)-1],","){
			if opt==subsystem{
				return fileds[4]
			}
		}
	}

	if err := mscanner.Err(); err != nil {
		return ""
	}
	return ""
}

func GetCgroupPath(subsystem string,cgroupPath string,autoCreate bool) (string,error){

	cgroupRoot:=findMountPoint(subsystem)
	if _,err:=os.Stat(path.Join(cgroupRoot,cgroupPath));err!=nil||(autoCreate&&os.IsNotExist(err)){
		if os.IsNotExist(err) {
			if err := os.Mkdir(path.Join(cgroupRoot,cgroupPath),0755);err==nil {
			} else {
				return "", fmt.Errorf("error create cgroup %v", err)
			}
		}
		return path.Join(cgroupRoot, cgroupPath), nil
	} else {
		return "", fmt.Errorf("cgroup path error %v", err)
	}

}
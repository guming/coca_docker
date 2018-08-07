package subsystems

import (
	"path"
	"strconv"
	"os"
	"io/ioutil"
	log "github.com/sirupsen/logrus"
)

type CpusetSubsystem struct {

}

func (cs *CpusetSubsystem) Set (cgroup string,config *ResourceConfig) error{
	subcgrouppath,err:=GetCgroupPath(cs.Name(),cgroup,true)
	if err==nil && subcgrouppath!="" {
		if err:=ioutil.WriteFile(path.Join(subcgrouppath, "cpuset.cpus"),
			[]byte(config.MemoryLimit), 0644);err!=nil{
			log.Errorf("set cgroup cpuset.cpus err %v",err)
		}
		return nil

	}else {
		return err
	}

}

func (cs *CpusetSubsystem) Apply(cgroup string,pid int) error{
	if subcgrouppath,err:=GetCgroupPath(cs.Name(),cgroup,false);err==nil{
		if err:=ioutil.WriteFile(path.Join(subcgrouppath,"task"),[]byte(strconv.Itoa(pid)),0644);err!=nil{
			return err
		}
		return nil
	}else{
		return err
	}
}

func (cs *CpusetSubsystem) Remove(cgroup string) error{
	if subcgrouppath,err:=GetCgroupPath(cs.Name(),cgroup,false);err==nil{
		err := os.RemoveAll(subcgrouppath)
		if err!=nil{
			return err
		}
		return nil
	}else {
		return err
	}
}

func (cs *CpusetSubsystem) Name() string{
	return "cpuset"
}

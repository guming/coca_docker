package subsystems

import (
	"os"
	"path"
	"strconv"
	"io/ioutil"
	log "github.com/sirupsen/logrus"
)

type CpuSubsystem struct {

}

func (cs *CpuSubsystem) Set (cgroup string,config *ResourceConfig) error{
	subcgrouppath,err:=GetCgroupPath(cs.Name(),cgroup,true)
	if err==nil && subcgrouppath!="" {
		if err:=ioutil.WriteFile(path.Join(subcgrouppath, "cpu.shares"),
			[]byte(config.MemoryLimit), 0644);err!=nil{
			log.Errorf("set cgroup cpu.shares err %v",err)
		}
		return nil

	}else {
		return err
	}

}

func (cs *CpuSubsystem) Apply(cgroup string,pid int) error{
	if subcgrouppath,err:=GetCgroupPath(cs.Name(),cgroup,false);err==nil{
		if err:=ioutil.WriteFile(path.Join(subcgrouppath,"task"),[]byte(strconv.Itoa(pid)),0644);err!=nil{
			return err
		}
		return nil
	}else{
		return err
	}
}

func (cs *CpuSubsystem) Remove(cgroup string) error{
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

func (cs *CpuSubsystem) Name() string{
	return "cpu"
}

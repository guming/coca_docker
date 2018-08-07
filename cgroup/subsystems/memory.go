package subsystems

import (
	"io/ioutil"
	"path"
	log "github.com/sirupsen/logrus"
	"os"
	"strconv"
)

type MemorySubSystem struct {
}
func (ms *MemorySubSystem) Set(cgroup string,config *ResourceConfig) error{
	subcgrouppath,err:=GetCgroupPath(ms.Name(),cgroup,true)
	if err==nil && subcgrouppath!="" {
		if err:=ioutil.WriteFile(path.Join(subcgrouppath, "memory.limit_in_bytes"),
			[]byte(config.MemoryLimit), 0644);err!=nil{
				log.Errorf("set cgroup memory limit err %v",err)
		}
		return nil

	}else {
		return err
	}

}

func (ms *MemorySubSystem) Apply(cgroup string,pid int) error{
	if subcgrouppath,err:=GetCgroupPath(ms.Name(),cgroup,false);err==nil{
		if err:=ioutil.WriteFile(path.Join(subcgrouppath,"task"),[]byte(strconv.Itoa(pid)),0644);err!=nil{
			return err
		}
		return nil
	}else{
		return err
	}
}

func (ms *MemorySubSystem) Remove(cgroup string) error{
	if subcgrouppath,err:=GetCgroupPath(ms.Name(),cgroup,false);err==nil{
		err := os.RemoveAll(subcgrouppath)
		if err!=nil{
			return err
		}
		return nil
	}else {
		return err
	}
}

func (ms *MemorySubSystem) Name() string{
	return "memory"
}

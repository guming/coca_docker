package cgroup

import (
	"github.com/coca_docker/cgroup/subsystems"
)



var (
	SubsystemsIns = []subsystems.Subsystem{
		&subsystems.CpusetSubsystem{},
		&subsystems.MemorySubSystem{},
		&subsystems.CpuSubsystem{},
	}
)

type CgroupManager struct {
	config *subsystems.ResourceConfig
	path string
}

func NewCgroupManager(path string) *CgroupManager {
	return &CgroupManager{
		path:path,
	}
}

func (cm *CgroupManager) Set(config *subsystems.ResourceConfig) error{
	for _,sub:=range SubsystemsIns{
		sub.Set(cm.path,config)
	}
	return nil
}


func (cm *CgroupManager) Apply(pid int) error{
	for _,sub:=range SubsystemsIns{
		sub.Apply(cm.path,pid)
	}
	return nil
}

func (cm *CgroupManager) Destory() error{
	for _,sub:=range SubsystemsIns{
		sub.Remove(cm.path)
	}
	return nil
}



package subsystems

type ResourceConfig struct{
	MemoryLimit string
	Cpuset string
	Cpushare string
}

type Subsystem interface {
	Set(path string,config *ResourceConfig) error
	Apply(path string,pid int) error
	Remove(path string) error
	Name() string
}


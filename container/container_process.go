package container

import (
	log "github.com/sirupsen/logrus"
	"syscall"
	"os"
	"os/exec"
)

func NewParentProcess(tty bool,command string) *exec.Cmd{
	args:=[]string{"init",command}
	cmd:=exec.Command("/proc/self/exe",args...)
	cmd.SysProcAttr=&syscall.SysProcAttr{Cloneflags:syscall.CLONE_NEWUTS|syscall.CLONE_NEWPID|
		syscall.CLONE_NEWNET| syscall.CLONE_NEWIPC|syscall.CLONE_NEWNS}

	if tty{
		cmd.Stdin=os.Stdin
		cmd.Stdout=os.Stdout
		cmd.Stderr=os.Stderr
	}
	return cmd
}

func RunContainerInit(command string,args []string) error{
	log.Infof("command %",command)
	defaultMountFlags:=syscall.MS_NOSUID|syscall.MS_NODEV|syscall.MS_NOEXEC
	syscall.Mount("proc","/proc","proc",uintptr(defaultMountFlags),"")
	argv:=[]string{command}
	err:=syscall.Exec(command,argv,os.Environ())
	if err!=nil{
		log.Errorf(err.Error())
	}
	return nil
}

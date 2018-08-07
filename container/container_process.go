package container

import (
	log "github.com/sirupsen/logrus"
	"syscall"
	"os"
	"os/exec"
	"io/ioutil"
	"strings"
	"fmt"
)

func NewParentProcess(tty bool) (*exec.Cmd,*os.File){

	readpip, writepip, err := NewPipe()

	if err != nil {
		log.Errorf("New pipe error %v", err)
		return nil, nil
	}
	cmd:=exec.Command("/proc/self/exe","init")
	cmd.SysProcAttr=&syscall.SysProcAttr{Cloneflags:syscall.CLONE_NEWUTS|syscall.CLONE_NEWPID|
		syscall.CLONE_NEWNET| syscall.CLONE_NEWIPC|syscall.CLONE_NEWNS}

	if tty{
		cmd.Stdin=os.Stdin
		cmd.Stdout=os.Stdout
		cmd.Stderr=os.Stderr
	}
	cmd.ExtraFiles=[]*os.File{readpip}
	return cmd,writepip
}

func RunContainerInit() error{
	//log.Infof("command %",command)

	cmds:=readCommand()
	if cmds==nil||len(cmds)==0{
		return fmt.Errorf("run container get command error, cmds is nil")
	}
	//defaultMountFlags:=syscall.MS_NOSUID|syscall.MS_NODEV|syscall.MS_NOEXEC
	//syscall.Mount("proc","/proc","proc",uintptr(defaultMountFlags),"")
	path,err:=exec.LookPath(cmds[0])
	if err!=nil{
		return fmt.Errorf("command lookup path error , cmds is %s",cmds[0])
	}
	err=syscall.Exec(path,cmds[0:],os.Environ())
	if err!=nil{
		log.Errorf(err.Error())
	}
	return nil
}

func NewPipe() (*os.File, *os.File, error) {
	read, write, err := os.Pipe()
	if err != nil {
		return nil, nil, err
	}
	return read, write, nil
}


func readCommand() []string{
	fpip:=os.NewFile(uintptr(3),"pipe")
	commands,err:=ioutil.ReadAll(fpip)
	if err!=nil{
		log.Errorf("init read pipe error %v",err.Error())
		return nil
	}
	msg:=string(commands)
	return strings.Split(msg," ")
}

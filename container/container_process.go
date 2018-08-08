package container

import (
	log "github.com/sirupsen/logrus"
	"syscall"
	"os"
	"os/exec"
	"io/ioutil"
	"strings"
	"fmt"
	"path/filepath"
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
	cmd.Dir="/opt/busybox"
	return cmd,writepip
}

func RunContainerInit() error{

	cmds:=readCommand()
	if cmds==nil||len(cmds)==0{
		return fmt.Errorf("run container get command error, cmds is nil")
	}
	//mout fs
	setUpMount()

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


func pivot_root(root string) error {
	log.Infof("root is %s",root)
	err:=syscall.Mount(root,root,"bind",syscall.MS_BIND|syscall.MS_REC,"")
	if err!=nil{
		return fmt.Errorf("pivot_root mount err %v",err)
	}
	pivotDir:=filepath.Join(root,".pivot_root")
	if err:=os.Mkdir(pivotDir,0777);err!=nil {
		return err
	}

	if err:=syscall.PivotRoot(root,pivotDir);err!=nil{
		return fmt.Errorf("pivot_root err %v",err)
	}

	if err:=syscall.Chdir("/");err!=nil{
		return fmt.Errorf("chdir / err %v",err)
	}
	pivotDir=filepath.Join("/",".pivot_root")
	if err:=syscall.Unmount(pivotDir,syscall.MNT_DETACH);err!=nil{
		return fmt.Errorf("unmount oldroot err %v",err)
	}
	return os.Remove(pivotDir)
}

func setUpMount(){
	dir,err:=os.Getwd()
	if err!=nil{
		log.Errorf("setup mount err %v",err)
		return
	}
	log.Infof("setup mount local dir is %s",dir)
	//change rootfs
	err:=pivot_root(dir)
	if err!=nil{
		log.Errorf("pivot root err %v",err)
	}
	//mount proc
	defaultMountFlags:=syscall.MS_NOSUID|syscall.MS_NODEV|syscall.MS_NOEXEC
	syscall.Mount("proc","/proc","proc",uintptr(defaultMountFlags),"")
	syscall.Mount("tmpfs", "/dev", "tmpfs", syscall.MS_NOSUID|syscall.MS_STRICTATIME, "mode=755")
}

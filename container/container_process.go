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

var (
	RUNNING             string = "running"
	STOP                string = "stopped"
	Exit                string = "exited"
	DefaultInfoLocation string = "/var/run/coca_docker/%s/"
	ConfigName          string = "config.json"
	ContainerLogFile	string = "container.log"
)

type ContainerInfo struct {
	Pid         string `json:"pid"` //容器的init进程在宿主机上的 PID
	Id          string `json:"id"`  //容器Id
	Name        string `json:"name"`  //容器名
	Command     string `json:"command"`    //容器内init运行命令
	CreatedTime string `json:"createTime"` //创建时间
	Status      string `json:"status"`     //容器的状态
	ImageName   string `json:"ImageName"` //容器镜像
}

var (
	RootURL string = "/root"
	MntURL string = "/root/mnt/%s"
	WriteLayer string = "/root/writelayer/%s"
)

func NewParentProcess(tty bool,volume string,containerName string,imageName string,envSlice []string) (*exec.Cmd,*os.File){

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
	}else{
		dirURL := fmt.Sprintf(DefaultInfoLocation, containerName)
		if err := os.MkdirAll(dirURL, 0622); err != nil {
			log.Errorf("new parent process mkdir %s error %v", dirURL, err)
			return nil, nil
		}
		stdLogFilePath := dirURL + ContainerLogFile
		stdLogFile, err := os.Create(stdLogFilePath)
		if err != nil {
			log.Errorf("new parent process create file %s error %v", stdLogFilePath, err)
			return nil, nil
		}
		cmd.Stdout = stdLogFile
	}
	cmd.ExtraFiles=[]*os.File{readpip}
	cmd.Env=envSlice
	//cmd.Dir="/root/busybox"
	//mntURL := "/root/mnt/"
	//rootURL := "/root/"
	NewWorkSpace(containerName, imageName,volume)
	cmd.Dir = fmt.Sprintf(MntURL,containerName)
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
	log.Infof("path is %s",path)
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
	log.Infof("msg is %s",msg)
	return strings.Split(msg," ")
}


func pivot_root(newroot string) error {

	log.Infof("root is %s",newroot)

	err:=syscall.Mount(newroot,newroot,"bind",uintptr(syscall.MS_BIND|syscall.MS_REC),"")
	if err!=nil{
		return err
	}

	putold:=filepath.Join(newroot,".pivot_root")
	if err:=os.MkdirAll(putold,0700);err!=nil {
		return err
	}

	if err:=syscall.PivotRoot(newroot,putold);err!=nil{
		return err
	}

	if err:=os.Chdir("/");err!=nil{
		return fmt.Errorf("chdir / err %v",err)
	}
	putold = "/.pivot_root"
	if err := syscall.Unmount(putold, syscall.MNT_DETACH); err != nil {
		return err
	}
	// remove putold
	if err := os.RemoveAll(putold); err != nil {
		return err
	}
	return nil
}

func setUpMount(){
	dir,err:=os.Getwd()
	if err!=nil{
		log.Errorf("setup mount err %v",err)
		return
	}
	log.Infof("setup mount local dir is %s",dir)
	//change rootfs
	err=pivot_root(dir)
	if err!=nil{
		log.Errorf("call pivot root err %v",err)
	}
	//mount proc centos 7
	defaultMountFlags:=syscall.MS_NOSUID|syscall.MS_NODEV|syscall.MS_NOEXEC
	if err := syscall.Mount("", "/", "", uintptr(defaultMountFlags|syscall.MS_PRIVATE|syscall.MS_REC), ""); err != nil {
		log.Errorf("mount err %v",err)
	}
	syscall.Mount("proc","/proc","proc",uintptr(defaultMountFlags),"")
	syscall.Mount("tmpfs", "/dev", "tmpfs", uintptr(syscall.MS_NOSUID|syscall.MS_STRICTATIME), "mode=755")
}


//Create a AUFS filesystem as container root workspace
func NewWorkSpace(containerName string, imageName string,volume string) {
	CreateReadOnlyLayer(imageName)
	CreateWriteLayer(containerName)
	CreateMountPoint(imageName, containerName)

	if volume!="" && len(volume)>1 {
		volumeDirs:=strings.Split(volume,":")
		if len(volumeDirs)>1{
			MountVolume(volumeDirs,containerName)
		}
	}
}

//Delete the AUFS filesystem while container exit
func DeleteWorkSpace(containerName string, imageName string,volume string) {
	var volumeDir=""
	if volume!="" && len(volume)>1 {
		volumeDirs:=strings.Split(volume,":")
		volumeDir=volumeDirs[1]
	}
	rootURL:=RootURL+"/"+imageName
	mntURL:=fmt.Sprintf(MntURL,containerName)
	DeleteMountPoint(volumeDir, mntURL)
	DeleteWriteLayer(rootURL)

}


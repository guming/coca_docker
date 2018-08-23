package network

import (
	"os"
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"path"
	"net"
	"strings"
)

//import "net"
const ipamDefaultAllocatorPath="/var/run/coca_docker/network/ipam/subnet.json"
type IPAM struct {
	SubnetAllocatorPath string
	Subnets *map[string]string
}
var ipAllocator = &IPAM{
	SubnetAllocatorPath: ipamDefaultAllocatorPath,
}

func (ipam *IPAM) load() error{
	_,err:=os.Stat(ipam.SubnetAllocatorPath)
	if os.IsNotExist(err) {
		return nil
	} else {
		return err
	}
	jsonfile,err:=os.Open(ipam.SubnetAllocatorPath)
	if err!=nil{
		return err
	}
	defer jsonfile.Close()
	subnetjson:=make([]byte,2000)
	n,err:=jsonfile.Read(subnetjson)

	if err!=nil{
		return err
	}

	err=json.Unmarshal(subnetjson[:n],ipam.Subnets)
	if err!=nil{
		log.Errorf("error load ipam json info, %v", err)
		return err
	}
	return nil
}

func (ipam *IPAM) dump() error{
	jsondir,_:=path.Split(ipam.SubnetAllocatorPath)
	if _, err := os.Stat(jsondir); err != nil {
		if os.IsNotExist(err) {
			os.MkdirAll(jsondir, 0644)
		} else {
			return err
		}
	}
	jsonfile,err:=os.OpenFile(ipam.SubnetAllocatorPath,os.O_TRUNC|os.O_CREATE|os.O_WRONLY,0644)
	defer jsonfile.Close()
	if err!=nil {
		return err
	}
	jsonbyte,err:=json.Marshal(ipam.Subnets)
	if err!=nil {
		return err
	}
	_,err=jsonfile.Write(jsonbyte)
	if err!=nil {
		log.Errorf("error dump ipam info, %v", err)
		return err
	}
	return nil
}


func (ipam *IPAM) Allocate (subnet *net.IPNet) (ip net.IP,err error) {
	ipam.Subnets=&map[string]string{}
	err=ipam.load()
	if err!=nil{
		log.Errorf("error allocation ipam info, %v", err)
	}
	n,size:=subnet.Mask.Size()
	if _,flag:=(*ipam.Subnets)[subnet.String()];!flag{
		(*ipam.Subnets)[subnet.String()]=strings.Repeat("0",1<<uint8(size-n))
	}
	for c:=range((*ipam.Subnets)[subnet.String()]){
		if (*ipam.Subnets)[subnet.String()][c]=='0'{
			ipalloc:=[]byte((*ipam.Subnets)[subnet.String()])
			ipalloc[c]='1'
			(*ipam.Subnets)[subnet.String()] = string(ipalloc)
			ip=subnet.IP
			for t:=uint(4);t>0;t-=1{
				[]byte(ip)[4-t]+=uint8(c>>((t-1)*8))
			}
			ip[3]+=1
			break
		}
	}
	ipam.dump()
	return
}
func (ipam *IPAM) Release(subnet *net.IPNet,ipaddr *net.IP) error{

	ipam.Subnets=&map[string]string{}
	_, subnet, _ = net.ParseCIDR(subnet.String())

	err := ipam.load()
	if err != nil {
		log.Errorf("error release ipam info, %v", err)
		return err
	}
	c := 0
	releaseIP := ipaddr.To4()
	releaseIP[3]-=1
	for t := uint(4); t > 0; t-=1 {
		c += int(releaseIP[t-1] - subnet.IP[t-1]) << ((4-t) * 8)
	}

	ipalloc := []byte((*ipam.Subnets)[subnet.String()])
	log.Infof("c is %d",c)
	ipalloc[c] = '0'
	(*ipam.Subnets)[subnet.String()] = string(ipalloc)

	ipam.dump()
	return nil
}


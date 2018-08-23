package network

import (
	"net"
	"github.com/vishvananda/netlink"
	"os"
	"encoding/json"
	"path"
	log "github.com/sirupsen/logrus"
)

type Network struct {
	Name string //网络名称
	IpRange net.IPNet //地址段
	Driver string //网络驱动名称 brige
}

type Endpoint struct {
	ID string `json:"id"`
	Device netlink.Veth `json:"dev"`
	IPAddress net.IP	`json:"ip_address"`
	MacAddress net.HardwareAddr `json:"mac_address"`
	PortMapping []string `json:"port_mapping"`
	Network *Network	`json:"network"`
}

type NetworkDriver interface {
	Name() string
	Create(subnet string,name string) (*Network,error)
	Delete(network Network) error
	Connect(endpoint *Endpoint,network *Network) error
	Disconnect(endpoint *Endpoint,network *Network) error
}


var (
	defaultNetworkPath = "/var/run/mydocker/network/network/"
	drivers = map[string]NetworkDriver{}
	networks = map[string]*Network{}
)

func (nw *Network) load(dumpPath string) error{
	_,err:=os.Stat(dumpPath)
	if os.IsNotExist(err) {
		return nil
	} else {
		return err
	}
	jsonfile,err:=os.Open(path.Join(dumpPath,nw.Name))
	if err!=nil{
		return err
	}
	defer jsonfile.Close()
	subnetjson:=make([]byte,2000)
	n,err:=jsonfile.Read(subnetjson)

	if err!=nil{
		return err
	}

	err=json.Unmarshal(subnetjson[:n],nw)
	if err!=nil{
		log.Errorf("error load nw json info, %v", err)
		return err
	}
	return nil
}

func (nw *Network) dump(dumpPath string) error{
	jsondir,_:=path.Split(dumpPath)
	if _, err := os.Stat(jsondir); err != nil {
		if os.IsNotExist(err) {
			os.MkdirAll(jsondir, 0644)
		} else {
			return err
		}
	}
	jsonfile,err:=os.OpenFile(dumpPath,os.O_TRUNC|os.O_CREATE|os.O_WRONLY,0644)
	defer jsonfile.Close()
	if err!=nil {
		return err
	}
	jsonbyte,err:=json.Marshal(nw)
	if err!=nil {
		return err
	}
	_,err=jsonfile.Write(jsonbyte)
	if err!=nil {
		log.Errorf("error dump nw info, %v", err)
		return err
	}
	return nil
}

func (nw *Network) remove(dumpPath string) error {
	if _, err := os.Stat(path.Join(dumpPath, nw.Name)); err != nil {
		if os.IsNotExist(err) {
			return nil
		} else {
			return err
		}
	} else {
		return os.Remove(path.Join(dumpPath, nw.Name))
	}
}

func CreateNetwork(driver,name,subnet string) error {
	_,cidr,_:=net.ParseCIDR(subnet)
	gatewayIp,err:=ipAllocator.Allocate(cidr)
	if err!=nil{
		return err
	}
	cidr.IP=gatewayIp
	nw,err:=drivers[driver].Create(cidr.String(),name)
	if err!=nil{
		return err
	}

	return nw.dump(defaultNetworkPath)
}

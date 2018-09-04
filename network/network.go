package network

import (
	"net"
	"github.com/vishvananda/netlink"
	"os"
	"encoding/json"
	"path"
	log "github.com/sirupsen/logrus"
	"github.com/coca_docker/container"
	"fmt"
	"github.com/vishvananda/netns"
	"runtime"
	"strings"
	"os/exec"
	"path/filepath"
	"text/tabwriter"
)

var (
	defaultNetworkPath = "/var/run/coca_docker/network/network/"
	drivers = map[string]NetworkDriver{}
	networks = map[string]*Network{}
)

type Network struct {
	Name string //网络名称
	IpRange *net.IPNet //地址段
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
	//create network: create veth peer and bridge
	Create(subnet string,name string) (*Network,error)
	Delete(network Network) error
	Connect(endpoint *Endpoint,network *Network) error
	Disconnect(endpoint *Endpoint,network *Network) error
}

func (nw *Network) load (dumpPath string) error{
	log.Infof("dump path is %s",dumpPath)

	jsonfile,err:=os.Open(dumpPath)
	defer jsonfile.Close()
	if err!=nil{
		return err
	}
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

func (nw *Network) dump (dumpPath string) error{
	jsondir,_:=path.Split(dumpPath)
	if _, err := os.Stat(jsondir); err != nil {
		if os.IsNotExist(err) {
			os.MkdirAll(jsondir, 0644)
		} else {
			return err
		}
	}
	nwPath := path.Join(dumpPath, nw.Name)
	jsonfile,err:=os.OpenFile(nwPath,os.O_TRUNC|os.O_CREATE|os.O_WRONLY,0644)
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

func Init() error {
	var bridgeDriver = BridgeNetworkDriver{}
	drivers[bridgeDriver.Name()] = &bridgeDriver

	if _, err := os.Stat(defaultNetworkPath); err != nil {
		if os.IsNotExist(err) {
			os.MkdirAll(defaultNetworkPath, 0644)
		} else {
			return err
		}
	}

	filepath.Walk(defaultNetworkPath, func(nwPath string, info os.FileInfo, err error) error {
		if strings.HasSuffix(nwPath, "/") {
			return nil
		}
		_, nwName := path.Split(nwPath)
		nw := &Network{
			Name: nwName,
		}
		log.Infof("network path is %s",nwPath)
		if err := nw.load(nwPath); err != nil {
			log.Errorf("error load network: %v", err)
		}

		networks[nwName] = nw
		return nil
	})
	return nil
}

func CreateNetwork(driver string,name string,subnet string) error {
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

func DeleteNetwork(networkName string) error {

	nw:=networks[networkName]
	if err:=ipAllocator.Release(nw.IpRange,&nw.IpRange.IP);err!=nil{
		return fmt.Errorf("release ip error %v",err)
	}
	if err:=drivers[nw.Driver].Delete(*nw);err!=nil{
		return fmt.Errorf("driver delete error %v",err)
	}
	return nw.remove(defaultNetworkPath)
}

func ListNetwork() {
	w := tabwriter.NewWriter(os.Stdout, 12, 1, 3, ' ', 0)
	fmt.Fprint(w, "NAME\tIpRange\tDriver\n")
	for _, nw := range networks {
		fmt.Fprintf(w, "%s\t%s\t%s\n",
			nw.Name,
			nw.IpRange.String(),
			nw.Driver,
		)
	}
	if err := w.Flush(); err != nil {
		log.Errorf("flush error %v", err)
		return
	}
}

func Connect(networkName string,cinfo *container.ContainerInfo) error {

	nw,ok:=networks[networkName]
	if !ok {
		return fmt.Errorf("no such network: %s", networkName)
	}
	log.Infof("nw iprange is %s and ip is %s",nw.IpRange.String(),nw.IpRange.IP.String())
	//allocate new ip for container
	ip,err:=ipAllocator.Allocate(nw.IpRange)
	if err!=nil{
		return err
	}
	//create endpoint for container
	enp:=&Endpoint{
		ID:fmt.Sprintf("%s-%s", cinfo.Id, networkName),
		Network:nw,
		IPAddress:ip,
		PortMapping:cinfo.PortMapping,
	}
	err=drivers[nw.Driver].Connect(enp,nw)
	if err!=nil{
		log.Errorf("network connect error %v %s %s ",err,nw.IpRange.String(),nw.IpRange.IP.String())
		ipAllocator.Release(nw.IpRange,&nw.IpRange.IP)
		return err
	}
	//set ip for container
	//add route for container
	//set netns for container
	err=configIpAddrRouteForEndpoint(enp,cinfo)
	if err!=nil {
		log.Errorf("config ipaddr route for endpoint error %v ",err)
		return fmt.Errorf("config ipaddr route for endpoint error %v",err)
	}
	return configPortMapping(enp,cinfo)
}

func Disconnect(networkName string, cinfo *container.ContainerInfo) error {
	return nil
}

func configIpAddrRouteForEndpoint(endpoint *Endpoint, cinfo *container.ContainerInfo) error {
	pname:=endpoint.Device.PeerName
	peerlink,_:=netlink.LinkByName(pname)
	//ipnet:=endpoint.Network.IpRange.String()
	defer enterContainerNetNS(&peerlink,cinfo)
	interfaceIP := *endpoint.Network.IpRange
	interfaceIP.IP = endpoint.IPAddress
	log.Infof("interfaceip is %s",interfaceIP.String())

	err:=setInterfaceIP(pname,interfaceIP.String())
	if err!=nil{
		log.Errorf("setInterfaceIP error %v %v",err,endpoint.Network)
		return err
	}
	err=setInterfaceUP(pname)
	if err!=nil{
		log.Errorf("setInterfaceUP error %v %s",err,pname)
		return err
	}
	//set up lo veth
	err=setInterfaceUP("lo")
	if err!=nil{
		log.Errorf("setInterfaceUP error %v lo",err)
		return err
	}
	//route config;all the traffics through the bridge(driver)
	_,cidr,_:=net.ParseCIDR("0.0.0.0/0")
	route:=&netlink.Route{
		LinkIndex:peerlink.Attrs().Index,
		Dst:cidr,
		Gw:endpoint.Network.IpRange.IP,
	}
	//add route
	err=netlink.RouteAdd(route)
	if err!=nil{
		log.Errorf("configIpAddrRouteForEndpoint RouteAdd error %v",err)
		return err
	}
	return nil
}

func enterContainerNetNS(link *netlink.Link,cinfo *container.ContainerInfo) func() {
	f,err:=os.OpenFile(fmt.Sprintf("/proc/%s/ns/net",cinfo.Pid),os.O_RDONLY,0)
	if err!=nil{
		log.Errorf("error get container net namespace, %v", err)
	}
	nsfd:=f.Fd()
	runtime.LockOSThread()
	if err:=netlink.LinkSetNsFd(*link,int(nsfd));err!=nil{
		log.Errorf("error set nsfd, %v", err)
	}

	origins,err:=netns.Get()
	if err!=nil{
		log.Errorf("error netns get, %v", err)
	}
	if err:=netns.Set(netns.NsHandle(nsfd));err!=nil{
		log.Errorf("error netns nshandle, %v", err)
	}
	return func(){
		netns.Set(origins)
		origins.Close()
		runtime.UnlockOSThread()
		f.Close()
	}
}

func configPortMapping(ep *Endpoint, cinfo *container.ContainerInfo) error {
	for _, pm := range ep.PortMapping {
		portMapping :=strings.Split(pm, ":")
		if len(portMapping) != 2 {
			log.Errorf("port mapping format error, %v", pm)
			continue
		}
		iptablesCmd := fmt.Sprintf("-t nat -A PREROUTING -p tcp -m tcp --dport %s -j DNAT --to-destination %s:%s",
			portMapping[0], ep.IPAddress.String(), portMapping[1])
		cmd := exec.Command("iptables", strings.Split(iptablesCmd, " ")...)
		output, err := cmd.Output()
		if err != nil {
			log.Errorf("iptables output, %v", output)
			continue
		}
	}
	return nil
}
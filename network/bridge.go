package network

import (
	"github.com/vishvananda/netlink"
	"net"
	"strings"
	log "github.com/sirupsen/logrus"
	"fmt"
	"time"
	"os/exec"
)

type BridgeNetworkDriver struct {
}

func (bd *BridgeNetworkDriver) Name() string {
	return "bridge"
}

func (bd *BridgeNetworkDriver) Create(subnet string,name string) (*Network,error) {
	ip,ipRange,_:=net.ParseCIDR(subnet)
	log.Infof("Create ip & iprange %s %s",ip.String(),ipRange.String())

	ipRange.IP=ip
	nw:=&Network{
		Name:name,
		IpRange:ipRange,
		Driver:bd.Name(),
	}
	err:=bd.initBridge(nw)
	if err!=nil{
		log.Errorf("init bridge error %v ",err)
	}
	log.Infof("create nw is %v",nw)
	return nw,err
}

func (bd *BridgeNetworkDriver) Delete (network Network) error {
	bname:=network.Name
	lb,err:=netlink.LinkByName(bname)
	if err!=nil{
		log.Errorf("delete bridge interface error %v",err)
		return err
	}
	return netlink.LinkDel(lb)
}

func (bd *BridgeNetworkDriver) Connect (endpoint *Endpoint,nw *Network) error {
	bridgeName := nw.Name
	br,err:=netlink.LinkByName(bridgeName)
	if err != nil {
		return err
	}
	la:=netlink.NewLinkAttrs()
	la.Name=endpoint.ID[:5]
	//put the veth to bridge
	la.MasterIndex=br.Attrs().Index

	//Veth peer
	endpoint.Device=netlink.Veth{
		LinkAttrs:la,
		PeerName:"cif-"+endpoint.ID[:5],
		}
	//create veth
	err=netlink.LinkAdd(&endpoint.Device)
	if err!=nil {
		return err
	}
	//set up veth
	err=netlink.LinkSetUp(&endpoint.Device)
	if err!=nil {
		return err
	}
	return nil
}

func (d *BridgeNetworkDriver) Disconnect(endpoint *Endpoint,network *Network) error {
	return nil
}

func (bd *BridgeNetworkDriver) initBridge(nw *Network) error {
	bname:=nw.Name
	if err:=createBridgeInterface(bname);err!=nil{
		log.Errorf("create bridge interface error %v ",err)
		return err
	}

	//set gateway ip
	gatewayIP := *nw.IpRange
	gatewayIP.IP = nw.IpRange.IP

	if err:=setInterfaceIP(bname,gatewayIP.String());err!=nil{
		log.Errorf("set interface ip error %v ",err)
		return err
	}

	if err:=setInterfaceUP(bname);err!=nil{
		log.Errorf("set interface up error %v ",err)
		return err
	}

	if err:=setIpTables(bname,nw.IpRange.String());err!=nil{
		log.Errorf("set iptables error %v ",err)
		return err
	}
	return nil
}

func createBridgeInterface(bname string) error{

	_,err:=net.InterfaceByName(bname)
	//err ==nil exist bridge
	if err==nil||!strings.Contains(err.Error(), "no such network interface"){
		return err
	}

	la:=netlink.NewLinkAttrs()
	la.Name=bname

	br:=&netlink.Bridge{LinkAttrs: la}
	err=netlink.LinkAdd(br)
	if err!=nil {
		return fmt.Errorf("create bridge interface %s error %v",bname,err)
	}
	return nil
}

func setInterfaceIP(bname string,ipnet string) error {

	retries := 2
	var br netlink.Link
	var err error
	for i := 0; i < retries; i++ {
		br, err = netlink.LinkByName(bname)
		if err == nil {
			break
		}
		log.Infof("error retrieving new bridge netlink link %s  retrying2", bname)
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		return fmt.Errorf("Abandoning retrieving the new bridge link from netlink, Run [ ip link ] to troubleshoot the error: %v", err)
	}
	log.Infof("setup ip is %s",ipnet)
	ipn,err:=netlink.ParseIPNet(ipnet)
	if err!=nil{
		return err
	}
	addr:=&netlink.Addr{IPNet:ipn}

	return netlink.AddrAdd(br,addr)
}

func setInterfaceUP(bname string) error {
	br,err:=netlink.LinkByName(bname)
	if err!=nil{
		return err
	}
	return netlink.LinkSetUp(br)
}

func setIpTables(bname string,ipnet string) error {
	cmdstr:=fmt.Sprintf(" -t nat -A POSTROUTING -s %s ! -o %s -j MASQUERADE", ipnet,bname)
	cmd:=exec.Command("iptables",strings.Split(cmdstr," ")...)
	result,err:=cmd.Output()
	if err!=nil{
		log.Errorf("iptables exec output %v ",result)
		log.Errorf("iptables exec error %v ",err)
		return err
	}
	return nil
}




package network

import (
	"net"
	"github.com/vishvananda/netlink"
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






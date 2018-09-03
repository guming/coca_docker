package network

import (
	"testing"
	"github.com/coca_docker/container"
)

func TestBridgeNetworkDriver_Create(t *testing.T) {
	bnd:=&BridgeNetworkDriver{}
	_,err:=bnd.Create("10.30.0.1/24","testbridge")
	t.Logf("err: %v", err)
}

func TestBridgeNetworkDriver_Connect(t *testing.T) {
	endpoint:=&Endpoint{
		ID:"testcontainer",
	}
	nw:=&Network{
		Name:"testbridge",
	}
	bnd:=&BridgeNetworkDriver{}
	err:=bnd.Connect(endpoint,nw)
	t.Logf("err: %v", err)
}

func TestConnect(t *testing.T) {
	cInfo := &container.ContainerInfo{
		Id: "testcontainer1",
		Pid: "36000",
	}
	d := BridgeNetworkDriver{}
	n, err := d.Create("10.50.0.1/24", "testbridge1")
	t.Logf("err: %v", n)
	Init()
	networks[n.Name] = n
	err = Connect(n.Name, cInfo)
	t.Logf("err: %v", err)
}

func TestLoad(t *testing.T){
	n := Network{
		Name: "testbridge",
	}
	n.load("/var/run/coca_docker/network/network/testbridge")
	t.Logf("network: %v", n)
}

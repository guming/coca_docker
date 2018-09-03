package network

import (
	"net"
	"testing"
	"github.com/sirupsen/logrus"
)

func TestAllocate(t *testing.T) {
	_, ipnet, _ := net.ParseCIDR("192.168.0.1/24")
	ip, _ := ipAllocator.Allocate(ipnet)
	t.Logf("alloc ip: %v", ip)
}


func TestRelease(t *testing.T) {
	ip, ipnet, _ := net.ParseCIDR("192.168.0.1/24")
	logrus.Printf("ip %s ",ip.String())
	ipAllocator.Release(ipnet, &ip)
}


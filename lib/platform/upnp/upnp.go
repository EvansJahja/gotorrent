package upnp

import (
	"errors"
	"fmt"
	"net"

	"example.com/gotorrent/lib/core/adapter/portexposer"
	"github.com/huin/goupnp/dcps/internetgateway2"
)

func New(localPort uint16) portexposer.PortExposer {
	return &impl{
		localPort:    localPort,
		startExtPort: 6083, // starting extPort
	}
}

type impl struct {
	localPort    uint16
	startExtPort uint16
	extPort      uint16
	client       *internetgateway2.WANIPConnection1
}

func (i *impl) Start() {
	clients, errs, err := internetgateway2.NewWANIPConnection1Clients()
	if len(errs) != 0 {
		panic(errs)
	}
	if err != nil {
		panic(errs)
	}
	if len(clients) == 0 {
		panic("no clients")
	}

	// Assume first IGD client is ours
	client := clients[0]

	myIP, err := findMyLocalIP(client.Location.Host)
	if err != nil {
		panic(err)
	}

Retry:
	internalPort, internalClient, enabled, portMappingDesc, leaseDuration, err :=
		client.GetSpecificPortMappingEntry("", i.startExtPort, "TCP")
	_ = internalClient
	_ = enabled
	_ = portMappingDesc
	_ = leaseDuration

	if err == nil {
		internalClientIP := net.ParseIP(internalClient)
		if !internalClientIP.Equal(myIP) {
			i.startExtPort++
			goto Retry
		}
		if internalPort != i.localPort {
			i.startExtPort++
			goto Retry
		}
	}

	if err := client.AddPortMapping("", i.startExtPort, "TCP", i.localPort, myIP.String(), false, "port desc", 0); err != nil {
		fmt.Println(err.Error())
	}
	i.client = client
	i.extPort = i.startExtPort

	fmt.Printf("UPnP success: %s exposing %d as %d\n", myIP.String(), i.localPort, i.startExtPort)
}

func (i *impl) Port() uint16 {
	return i.extPort
}

func (i *impl) Stop() {
	if i.client == nil {
		return
	}
	i.client.DeletePortMapping("", i.startExtPort, "TCP")
}

func findMyLocalIP(igdHostname string) (net.IP, error) {

	gwHPs := igdHostname
	gwIps, _, err := net.SplitHostPort(gwHPs)
	if err != nil {
		return nil, err
	}

	gwIP := net.ParseIP(gwIps)

	// Find our IP based on interface that shares IGD IP
	nwIfs, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	for _, nwIf := range nwIfs {
		addresses, err := nwIf.Addrs()
		if err != nil {
			return nil, err
		}

		for _, addr := range addresses {
			ip, ipNet, err := net.ParseCIDR(addr.String())
			if err != nil {
				return nil, err
			}
			if ipNet.Contains(gwIP) {
				return ip, nil
			}

		}

	}

	return nil, errors.New("no interface matches given IGD")
}

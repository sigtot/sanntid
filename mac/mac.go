package mac

import (
	"fmt"
	"net"
)

// GetMacAddr gets the network-cards mac address
func GetMacAddr() (string, error) {
	mac, err := net.InterfaceByName("eno1")
	if err != nil {
		mac, err = net.InterfaceByName("eth1")
		if err != nil {
			mac, err = net.InterfaceByName("eth0")
			if err != nil {
				panic(fmt.Sprintf("No MAC address found. %s", err.Error()))
			}
		}
	}
	as := mac.HardwareAddr.String()

	return as, nil
}

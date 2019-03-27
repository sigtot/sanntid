package mac

import (
	"net"
)

// GetMacAddr gets the network-cards mac address
func GetMacAddr() (string, error) {
	mac, err := net.InterfaceByName("eno1")
	if err != nil {
		mac, err = net.InterfaceByName("eth1")
		if err != nil {
			panic("No MAC address found.")
		}
	}
	as := mac.HardwareAddr.String()

	return as, nil
}

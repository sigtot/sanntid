package mac

import (
	"net"
)

func GetMacAddr() (string, error) {
	ifs, err := net.InterfaceByName("eno1")
	if err != nil {
		ifs, err = net.InterfaceByName("eth1")
		if err != nil {
			panic("No MAC address found.")
		}
	}
	as := ifs.HardwareAddr.String()

	return as, nil
}

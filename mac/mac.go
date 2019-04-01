/*
Package mac contains the functionality for finding the eno1, eth1 or eth0 mac-addresses on linux-ubuntu machines.
*/
package mac

import (
	"net"
)

var interfaceNames = []string{"eno1", "eth1", "eth0"}

// GetMacAddr gets the network-cards mac address
func GetMacAddr() (string, error) {
	err := error(nil)
	for _, interfaceName := range interfaceNames {
		mac, newErr := net.InterfaceByName(interfaceName)
		err = mergeErrors(err, newErr)
		if err == nil {
			return mac.HardwareAddr.String(), err
		}
	}
	return "", err
}

func mergeErrors(oldErr error, newErr error) error {
	if oldErr != nil {
		return oldErr
	}
	if newErr != nil {
		return newErr
	}
	return nil
}

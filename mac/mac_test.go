package mac

import (
	"fmt"
	"testing"
)

func TestGetMacAddr(t *testing.T) {
	addr, err := GetMacAddr()
	if err != nil {
		panic(fmt.Sprintf("Could not get mac address, %s", err))
	}
	fmt.Printf("%s", addr)
}

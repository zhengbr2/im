package net

import (
	"fmt"
	"strings"
)

const (
	networkSpliter = "@"
)

func ParseNetwork(str string) (network, addr string, err error) {
	var idx int
	if idx = strings.Index(str, networkSpliter); idx == -1 {
		err = fmt.Errorf("addr: \"%s\" error, must be network@tcp:port or network@unixsocket", str)
		return
	}
	network = str[:idx]
	addr = str[idx+1:]
	return
}

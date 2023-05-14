//go:build windows
// +build windows

package nat

import "net"

func NewNATDialer() *net.Dialer {
	return &net.Dialer{}
}

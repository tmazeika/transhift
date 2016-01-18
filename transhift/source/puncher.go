package source

import (
	"github.com/transhift/transhift/transhift/puncher"
	"github.com/transhift/transhift/common/protocol"
	"crypto/tls"
	"fmt"
	"errors"
	"net"
)

func punchHole(host string, port int, cert tls.Certificate, id string) (laddr net.Addr, targetAddr string, err error) {
	p := puncher.New(host, port, protocol.SourceNode, cert)
	if laddr, err = p.Connect(); err != nil {
		return
	}

	// Send ID.
	if err = p.Enc().Encode(id); err != nil {
		return
	}

	// Expect signal.
	var sig protocol.Signal
	if err = p.Dec().Decode(&sig); err != nil {
		return
	}

	switch sig {
	case protocol.TargetNotFoundSignal:
		err = errors.New("unknown ID")
		return
	case protocol.OkaySignal:
	default:
		err = fmt.Errorf("protocol error: unknown signal 0x%x", sig)
		return
	}

	// Expect target address.
	err = p.Dec().Decode(&targetAddr)
	return
}

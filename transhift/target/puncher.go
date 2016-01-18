package target

import (
	"github.com/transhift/transhift/transhift/puncher"
	"github.com/transhift/transhift/common/protocol"
	"crypto/tls"
)

func punchHole(host, port string, cert *tls.Certificate) (sourceAddr string, id string, err error) {
	p := puncher.New(host, port, protocol.TargetNode, cert)

	if err := p.Connect(); err != nil {
		return err
	}

	// Expect ID.
	if err = p.Dec().Decode(&id); err != nil {
		return
	}

	// Expect target address.
	err = p.Dec().Decode(&sourceAddr)
	return
}

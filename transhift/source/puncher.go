package source

import (
	"github.com/transhift/transhift/transhift/puncher"
	"github.com/transhift/transhift/common/protocol"
	"crypto/tls"
)

func punchHole(host, port string, cert *tls.Certificate, id string) (targetAddr string, err error) {
	p := puncher.New(host, port, protocol.SourceNode, cert)

	if err := p.Connect(); err != nil {
		return err
	}

	// Send ID.
	if err = p.Enc().Encode(id); err != nil {
		return
	}

	// Expect target address.
	err = p.Dec().Decode(&targetAddr)
	return
}

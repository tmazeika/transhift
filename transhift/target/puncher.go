package target

import (
	"github.com/transhift/transhift/transhift/puncher"
	"github.com/transhift/transhift/common/protocol"
	"crypto/tls"
	"log"
)

func punchHole(host string, port int, cert tls.Certificate) (sourceAddr string, err error) {
	p := puncher.New(host, port, protocol.TargetNode, cert)

	if err = p.Connect(); err != nil {
		return
	}

	// Expect ID.
	var id string
	if err = p.Dec().Decode(&id); err != nil {
		return
	}

	log.Printf("Your ID is '%f'", id)

	// Expect target address.
	err = p.Dec().Decode(&sourceAddr)
	return
}

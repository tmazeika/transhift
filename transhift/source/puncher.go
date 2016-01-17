package source

import (
	"github.com/transhift/transhift/transhift/puncher"
	"crypto/tls"
	"github.com/transhift/transhift/common/protocol"
)

type sourcePuncher struct {
	puncher.Puncher
}

func NewPuncher(host, port string, cert *tls.Certificate) *sourcePuncher {
	return puncher.New(host, port, protocol.SourceNode, cert)
}

func (p *sourcePuncher) sendId(id string) error {
	return p.Enc().Encode(id)
}

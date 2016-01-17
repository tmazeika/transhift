package puncher

import (
	"net"
	"crypto/tls"
	"encoding/gob"
	"github.com/transhift/transhift/common/protocol"
)

type Puncher interface {
	Connect() error

	Enc() *gob.Encoder

	Dec() *gob.Decoder
}

type puncher struct {
	net.Conn

	host     string
	port     string
	nodeType protocol.NodeType
	cert     *tls.Certificate
	enc      *gob.Encoder
	dec      *gob.Decoder
}

func New(host, port string, nodeType protocol.NodeType, cert *tls.Certificate) *puncher {
	return &puncher{
		host:     host,
		port:     port,
		nodeType: nodeType,
		cert:     cert,
	}
}

func (p *puncher) Connect() (err error) {
	if p.Conn, err = tls.Dial("tcp", net.JoinHostPort(p.host, p.port), p.tlsConfig()); err != nil {
		return
	}

	p.enc = gob.NewEncoder(p.Conn)
	p.dec = gob.NewDecoder(p.Conn)

	// Send NodeType.
	return p.enc.Encode(p.nodeType)
}

func (p *puncher) Enc() *gob.Encoder {
	return p.enc
}

func (p *puncher) Dec() *gob.Decoder {
	return p.dec
}

func (p *puncher) tlsConfig() *tls.Config {
	return &tls.Config{
		Certificates:       []tls.Certificate{*p.cert},
		InsecureSkipVerify: true,
		MinVersion:         tls.VersionTLS12,
	}
}


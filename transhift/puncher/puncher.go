package puncher

import (
	"net"
	"crypto/tls"
	"encoding/gob"
	"github.com/transhift/transhift/common/protocol"
	"github.com/transhift/transhift/transhift/tprotocol"
	"strconv"
)

type puncher struct {
	net.Conn

	host     string
	port     int
	nodeType protocol.NodeType
	cert     tls.Certificate
	enc      *gob.Encoder
	dec      *gob.Decoder
}

func New(host string, port int, nodeType protocol.NodeType, cert tls.Certificate) *puncher {
	return &puncher{
		host:     host,
		port:     port,
		nodeType: nodeType,
		cert:     cert,
	}
}

func (p *puncher) Connect() (laddr net.Addr, err error) {
	if p.Conn, err = tls.Dial("tcp", net.JoinHostPort(p.host, strconv.Itoa(p.port)), tprotocol.TlsConfig(p.cert)); err != nil {
		return
	}

	laddr = p.Conn.LocalAddr()
	p.enc = gob.NewEncoder(p.Conn)
	p.dec = gob.NewDecoder(p.Conn)

	// Send NodeType.
	err = p.enc.Encode(p.nodeType)
	return
}

func (p *puncher) Enc() *gob.Encoder {
	return p.enc
}

func (p *puncher) Dec() *gob.Decoder {
	return p.dec
}

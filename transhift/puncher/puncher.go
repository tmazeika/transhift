package puncher

import (
	"net"
	"crypto/tls"
	"encoding/gob"
)

type puncher struct {
	net.Conn

	host string
	port string
	cert *tls.Certificate
	enc  *gob.Encoder
	dec  *gob.Decoder
}

func New(host, port string, cert *tls.Certificate) *puncher {
	return &puncher{
		host: host,
		port: port,
		cert: cert,
	}
}

func (p *puncher) Connect() (err error) {
	if p.Conn, err = tls.Dial("tcp", net.JoinHostPort(p.host, p.port), p.tlsConfig()); err != nil {
		return
	}

	p.enc = gob.NewEncoder(p.Conn)
	p.dec = gob.NewDecoder(p.Conn)
	return
}

func (p *puncher) tlsConfig() *tls.Config {
	return &tls.Config{
		Certificates:       []tls.Certificate{*p.cert},
		InsecureSkipVerify: true,
		MinVersion:         tls.VersionTLS12,
	}
}


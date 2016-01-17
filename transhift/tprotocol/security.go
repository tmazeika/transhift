package tprotocol

import "crypto/tls"

func TlsConfig(cert *tls.Certificate) *tls.Config {
	return &tls.Config{
		Certificates:       []tls.Certificate{*cert},
		InsecureSkipVerify: true,
		MinVersion:         tls.VersionTLS12,
	}
}

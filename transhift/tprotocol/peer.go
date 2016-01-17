package tprotocol

import (
	"net"
	"crypto/tls"
	"time"
	"encoding/gob"
)

type peer struct {
	net.Conn

	Enc   *gob.Encoder
	Dec   *gob.Decoder
	cert  *tls.Certificate
	raddr string
}

func NewPeer(raddr string) *peer {
	return &peer{
		raddr: raddr,
	}
}

func (p *peer) Connect() error {
	tlsConf := TlsConfig(p.cert)

	done := make(chan struct{})
	t := ticker(done)

	// Keep trying to connect. The first connect is called at:
	//   floor(now + 2 seconds)
	// Subsequent connects are called every 1 second from that point forward.
	// The purpose of this is to employ TCP simultaneous open by sending SYN's
	// every 5 seconds on the second, which the peer should be doing as well.
	for {
		// Wait for interval.
		<-t

		var err error
		if p.Conn, err = tls.Dial("tcp", p.raddr, tlsConf); err != nil {
			done <- struct{}{}
			break
		}
	}

	p.Enc = gob.NewEncoder(p.Conn)
	p.Dec = gob.NewDecoder(p.Conn)
	return nil
}

func ticker(done <-chan struct{}) <-chan struct{} {
	const Interval = time.Second * 5
	ch := make(chan struct{})

	go func() {
		defer close(ch)

		waitForNextSecondCeiled()
		t := time.NewTicker(Interval)
		ch <- struct{}{}

		for {
			select {
			case <-t.C:
				ch <- struct{}
			case <-done:
				t.Stop()
				return
			}
		}
	}()

	return ch
}

func waitForNextSecondCeiled() {
	nextSecondCeiled := time.Now().Add(time.Second * 2).Truncate(time.Second)

	for {
		if !time.Now().Before(nextSecondCeiled) {
			break
		}
	}
}

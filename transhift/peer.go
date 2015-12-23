package transhift

import (
    "net"
    "fmt"
    "time"
)

type DownloadPeer struct {
    conn net.Conn
}

func (p *DownloadPeer) Connect(host string, port string) {
    fmt.Printf("Dialing %s... ", host)

    for p.conn == nil {
        p.conn = net.Dial("tcp", net.JoinHostPort(host, port))

        if p.conn == nil {
            time.Sleep(time.Second)
        }
    }

    fmt.Println("connected")
}

package transhift

import (
    "net"
    "fmt"
    "time"
    "encoding/binary"
    "bytes"
)

type Peer struct {
    conn net.Conn
}

func (p *Peer) SendRaw(data []byte) {
    var buff bytes.Buffer
    intBuff := make([]byte, 8)

    binary.BigEndian.PutUint64(intBuff, uint64(len(data)))
    buff.Write(intBuff)
    buff.Write(data)
    p.conn.Write(buff.Bytes())
}

func (p *Peer) SendInt(data uint64) {
    buff := make([]byte, 8)

    binary.BigEndian.PutUint64(buff, data)
    p.SendRaw(buff);
}

func (p *Peer) ParseInt(data []byte) uint64 {
    return binary.BigEndian.Uint64(data)
}

type DownloadPeer struct {
    peer Peer
}

func (d *DownloadPeer) Connect(host string, port string) {
    fmt.Printf("Dialing %s... ", host)

    for d.peer.conn == nil {
        d.peer.conn, _ = net.Dial("tcp", net.JoinHostPort(host, port))

        if d.peer.conn == nil {
            time.Sleep(time.Second)
        }
    }

    fmt.Println("connected")
}

func (p *DownloadPeer) SendPassword(password string) {
    p.peer.SendRaw(stringChecksum(password))
}

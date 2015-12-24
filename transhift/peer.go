package transhift

import (
    "net"
    "fmt"
    "time"
    "log"
    "bufio"
)

type DownloadPeer struct {
    conn net.Conn

    ch chan []byte
}

func (d *DownloadPeer) Connect(host string, port string) {
    fmt.Printf("Dialing %s... ", host)

    for d.conn == nil {
        d.conn = net.Dial("tcp", net.JoinHostPort(host, port))

        if d.conn == nil {
            time.Sleep(time.Second)
        }
    }

    fmt.Println("connected")
}

func (d *DownloadPeer) Channel() chan []byte {
    d.ch = make(chan []byte)

    // read
    go func() {
        reader := bufio.NewReader(d.conn)

        for {
            data, err := reader.ReadBytes('\n')

            if err != nil {
                log.Fatalln("Error reading line: ", err)
            }

            d.ch <- data
        }
    }()

    return d.ch
}

func (d *DownloadPeer) SendPassword(password string) {
    fmt.Fprintln(d.conn, stringChecksum(password))
}

func (d *DownloadPeer) SendFileInfo(name string, size uint64) {
    // name
    fmt.Fprintln(d.conn, name)
    // size
    fmt.Fprintln(d.conn, size)
}

func (d *DownloadPeer) SendFileChunk(chunk []byte) {
    d.conn.Write(chunk)
}

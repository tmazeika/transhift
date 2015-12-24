package transhift

import (
    "net"
    "fmt"
    "time"
    "log"
    "bufio"
    "strconv"
    "os"
)

func portStr(port uint16) string {
    return strconv.Itoa(int(port))
}

type DownloadPeer struct {
    conn net.Conn

    ch chan []byte
}

func (d *DownloadPeer) Connect(host string, port uint16) {
    fmt.Printf("Dialing %s:%d... ", host, port)

    for d.conn == nil {
        d.conn, _ = net.Dial("tcp", net.JoinHostPort(host, portStr(port)))

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
                d.conn.Close()
                fmt.Fprintln(os.Stderr, "Error reading line: ", err)
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

type UploadPeer struct {
    conn net.Conn

    ch chan []byte
}

func (d *UploadPeer) Connect(port uint16) error {
    fmt.Printf("Listening on port %d... ", port)

    var err error
    listener, err := net.Listen("tcp", net.JoinHostPort("", portStr(port)))

    if err != nil {
        return err
    }

    d.conn, err = listener.Accept()

    if err != nil {
        return err
    }

    fmt.Println("connected")

    return nil
}

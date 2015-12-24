package transhift

import (
    "net"
    "fmt"
    "time"
    "bufio"
    "strconv"
    "os"
    "encoding/binary"
    "bytes"
)

// protocol errors
const (
    PasswordMismatch = byte(iota)
    ChecksumMismatch = byte(iota)
)

func portStr(port uint16) string {
    return strconv.Itoa(int(port))
}

func makePeerReadChannel(reader *bufio.Reader) (ch chan []byte) {
    ch = make(chan []byte)

    go func() {
        for {
            data, err := reader.ReadBytes('\n')

            if err != nil {
                fmt.Fprintln(os.Stderr, "Error reading line: ", err)
            }

            ch <- data
        }
    }()
}

// ************************************************************************** //
// * DownloadPeer                                                           * //
// ************************************************************************** //

type DownloadPeer struct {
    conn   net.Conn
    reader bufio.Reader
    writer bufio.Writer
}

func (d *DownloadPeer) Connect(host string, port uint16) {
    fmt.Printf("Dialing %s:%d... ", host, port)

    for d.conn == nil {
        d.conn, _ = net.Dial("tcp", net.JoinHostPort(host, portStr(port)))

        if d.conn == nil {
            time.Sleep(time.Second)
        }
    }

    d.reader = bufio.NewReader(d.conn)
    d.writer = bufio.NewWriter(d.conn)

    fmt.Println("connected")
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

// ************************************************************************** //
// * UploadPeer                                                             * //
// ************************************************************************** //

type UploadPeer struct {
    conn     net.Conn
    reader   bufio.Reader
    writer   bufio.Writer
    in       chan []byte
    fileSize uint64
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

    d.reader = bufio.NewReader(d.conn)
    d.writer = bufio.NewWriter(d.conn)
    d.in = make(chan []byte)

    go func() {
        for {
            data, err := d.reader.ReadBytes('\n')

            if err != nil {
                fmt.Fprintln(os.Stderr, "Error reading line: ", err)
                break
            }

            d.in <- data
        }
    }()

    fmt.Println("connected")

    return nil
}

func (d *UploadPeer) ReceivePassword() string {
    return string(<- d.in)
}

func (d *UploadPeer) ReceiveFileInfo() (name string, size uint64) {
    name = string(<- d.in)
    size = binary.BigEndian.Uint64(<- d.in)
    d.fileSize = size
    return
}

func (d *UploadPeer) ReceiveFileChunks(chunkSize uint64) chan []byte {
    ch := make(chan []byte)

    var totalRead uint64

    go func() {
        for totalRead < d.fileSize {
            dataBuff := make([]byte, chunkSize)

            var chunkRead uint64

            for chunkRead < chunkSize {
                subDataRead, err := d.reader.Read(dataBuff[chunkRead:])

                if err != nil {
                    fmt.Fprintln(os.Stderr, "Error reading line: ", err)
                    return
                }

                chunkRead += subDataRead
            }
        }
    }()

    return ch
}

func (d *UploadPeer) SendProtocolError(err byte) {
    d.writer.WriteByte(err)
}

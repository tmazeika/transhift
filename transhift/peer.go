package transhift

import (
    "net"
    "fmt"
    "time"
    "bufio"
    "strconv"
    "os"
    "encoding/binary"
)

// protocol errors
const (
    PasswordMismatch = byte(iota)
    ChecksumMismatch = byte(iota)
)

func portStr(port uint16) string {
    return strconv.Itoa(int(port))
}

func min(x, y uint64) uint64 {
    if x < y {
        return x
    }

    return y
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
        // for as long as the total amount of bytes read is less than the file
        // size...
        for totalRead < d.fileSize {
            // this is the chunk size we'll actually be using; use either the
            // given chunk size or the number of bytes left until the end of the
            // file, whichever is most restrictive (smaller)
            adjustedChunkSize := min(d.fileSize - totalRead, chunkSize)

            // make a new buffer to hold the bytes for this chunk
            dataBuff := make([]byte, adjustedChunkSize)

            var chunkRead uint64

            for chunkRead < adjustedChunkSize {
                dataRead, err := d.reader.Read(dataBuff[chunkRead:])

                if err != nil {
                    fmt.Fprintln(os.Stderr, "Error reading line: ", err)
                    return
                }

                chunkRead += dataRead
            }
        }
    }()

    return ch
}

func (d *UploadPeer) SendProtocolError(err byte) {
    d.writer.WriteByte(err)
}

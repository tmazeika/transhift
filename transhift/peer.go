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

func (d *DownloadPeer) SendPasswordHash(password string) {
    fmt.Fprintln(d.conn, stringChecksum(password))
}

func (d *DownloadPeer) SendFileInfo(name string, size uint64, checksum []byte) {
    // name
    fmt.Fprintln(d.conn, name)
    // size
    sizeBuff := make([]byte, 8)
    binary.BigEndian.PutUint64(sizeBuff, size)
    d.conn.Write(sizeBuff)
    // checksum
    d.conn.Write(checksum)
    d.conn.Write(byte('\n'))
}

func (d *DownloadPeer) SendFileChunk(chunk []byte) {
    d.conn.Write(chunk)
}

func (d *DownloadPeer) ProtocolErrorChannel() chan byte {
    ch := make(chan byte)

    go func() {
        for {
            // read a single byte
            dataBuff := make([]byte, 1)
            _, err := d.conn.Read(dataBuff)

            if err != nil {
                fmt.Fprintln(os.Stderr, "Error reading byte: ", err)
                return
            }

            ch <- dataBuff[0]
        }
    }()

    return ch
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
            // read line from connection up to NL
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

func (d *UploadPeer) ReceivePasswordHash() string {
    return string(<- d.in)
}

func (d *UploadPeer) ReceiveFileInfo() (name string, size uint64, checksum []byte) {
    name = string(<- d.in)
    size = binary.BigEndian.Uint64(<- d.in)
    d.fileSize = size
    checksum = <- d.in
    return
}

func (d *UploadPeer) ReceiveFileChunks(chunkSize uint64, handler func([]byte)) {
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

            // for as long as we haven't finished filling up the buffer for the
            // chunk...
            for chunkRead < adjustedChunkSize {
                // read bytes from the connection starting at the last position
                // we read
                dataRead, err := d.reader.Read(dataBuff[chunkRead:])

                if err != nil {
                    fmt.Fprintln(os.Stderr, "Error reading line: ", err)
                    return
                }

                // add to the bytes read of this chunk whatever was just read
                chunkRead += dataRead
            }

            // the chunk is done being read, so off to get handled...
            handler(dataBuff)
        }
    }()
}

func (d *UploadPeer) SendProtocolError(err byte) {
    d.writer.WriteByte(err)
}

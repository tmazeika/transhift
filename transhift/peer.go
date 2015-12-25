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

// protocol responses
const (
    // from download peer; indicates the sent password was incorrect and the
    // process should not continue
    PasswordMismatch = byte(iota)

    // from download peer; indicates the sent password was correct and the
    // process should continue
    PasswordMatch    = byte(iota)

    // from download peer; indicates the received file's checksum did not match
    // the checksum sent from the upload peer
    ChecksumMismatch = byte(iota)

    // from download peer; indicates the received file's checksum matched the
    // checksum sent from the upload peer
    ChecksumMatch    = byte(iota)

    // from either endpoint; indicates the user has stopped sending/receiving
    // - download peer will send this via SendProtocolResponse() when needed,
    //   and the upload peer should listen
    // - upload peer will send this OR Continue as the first byte of every chunk
    Terminate        = byte(iota)

    // from upload peer; this OR Terminate is sent as the first byte of every
    // chunk
    Continue         = byte(iota)
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
    reader *bufio.Reader
    writer *bufio.Writer
}

func (d *DownloadPeer) Connect(host string, port uint16) {
    for d.conn == nil {
        d.conn, _ = net.Dial("tcp", net.JoinHostPort(host, portStr(port)))

        if d.conn == nil {
            fmt.Print(".")
            time.Sleep(time.Second)
        }
    }

    d.writer = bufio.NewWriter(d.conn)
    d.reader = bufio.NewReader(d.conn)
}

type MetaInfo struct {
    passHash []byte
    fileName string
    fileSize uint64
    fileHash []byte
}

func (d *DownloadPeer) SendMetaInfo(metaInfo *MetaInfo) {
    // passHash
    d.writer.Write(metaInfo.passHash)
    d.writer.WriteRune('\n')
    // fileName
    d.writer.WriteString(metaInfo.fileName)
    d.writer.WriteRune('\n')
    // fileSize
    sizeBuff := make([]byte, 8)
    binary.BigEndian.PutUint64(sizeBuff, metaInfo.fileSize)
    d.writer.Write(sizeBuff)
    d.writer.WriteRune('\n')
    // fileHash
    d.writer.Write(metaInfo.fileHash)
    d.writer.WriteRune('\n')

    d.writer.Flush()
}

func (d *DownloadPeer) SendFileChunk(fileChunk *FileChunk) {
    // good
    if fileChunk.good {
        d.writer.WriteByte(Continue)
    } else {
        d.writer.WriteByte(Terminate)
    }
    // data
    d.writer.Write(fileChunk.data)

    d.writer.Flush()
}

func (d *DownloadPeer) ProtocolResponseChannel() chan byte {
    ch := make(chan byte)

    go func() {
        for {
            // read a single byte
            dataBuff := make([]byte, 1)
            _, err := d.reader.Read(dataBuff)

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
    reader   *bufio.Reader
    writer   *bufio.Writer
    metaInfo *MetaInfo
}

func (d *UploadPeer) Connect(port uint16) error {
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

    return nil
}

func (d *UploadPeer) ReceiveMetaInfo() {
    in := make(chan []byte)

    const ExpectedReads = 4

    go func() {
        for i := 0; i < ExpectedReads; i++ {
            data, err := d.reader.ReadBytes('\n')

            if err != nil {
                fmt.Fprintln(os.Stderr, "Error receiving meta info: ", err)
                break
            }

            in <- data[:len(data) - 1]
        }
    }()

    d.metaInfo = &MetaInfo{
        passHash: <- in,
        fileName: string(<- in),
        fileSize: binary.BigEndian.Uint64(<- in),
        fileHash: <- in,
    }
}

type FileChunk struct {
    good bool
    data []byte
}

func (d *UploadPeer) ReceiveFileChunks(chunkSize uint64) (ch chan *FileChunk) {
    ch = make(chan *FileChunk)

    var totalRead uint64

    go func() {
        for totalRead < d.metaInfo.fileSize {
            adjustedChunkSize := min(d.metaInfo.fileSize - totalRead, chunkSize)

            fmt.Println("adjustedChunkSize", adjustedChunkSize)

            dataBuff := make([]byte, adjustedChunkSize)

            var chunkRead uint64

            for chunkRead < adjustedChunkSize {
                fmt.Println("chunkRead", chunkRead)
                dataRead, err := d.reader.Read(dataBuff[chunkRead:])
                fmt.Println("dataRead", dataRead)

                if err != nil {
                    fmt.Fprintln(os.Stderr, "Error reading bytes: ", err)
                    return
                }

                chunkRead += uint64(dataRead)
            }

            fmt.Println("chunkRead", chunkRead)

            ch <- &FileChunk{
                good: dataBuff[0] == Continue,
                data: dataBuff[1:],
            }

            fmt.Println(FileChunk{
                good: dataBuff[0] == Continue,
                data: dataBuff[1:],
            })

            totalRead += chunkRead
            fmt.Println("totalRead", totalRead)
        }
    }()

    return
}

func (d *UploadPeer) SendProtocolResponse(res byte) {
    d.writer.WriteByte(res)
    d.writer.Flush()
}

func (d *UploadPeer) Close() {
    d.conn.Close()
}

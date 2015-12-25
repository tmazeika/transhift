package transhift

import (
    "github.com/codegangsta/cli"
    "net"
    "bufio"
    "bytes"
    "fmt"
    "os"
)

type DownloadArgs struct {
    password    string
    destination string
}

func (a DownloadArgs) PasswordHash() []byte {
    return calculateStringChecksum(a.password)
}

func (a DownloadArgs) DestinationOrDef(def string) string {
    if a.destination == "" {
        return def
    }
    return a.destination
}

type UploadPeer struct {
    conn     net.Conn
    reader   *bufio.Reader
    writer   *bufio.Writer
    metaInfo *ProtoMetaInfo
}

func (p *UploadPeer) Connect(args DownloadArgs) error {
    listener, err := net.Listen("tcp", net.JoinHostPort("", ProtoPortStr))

    if err != nil {
        return err
    }

    p.conn, err = listener.Accept()

    if err != nil {
        return err
    }

    p.reader = bufio.NewReader(p.conn)
    p.writer = bufio.NewWriter(p.conn)
    return nil
}

func (p *UploadPeer) ReceiveMetaInfo() {
    const ExpectedNLCount = 4

    var buffer bytes.Buffer

    for i := 0; i < ExpectedNLCount; i++ {
        line, err := p.reader.ReadBytes('\n')

        if err != nil {
            fmt.Fprintln(os.Stderr, err)
            os.Exit(1)
        }

        buffer.Write(line)
    }

    p.metaInfo = &ProtoMetaInfo{}
    p.metaInfo.Deserialize(buffer.Bytes())
}

func (p *UploadPeer) ReceiveChunks() chan *ProtoChunk {
    ch := make(chan *ProtoChunk)
    var bytesRead uint64

    go func() {
        for bytesRead < p.metaInfo.fileSize {
            adjustedChunkSize := uint64Min(p.metaInfo.fileSize - bytesRead, ProtoChunkSize)
            chunkBuffer := make([]byte, adjustedChunkSize)
            var chunkBytesRead uint64

            for chunkBytesRead < adjustedChunkSize {
                subChunkBytesRead, _ := p.reader.Read(chunkBuffer[chunkBytesRead:])
                chunkBytesRead += subChunkBytesRead
            }

            bytesRead += adjustedChunkSize
            chunk := &ProtoChunk{}
            chunk.Deserialize(chunkBuffer)
            chunk.last = (bytesRead == p.metaInfo.fileSize)
            ch <- &chunk
            // TODO: chunk.close
        }
    }()

    return ch
}

func (p *UploadPeer) SendMessage(message byte) {
    p.writer.WriteByte(message)
    p.writer.Flush()
}

func Download(c *cli.Context) {
    args := DownloadArgs{
        password:    c.Args()[0],
        destination: c.String("destination"),
    }

    peer := &UploadPeer{}
    fmt.Print("Connecting to peer...")
    peer.ReceiveMetaInfo()

    // verify password
    if bytes.Equal(args.PasswordHash(), peer.metaInfo.passwordChecksum) {
        peer.SendMessage(ProtoMsgPasswordMatch)
        fmt.Println("done")
    } else {
        peer.SendMessage(ProtoMsgPasswordMismatch)
        fmt.Fprintln(os.Stderr, "password mismatch")
        os.Exit(1)
    }

    fmt.Print("Downloading... ")
    file, err := os.Create(args.DestinationOrDef(peer.metaInfo.fileName))

    if err != nil {
        fmt.Fprintln(os.Stderr, err)
        os.Exit(1)
    }

    ch := peer.ReceiveChunks()
    var bytesRead uint64

    for {
        chunk := <- ch
        file.WriteAt(chunk.data, int64(bytesRead))
        bytesRead += uint64(len(chunk.data))

        if chunk.last {
            break
        }
    }

    fmt.Println("done")
    fmt.Print("Verifying file... ")

    if bytes.Equal(calculateFileChecksum(file), peer.metaInfo.fileChecksum) {
        peer.SendMessage(ProtoMsgChecksumMatch)
        fmt.Println("done")
    } else {
        peer.SendMessage(ProtoMsgChecksumMismatch)
        fmt.Fprintln(os.Stderr, "checksum mismatch")
        os.Exit(1)
    }
}

func uint64Min(x, y uint64) uint64 {
    if x < y {
        return x
    }
    return y
}

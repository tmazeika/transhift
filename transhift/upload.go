package transhift

import (
    "github.com/codegangsta/cli"
    "path/filepath"
    "net"
    "bufio"
    "fmt"
    "os"
)

type UploadArgs struct {
    peerHost string
    password string
    filePath string
}

func (a UploadArgs) PasswordHash() []byte {
    return calculateStringChecksum(a.password)
}

func (a UploadArgs) AbsFilePath() string {
    filePath, _ := filepath.Abs(a.filePath)
    return filePath
}

type DownloadPeer struct {
    conn     net.Conn
    reader   *bufio.Reader
    writer   *bufio.Writer
    metaInfo *ProtoMetaInfo
}

func (p *DownloadPeer) Connect(args UploadArgs) {
    for p.conn == nil {
        p.conn, _ = net.Dial("tcp", net.JoinHostPort(args.peerHost, ProtoPortStr))
    }

    p.reader = bufio.NewReader(p.conn)
    p.writer = bufio.NewWriter(p.conn)
}

func (p *DownloadPeer) SendMetaInfo(metaInfo *ProtoMetaInfo) {
    p.writer.Write(metaInfo.Serialize())
    p.writer.Flush()
}

func (p *DownloadPeer) SendChunk(chunk []byte) {
    p.writer.Write(chunk)
    p.writer.Flush()
}

func (p *DownloadPeer) ReceiveMessages() chan byte {
    ch := make(chan byte)

    go func() {
        for {
            buffer := make([]byte, 1)
            p.reader.Read(buffer)
            ch <- buffer[0]
        }
    }()

    return ch
}

/**/

func Upload(c *cli.Context) {
    args := UploadArgs{
        peerHost: c.Args()[0],
        password: c.Args()[1],
        filePath: c.Args()[2],
    }

    peer := &DownloadPeer{}
    fmt.Print("Connecting to peer... ")
    peer.Connect(args)
    msgCh := peer.ReceiveMessages()
    file, err := os.Open(args.AbsFilePath())

    if err != nil {
        fmt.Fprintln(os.Stderr, err)
        os.Exit(1)
    }

    fileInfo, err := file.Stat()

    if err != nil {
        fmt.Fprintln(os.Stderr, err)
        os.Exit(1)
    }

    peer.SendMetaInfo(&ProtoMetaInfo{
        passwordChecksum: calculateStringChecksum(args.password),
        fileName: file.Name(),
        fileSize: uint64(fileInfo.Size()),
        fileChecksum: calculateFileChecksum(file),
    })

    switch <- msgCh {
    case ProtoMsgPasswordMatch:
        fmt.Println("done")
    case ProtoMsgPasswordMismatch:
        fmt.Fprintln(os.Stderr, "password mismatch")
        os.Exit(1)
    default:
        fmt.Fprintln(os.Stderr, "protocol error")
        os.Exit(1)
    }

    fmt.Print("Uploading... ")
    var bytesWritten uint64

    for bytesWritten < uint64(fileInfo.Size()) {
        adjustedChunkSize := uint64Min(uint64(fileInfo.Size()) - bytesWritten, ProtoChunkSize)
        chunkBuffer := make([]byte, adjustedChunkSize)
        chunkBytesWritten, _ := file.ReadAt(chunkBuffer, int64(bytesWritten))
        bytesWritten += uint64(chunkBytesWritten)
        peer.SendChunk(chunkBuffer)
    }

    fmt.Println("done")
    fmt.Print("Verifying file... ")

    switch <- msgCh {
    case ProtoMsgChecksumMatch:
        fmt.Println("done")
    case ProtoMsgChecksumMismatch:
        fmt.Fprintln(os.Stderr, "checksum mismatch")
        os.Exit(1)
    default:
        fmt.Fprintln(os.Stderr, "protocol error")
        os.Exit(1)
    }
}

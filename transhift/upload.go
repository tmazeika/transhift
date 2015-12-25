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

func (a UploadArgs) PasswordChecksum() []byte {
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
}

func (p *DownloadPeer) Connect(args UploadArgs) error {
    for p.conn == nil {
        p.conn, _ = net.Dial("tcp", net.JoinHostPort(args.peerHost, ProtoPortStr))
    }

    p.reader = bufio.NewReader(p.conn)
    p.writer = bufio.NewWriter(p.conn)

    err := checkCompatibility(p.reader, p.writer)

    if err != nil {
        return err
    }

    return nil
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

func Upload(c *cli.Context) {
    args := UploadArgs{
        peerHost: c.Args()[0],
        password: c.Args()[1],
        filePath: c.Args()[2],
    }

    peer := &DownloadPeer{}
    fmt.Printf("Connecting to '%s'... ", args.peerHost)
    err := peer.Connect(args)
    defer peer.conn.Close()

    if err != nil {
        fmt.Fprintln(os.Stderr, err)
        os.Exit(1)
    }

    fmt.Println("done")
    fmt.Print("Sending file info... ")

    msgCh := peer.ReceiveMessages()
    file, err := os.Open(args.filePath)
    defer file.Close()

    if err != nil {
        fmt.Fprintln(os.Stderr, err)
        os.Exit(1)
    }

    fileInfo, err := file.Stat()

    if err != nil {
        fmt.Fprintln(os.Stderr, err)
        os.Exit(1)
    }

    metaInfo := &ProtoMetaInfo{
        passwordChecksum: calculateStringChecksum(args.password),
        fileName: filepath.Base(file.Name()),
        fileSize: uint64(fileInfo.Size()),
        fileChecksum: calculateFileChecksum(file),
    }

    peer.SendMetaInfo(metaInfo)

    switch <- msgCh {
    case ProtoMsgPasswordMatch:
        fmt.Println(metaInfo)
    case ProtoMsgPasswordMismatch:
        fmt.Fprintln(os.Stderr, "password mismatch")
        os.Exit(1)
    default:
        fmt.Fprintln(os.Stderr, "protocol error")
        os.Exit(1)
    }

    fmt.Printf("Uploading '%s'...\n", args.AbsFilePath())
    var bytesWritten uint64
    progressBar := ProgressBar{
        current: &bytesWritten,
        total:   uint64(fileInfo.Size()),
    }
    progressBar.Start()

    for bytesWritten < uint64(fileInfo.Size()) {
        adjustedChunkSize := uint64Min(uint64(fileInfo.Size()) - bytesWritten, ProtoChunkSize)
        chunkBuffer := make([]byte, adjustedChunkSize)
        chunkBytesWritten, _ := file.ReadAt(chunkBuffer, int64(bytesWritten))
        bytesWritten += uint64(chunkBytesWritten)
        peer.SendChunk(chunkBuffer)
    }

    progressBar.Stop(true)
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

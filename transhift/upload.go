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
    peer     string
    filePath string
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

func (p *DownloadPeer) PunchHole(args UploadArgs) (remoteAddr string, err error) {
    conn, err := net.Dial("tcp", net.JoinHostPort(PuncherHost, PuncherPortStr))

    if err != nil {
        return "", err
    }

    defer conn.Close()
    conn.Write([]byte{byte(ProtoMsgClientTypeUL)})
    conn.Write([]byte(args.peer))
    line, err := bufio.NewReader(conn).ReadBytes('\n')

    line = line[:len(line) - 1] // trim trailing \n

    if err != nil {
        return "", err
    }

    return string(line), nil
}

func (p *DownloadPeer) Connect(remoteAddr string) error {
    for p.conn == nil {
        p.conn, _ = net.Dial("tcp", remoteAddr)
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
    p.conn.Write(metaInfo.Serialize())
}

func (p *DownloadPeer) SendChunk(chunk []byte) {
    p.conn.Write(chunk)
}

func (p *DownloadPeer) ReceiveMessages() chan ProtoMsg {
    ch := make(chan ProtoMsg)

    go func() {
        for {
            buffer := make([]byte, 1)
            p.conn.Read(buffer)
            ch <- ProtoMsg(buffer[0])
        }
    }()

    return ch
}

func Upload(c *cli.Context) {
    args := UploadArgs{
        peer:     c.Args()[0],
        filePath: c.Args()[1],
    }

    if len(args.peer) != ProtoPeerUIDLen {
        fmt.Fprintf(os.Stderr, "Peer UID should be %d characters\n", ProtoPeerUIDLen)
        os.Exit(1)
    }

    peer := &DownloadPeer{}
    fmt.Print("Waiting for peer... ")
    remoteAddr, err := peer.PunchHole(args)

    if err != nil {
        fmt.Fprintln(os.Stderr, err)
        os.Exit(1)
    }

    fmt.Println("done")
    fmt.Printf("Connecting to '%s'... ", args.peer)
    err = peer.Connect(remoteAddr)
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
        fileName: filepath.Base(file.Name()),
        fileSize: uint64(fileInfo.Size()),
        fileChecksum: calculateFileChecksum(file),
    }

    peer.SendMetaInfo(metaInfo)
    fmt.Println(metaInfo)
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

package transhift

import (
    "github.com/codegangsta/cli"
    "path/filepath"
    "net"
    "bufio"
    "fmt"
    "os"
    "crypto/tls"
    "bytes"
    "time"
)

type UploadArgs struct {
    peerUid  string
    filePath string
    appDir   string
}

func (u UploadArgs) AbsFilePath() string {
    filePath, _ := filepath.Abs(u.filePath)
    return filePath
}

type DownloadPeer struct {
    conn  *tls.Conn
    inOut *bufio.ReadWriter
}

func (DownloadPeer) PunchHole(peerUid string, config *Config) (remoteAddr string, err error) {
    // TODO: use tls.Dial
    conn, err := net.Dial("tcp", net.JoinHostPort(config.PuncherHost, config.PuncherPortStr()))

    if err != nil {
        return "", err
    }

    defer conn.Close()

    var buffer bytes.Buffer

    if _, err := buffer.Write(messageToBytes(UploadClientType)); err != nil {
        return "", err
    }

    if _, err := buffer.WriteString(peerUid); err != nil {
        return "", err
    }

    if _, err := conn.Write(buffer.Bytes()); err != nil {
        return "", err
    }

    scanner := bufio.NewScanner(bufio.NewReader(conn))

    if ! scanner.Scan() {
        return "", scanner.Err()
    }

    return scanner.Text(), nil
}

func (p *DownloadPeer) Connect(cert tls.Certificate, remoteAddr string) error {
    tlsConfig := &tls.Config{
        Certificates: []tls.Certificate{cert},
        InsecureSkipVerify: true,
        MinVersion: tls.VersionTLS12,
    }

    for p.conn == nil {
        var err error
        p.conn, err = tls.Dial("tcp", remoteAddr, tlsConfig)

        if err != nil {
            time.Sleep(time.Second)
        }
    }

    p.inOut = bufio.NewReadWriter(bufio.NewReader(p.conn), bufio.NewWriter(p.conn))

    return CheckCompatibility(p.inOut)
}

func (p DownloadPeer) SendFileInfo(fileInfo FileInfo) error {
    data, err := fileInfo.MarshalBinary()

    if err != nil {
        return err
    }

    if _, err := p.conn.Write(data); err != nil {
        return err
    }

    return nil
}

func (p DownloadPeer) ReceiveMessages() (ch chan ProtocolMessage) {
    ch = make(chan ProtocolMessage)
    scanner := bufio.NewScanner(p.inOut.Reader)

    scanner.Split(bufio.ScanBytes)

    go func() {
        for scanner.Scan() {
            ch <- ProtocolMessage(scanner.Bytes()[0])
        }
    }()

    return
}

func Upload(c *cli.Context) {
    args := UploadArgs{
        peerUid:  c.Args()[0],
        filePath: c.Args()[1],
        appDir:   c.GlobalString("app-dir"),
    }

    if len(args.peerUid) != UidLength {
        fmt.Fprintf(os.Stderr, "Peer UID should be %d characters\n", UidLength)
        os.Exit(1)
    }

    peer := DownloadPeer{}
    storage := Storage{
        customDir: args.appDir,
    }
    config, err := storage.Config()

    if err != nil {
        fmt.Fprintln(os.Stderr, err)
        os.Exit(1)
    }

    cert, err := storage.Crypto()

    if err != nil {
        fmt.Fprintln(os.Stderr, err)
        os.Exit(1)
    }

    fmt.Print("Waiting for peer... ")

    remoteAddr, err := peer.PunchHole(args.peerUid, config)

    if err != nil {
        fmt.Fprintln(os.Stderr, err)
        os.Exit(1)
    }

    fmt.Println("done")
    fmt.Printf("Connecting to '%s'... ", args.peerUid)

    err = peer.Connect(cert, remoteAddr)

    if err != nil {
        fmt.Fprintln(os.Stderr, err)
        os.Exit(1)
    }

    defer peer.conn.Close()

    fmt.Println("done")
    fmt.Print("Sending file info... ")

    msgCh := peer.ReceiveMessages()
    file, err := os.Open(args.filePath)

    if err != nil {
        fmt.Fprintln(os.Stderr, err)
        os.Exit(1)
    }

    defer file.Close()

    fileInfo, err := file.Stat()

    if err != nil {
        fmt.Fprintln(os.Stderr, err)
        os.Exit(1)
    }

    checksum, err := calculateFileChecksum(file)

    if err != nil {
        fmt.Fprintln(os.Stderr, err)
        os.Exit(1)
    }

    metaInfo := FileInfo{
        name:     filepath.Base(file.Name()),
        size:     uint64(fileInfo.Size()),
        checksum: checksum,
    }

    peer.SendFileInfo(metaInfo)
    fmt.Println(metaInfo)
    fmt.Printf("Uploading '%s'...\n", args.AbsFilePath())

    var bytesWritten uint64
    progressBar := ProgressBar{
        current: &bytesWritten,
        total:   uint64(fileInfo.Size()),
    }

    progressBar.Start()

    for bytesWritten < uint64(fileInfo.Size()) {
        adjustedChunkSize := uint64Min(uint64(fileInfo.Size()) - bytesWritten, ChunkSize)
        chunkBuffer := make([]byte, adjustedChunkSize)
        chunkBytesWritten, _ := file.ReadAt(chunkBuffer, int64(bytesWritten))
        bytesWritten += uint64(chunkBytesWritten)
        peer.conn.Write(chunkBuffer)
    }

    progressBar.Stop(true)
    fmt.Print("Verifying file... ")

    switch <- msgCh {
    case ChecksumMatch:
        fmt.Println("done")
    case ChecksumMismatch:
        fmt.Fprintln(os.Stderr, "checksum mismatch")
        os.Exit(1)
    default:
        fmt.Fprintln(os.Stderr, "protocol error")
        os.Exit(1)
    }
}

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
    "github.com/transhift/common/common"
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

func (DownloadPeer) PunchHole(peerUid string, cert tls.Certificate, config common.Config) (remoteAddr string, err error) {
    conn, err := tls.Dial("tcp", net.JoinHostPort(config["puncher_host"], config["puncher_port"]), &tls.Config{
        Certificates: []tls.Certificate{cert},
        InsecureSkipVerify: true,
        MinVersion: tls.VersionTLS12,
    })

    if err != nil {
        return "", err
    }

    defer conn.Close()

    var buffer bytes.Buffer

    if _, err := buffer.Write(common.Mtob(common.UploadClientType)); err != nil {
        return "", err
    }

    if _, err := buffer.WriteString(peerUid); err != nil {
        return "", err
    }

    if _, err := conn.Write(buffer.Bytes()); err != nil {
        return "", err
    }

    scanner := bufio.NewScanner(bufio.NewReader(conn))

    scanner.Split(bufio.ScanBytes)

    for {
        if ! scanner.Scan() {
            return "", scanner.Err()
        }

        puncherResponse := scanner.Bytes()[0]

        switch puncherResponse {
        case common.PuncherPing:
            // TODO: error check
            conn.Write(common.Mtob(common.PuncherPong))
        case common.PuncherEndPing:
            break
        default:
            return "", fmt.Errorf("protocol error: expected one of valid responses, got 0x%X", puncherResponse)
        }
    }

    scanner.Split(bufio.ScanLines)

    if ! scanner.Scan() {
        return "", scanner.Err()
    }

    remoteAddr = scanner.Text()

    scanner.Split(bufio.ScanBytes)

    for {
        if ! scanner.Scan() {
            return "", scanner.Err()
        }

        puncherResponse := scanner.Bytes()[0]

        switch puncherResponse {
        case common.PuncherReady:
            break
        case common.PuncherNotReady:
            return "", fmt.Errorf("peer disconnected")
        default:
            return "", fmt.Errorf("protocol error: expected one of valid responses, got 0x%X", puncherResponse)
        }
    }

    return
}

func (p *DownloadPeer) Connect(cert tls.Certificate, remoteAddr string) error {
    for p.conn == nil {
        var err error
        p.conn, err = tls.Dial("tcp", remoteAddr, &tls.Config{
            Certificates: []tls.Certificate{cert},
            InsecureSkipVerify: true,
            MinVersion: tls.VersionTLS12,
        })

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

func (p DownloadPeer) ReceiveMessages() (ch chan common.ProtocolMessage) {
    ch = make(chan common.ProtocolMessage)
    scanner := bufio.NewScanner(p.inOut.Reader)

    scanner.Split(bufio.ScanBytes)

    go func() {
        for scanner.Scan() {
            ch <- common.ProtocolMessage(scanner.Bytes()[0])
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

    if len(args.peerUid) != common.UidLength {
        fmt.Fprintf(os.Stderr, "Peer UID should be %d characters\n", common.UidLength)
        os.Exit(1)
    }

    peer := DownloadPeer{}
    storage := &common.Storage{
        CustomDir: args.appDir,
        Config: common.Config{
            "puncher_host": "104.236.76.95",
            "puncher_port": "50977",
        },
    }
    err := storage.LoadConfig()

    if err != nil {
        fmt.Fprintln(os.Stderr, err)
        os.Exit(1)
    }

    cert, err := storage.Certificate(CertFileName, KeyFileName)

    if err != nil {
        fmt.Fprintln(os.Stderr, err)
        os.Exit(1)
    }

    fmt.Print("Waiting for peer... ")

    remoteAddr, err := peer.PunchHole(args.peerUid, cert, storage.Config)

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

    checksum, err := common.CalculateFileChecksum(file)

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
    case common.ChecksumMatch:
        fmt.Println("done")
    case common.ChecksumMismatch:
        fmt.Fprintln(os.Stderr, "checksum mismatch")
        os.Exit(1)
    default:
        fmt.Fprintln(os.Stderr, "protocol error")
        os.Exit(1)
    }
}

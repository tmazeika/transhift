package transhift

import (
    "github.com/codegangsta/cli"
    "path/filepath"
    "net"
    "bufio"
    "fmt"
    "os"
    "crypto/tls"
    "time"
    "github.com/transhift/common/common"
)

type UploadArgs struct {
    uid      string
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

func (p *DownloadPeer) ConnectToPuncher(cert tls.Certificate, host string, port string) (err error) {
    p.conn, err = tls.Dial("tcp", net.JoinHostPort(host, port), &tls.Config{
        Certificates: []tls.Certificate{cert},
        InsecureSkipVerify: true,
        MinVersion: tls.VersionTLS12,
    })

    return
}

func (p *DownloadPeer) PunchHole(uid string) (remoteAddr string, err error) {
    defer p.conn.Close()

    in, out := common.MessageChannel(p.conn)

    // Send client type.
    out <- common.NewMesssageWithByte(common.ClientType, common.UploaderClientType)

    // Send uid.
    out <- common.Message{
        Packet: common.UidRequest,
        Body:   []byte(uid),
    }

    // Wait for PeerReady (or PeerNotFound).
    msg, ok := <- in

    if ! ok {
        // Some IO error occurred, so shut down the connection.
        return "", handleError(p.conn, out, true, "closing connection")
    }

    switch msg.Packet {
    case common.PeerReady:
        // If peer is ready, set the remote address.
        remoteAddr = string(msg.Body)
    case common.PeerNotFound:
        // If peer was not found, return an error. The puncher shuts down the
        // connection itself, so no need to #handleError
        return "", fmt.Errorf("peer not found")
    default:
        // If the packet was not recognized, return an error and shut down the
        // connection.
        return "", handleError(p.conn, out, false, "expected peer status, got 0x%x", msg.Packet)
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

func handleError(conn net.Conn, out chan common.Message, internal bool, format string, a ...interface{}) (err error) {
    var packet common.Packet
    err = fmt.Errorf(format, a)

    if internal {
        packet = common.InternalError
    } else {
        packet = common.ProtocolError
    }

    out <- common.Message{
        Packet: packet,
        Body:   []byte(err),
    }

    return
}

func logAndExit(err error) {
    fmt.Fprintln(os.Stderr, err)
    os.Exit(1)
}

func Upload(c *cli.Context) {
    args := UploadArgs{
        uid:      c.Args()[0],
        filePath: c.Args()[1],
        appDir:   c.GlobalString("app-dir"),
    }

    peer := &DownloadPeer{}
    storage := &common.Storage{
        CustomDir: args.appDir,

        // This is the default config. If it already exists, this means nothing.
        Config: common.Config{
            "puncher_host": "104.236.76.95",
            "puncher_port": "50977",
        },
    }
    err := storage.LoadConfig()

    if err != nil {
        logAndExit(err)
    }

    cert, err := storage.Certificate(CertFileName, KeyFileName)

    if err != nil {
        logAndExit(err)
    }

    fmt.Print("Waiting for peer... ")

    err = peer.ConnectToPuncher(cert, storage.Config["puncher_host"], storage.Config["puncher_port"])

    if err != nil {
        logAndExit(err)
    }

    remoteAddr, err := peer.PunchHole(args.uid)

    if err != nil {
        logAndExit(err)
    }

    fmt.Println("done")
    fmt.Printf("Connecting to '%s'... ", args.uid)

    err = peer.Connect(cert, remoteAddr)

    if err != nil {
        logAndExit(err)
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

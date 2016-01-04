package transhift

import (
    "github.com/codegangsta/cli"
    "path/filepath"
    "net"
    "fmt"
    "os"
    "crypto/tls"
    "time"
    "github.com/transhift/common/common"
    "errors"
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
    InOut

    cert *tls.Certificate
    conn *tls.Conn
    addr string
}

func (p *DownloadPeer) connectToPuncher(host string, port string) (err error) {
    p.conn, err = tls.Dial("tcp", net.JoinHostPort(host, port), &tls.Config{
        Certificates: []tls.Certificate{p.cert},
        InsecureSkipVerify: true,
        MinVersion: tls.VersionTLS12,
    })

    p.in, p.out = common.MessageChannel(p.conn)
    return
}

func (p *DownloadPeer) sendClientType() error {
    p.out.Ch <- common.Message{common.Uploader, nil}
    <- p.out.Done
    return p.out.Err
}

func (p *DownloadPeer) sendUid(uid string) error {
    p.out.Ch <- common.Message{common.UidRequest, []byte(uid)}
    <- p.out.Done
    return p.out.Err
}

func (p *DownloadPeer) receiveAddr() error {
    msg, ok := <- p.in.Ch

    if ! ok {
        return p.in.Err
    }

    switch msg.Packet {
    case common.PeerReady:
        p.addr = string(msg.Body)
    case common.PeerNotFound:
        return fmt.Errorf("peer not found")
    default:
        return fmt.Errorf("expected peer status, got 0x%x", msg.Packet)
    }

    return nil
}

func (p *DownloadPeer) PunchHole(host, port, uid string) error {
    err := p.connectToPuncher(host, port)

    if err != nil {
        return  err
    }

    defer p.conn.Close()

    // Send client type.
    err = p.sendClientType()

    if err != nil {
        return err
    }

    // Send uid.
    err = p.sendUid(uid)

    if err != nil {
        return err
    }

    // Expect peer address.
    return p.receiveAddr()
}

func (p *DownloadPeer) connectToPeer() error {
    // TODO: timeout

    for {
        conn, err := tls.Dial("tcp", p.addr, &tls.Config{
            Certificates: []tls.Certificate{p.cert},
            InsecureSkipVerify: true,
            MinVersion: tls.VersionTLS12,
        })

        if err != nil {
            time.Sleep(time.Second)
        } else {
            p.conn = conn
            break
        }
    }

    return nil
}

func (p *DownloadPeer) Connect() error {
    err := p.connectToPeer()

    if err != nil {
        return err
    }

    return CheckCompatibility(p)
}

func Upload(c *cli.Context) {
    args := UploadArgs{
        uid:      c.Args()[0],
        filePath: c.Args()[1],
        appDir:   c.GlobalString("app-dir"),
    }

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
        common.LogAndExit(err)
    }

    cert, err := storage.Certificate(CertFileName, KeyFileName)

    if err != nil {
        common.LogAndExit(err)
    }

    peer := &DownloadPeer{
        cert: cert,
    }

    fmt.Print("Getting peer address... ")

    err = peer.PunchHole(storage.Config["puncher_host"], storage.Config["puncher_port"], args.uid)

    if err != nil {
        HandleError(peer, err)
    }

    fmt.Println("done")
    fmt.Print("Connecting... ")

    err = peer.Connect()

    if err != nil {
        HandleError(peer, err)
    }

    defer peer.conn.Close()

    fmt.Println("done")
    fmt.Print("Sending file info... ")

    file, err := os.Open(args.filePath)

    if err != nil {
        HandleError(peer, nil)
        common.LogAndExit(err)
    }

    defer file.Close()

    fileInfo, err := file.Stat()

    if err != nil {
        logAndExit(err)
    }

    out <- common.Message{
        Packet: common.FileName,
        Body:   []byte(filepath.Base(file.Name())),
    }

    fileHash, err := common.CalculateFileChecksum(file)

    if err != nil {
        logAndExit(err)
    }

    out <- common.Message{
        Packet: common.FileSize,
        Body:   []byte{uint64(fileInfo.Size())},
    }

    out <- common.Message{
        Packet: common.FileHash,
        Body:   fileHash,
    }

    fmt.Println("done")
    fmt.Printf("Uploading '%s'...\n", args.AbsFilePath())

    // TODO: remove (placeholder)
    <- in

    // TODO: redo
    /*var bytesWritten uint64
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
    }*/
}

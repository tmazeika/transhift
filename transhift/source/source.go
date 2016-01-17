package source

import (
    "github.com/codegangsta/cli"
    "path/filepath"
    "net"
    "fmt"
    "os"
    "crypto/tls"
    "time"
    "github.com/transhift/common/common"
    "github.com/transhift/transhift/common/storage"
    "log"
    "github.com/transhift/transhift/transhift/tstorage"
    "github.com/transhift/transhift/transhift/tprotocol"
)

func (p *DownloadPeer) SendFileInfo(fileInfo os.FileInfo, hash []byte) error {
    p.out.Ch <- common.Message{
        Packet: common.FileName,
        Body:   []byte(filepath.Base(fileInfo.Name())),
    }
    <- p.out.Done

    if p.out.Err != nil {
        return p.out.Err
    }

    p.out.Ch <- common.Message{
        Packet: common.FileSize,
        Body:   uint64ToBytes(uint64(fileInfo.Size())),
    }
    <- p.out.Done

    if p.out.Err != nil {
        return p.out.Err
    }

    p.out.Ch <- common.Message{
        Packet: common.FileHash,
        Body:   []byte(hash),
    }
    <- p.out.Done

    return p.out.Err
}

type args struct {
    id       string
    filePath string
    appDir   string
}

func Start(c *cli.Context) {
    log.SetOutput(os.Stdout)
    log.SetFlags(0)

    a := args{
        id:       c.Args()[0],
        filePath: c.Args()[1],
        appDir:   c.GlobalString("app-dir"),
    }

    s, err := tstorage.Prepare(a.appDir)
    if err != nil {
        log.SetOutput(os.Stderr)
        log.Fatalln("Error:", err)
    }

    conf, err := s.Config().(tstorage.Config)
    if err != nil {
        log.SetOutput(os.Stderr)
        log.Fatalln("Error:", err)
    }

    cert, err := s.Certificate(tstorage.KeyFileName, tstorage.CertFileName)
    if err != nil {
        log.SetOutput(os.Stderr)
        log.Fatalln("Error:", err)
    }

    log.Print("Getting peer address... ")

    // Punch TCP hole.
    targetAddr, err := punchHole(conf.Puncher["host"], conf.Puncher["port"], cert, a.id)
    if err != nil {
        log.SetOutput(os.Stderr)
        log.Fatalln("error:", err)
    }

    log.Println("done")
    log.Print("Connecting... ")

    // Connect to peer.
    peer := tprotocol.NewPeer(targetAddr)
    if err := peer.Connect(); err != nil {
        log.SetOutput(os.Stderr)
        log.Fatalln("error:", err)
    }
    defer peer.Close()

    log.Println("done")
    log.Println("Sending file info...")

    file, info, err := getFile(args.filePath)
    if err != nil {
        log.SetOutput(os.Stderr)
        log.Fatalln("error:", err)
    }

    // Send file info.

////////////////////////////////////////////////////////////////////////////////

    defer peer.conn.Close()

    fmt.Println("done")
    fmt.Print("Sending file info... ")

    file, err := os.Open(args.filePath)

    if err != nil {
        fmt.Fprintln(os.Stderr, err)
        return
    }

    defer file.Close()

    fileInfo, err := file.Stat()

    if err != nil {
        fmt.Fprintln(os.Stderr, err)
        return
    }

    fileHash, err := common.CalculateFileHash(file)

    if err != nil {
        fmt.Fprintln(os.Stderr, err)
        return
    }

    err = peer.SendFileInfo(fileInfo, fileHash)

    if err != nil {
        fmt.Fprintln(os.Stderr, err)
        return
    }

    fmt.Println("done")
    fmt.Printf("Uploading %s ...\n", args.AbsFilePath())

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

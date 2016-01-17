package download

import (
    "github.com/codegangsta/cli"
//    "github.com/transhift/common/common"
//    "net"
//    "bufio"
//    "bytes"
//    "fmt"
//    "os"
//    "crypto/tls"
)

/*type DownloadArgs struct {
    destination string
    appDir      string
}

func (a DownloadArgs) DestinationOrDef(def string) string {
    if len(a.destination) == 0 {
        return def
    }

    return a.destination
}

type UploadPeer struct {
    conn     net.Conn
    inOut    *bufio.ReadWriter
    fileInfo FileInfo
}

func (UploadPeer) PunchHole(cert tls.Certificate, config common.Config) (uid string, localPort string, err error) {
    conn, err := tls.Dial("tcp", net.JoinHostPort(config["puncher_host"], config["puncher_port"]), &tls.Config{
        Certificates: []tls.Certificate{cert},
        InsecureSkipVerify: true,
        MinVersion: tls.VersionTLS12,
    })

    if err != nil {
        return "", "", err
    }

    defer conn.Close()

    if _, err := conn.Write(common.Mtob(common.DownloadClientType)); err != nil {
        return "", "", err
    }

    uidBuffer := make([]byte, common.UidLength)

    if _, err := conn.Read(uidBuffer); err != nil {
        return "", "", err
    }

    _, localPort, err = net.SplitHostPort(conn.LocalAddr().String())

    if err != nil {
        return "", "", err
    }

    uid = string(uidBuffer)

    scanner := bufio.NewScanner(bufio.NewReader(conn))

    scanner.Split(bufio.ScanBytes)

    for {
        if ! scanner.Scan() {
            return "", "", scanner.Err()
        }

        puncherResponse := scanner.Bytes()[0]

        switch common.ProtocolMessage(puncherResponse) {
        case common.PuncherPing:
            // TODO: error check
            conn.Write(common.Mtob(common.PuncherPong))
        case common.PuncherReady:
            break
        default:
            return "", "", fmt.Errorf("protocol error: expected one of valid responses, got 0x%X", puncherResponse)
        }
    }

    return uid, localPort, nil
}

func (p *UploadPeer) Connect(port string, cert tls.Certificate) error {
    listener, err := tls.Listen("tcp", net.JoinHostPort("", port), &tls.Config{
        Certificates: []tls.Certificate{cert},
        MinVersion: tls.VersionTLS12,
    })

    if err != nil {
        return err
    }

    p.conn, err = listener.Accept()

    if err != nil {
        return err
    }

    p.inOut = bufio.NewReadWriter(bufio.NewReader(p.conn), bufio.NewWriter(p.conn))

    return CheckCompatibility(p.inOut)
}

func (p *UploadPeer) ReceiveFileInfo() error {
    const NLCount = 3
    var buffer bytes.Buffer

    for i := 0; i < NLCount; i++ {
        line, err := p.inOut.ReadBytes('\n')

        if err != nil {
            return err
        }

        buffer.Write(line)
    }

    p.fileInfo = FileInfo{}

    return p.fileInfo.UnmarshalBinary(buffer.Bytes())
}

func (p UploadPeer) ReceiveChunks() (ch chan []byte) {
    ch = make(chan []byte)
    var bytesRead uint64

    go func() {
        for bytesRead < p.fileInfo.size {
            adjustedChunkSize := uint64Min(p.fileInfo.size - bytesRead, ChunkSize)
            chunkBuffer := make([]byte, adjustedChunkSize)
            chunkBytesRead, _ := p.conn.Read(chunkBuffer)
            bytesRead += uint64(chunkBytesRead)
            ch <- chunkBuffer
        }
    }()

    return
}*/

func Download(c *cli.Context) {
    /*args := DownloadArgs{
        destination: c.String("destination"),
        appDir:      c.GlobalString("app-dir"),
    }

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

    peer := UploadPeer{}
    uid, localPort, err := peer.PunchHole(cert, storage.Config)

    if err != nil {
        fmt.Fprintln(os.Stderr, "Unable to retrieve UID")
        os.Exit(1)
    }

    fmt.Printf("Received UID: '%s'\n", uid)
    fmt.Print("Listening for peer... ")

    err = peer.Connect(localPort, cert)

    if err != nil {
        fmt.Fprintln(os.Stderr, err)
        os.Exit(1)
    }

    defer peer.conn.Close()

    fmt.Println("done")
    fmt.Print("Waiting for file info... ")

    err = peer.ReceiveFileInfo()

    if err != nil {
        fmt.Fprintln(os.Stderr, err)
        os.Exit(1)
    }

    fmt.Println("Downloading... ")

    file, err := os.Create(args.DestinationOrDef(peer.fileInfo.name))

    if err != nil {
        fmt.Fprintln(os.Stderr, err)
        os.Exit(1)
    }

    defer file.Close()

    ch := peer.ReceiveChunks()
    var bytesRead uint64
    progressBar := ProgressBar{
        current: &bytesRead,
        total:   peer.fileInfo.size,
    }

    progressBar.Start()

    for bytesRead < peer.fileInfo.size {
        chunk := <- ch
        file.WriteAt(chunk, int64(bytesRead))
        bytesRead += uint64(len(chunk))
    }

    progressBar.Stop(true)
    fmt.Print("Verifying file... ")

    checksum, err := common.CalculateFileChecksum(file)

    if err != nil {
        fmt.Fprintln(os.Stderr, err)
        os.Exit(1)
    }

    if bytes.Equal(checksum, peer.fileInfo.checksum) {
        peer.SendMessage(common.ChecksumMatch)
        fmt.Println("done")
    } else {
        peer.SendMessage(common.ChecksumMismatch)
        fmt.Fprintln(os.Stderr, "checksum mismatch")
        os.Exit(1)
    }*/
}

package transhift

import (
    "github.com/codegangsta/cli"
    "net"
    "bufio"
    "bytes"
    "fmt"
    "os"
    "crypto/tls"
)

type DownloadArgs struct {
    appDir      string
    destination string
}

func (a DownloadArgs) DestinationOrDef(def string) string {
    if len(a.destination) == 0 {
        return def
    }

    return a.destination
}

type UploadPeer struct {
    conn     tls.Conn
    inOut    bufio.ReadWriter
    fileInfo FileInfo
}

func (UploadPeer) PunchHole(config Config) (uid string, localPort string, err error) {
    // TODO: use tls.Dial
    conn, err := net.Dial("tcp", net.JoinHostPort(config.PuncherHost, config.PuncherPortStr()))

    if err != nil {
        return "", "", err
    }

    defer conn.Close()

    if _, err := conn.Write(messageToBytes(DownloadClientType)); err != nil {
        return "", "", err
    }

    uidBuffer := make([]byte, UidLength)

    if _, err := conn.Read(uidBuffer); err != nil {
        return "", "", err
    }

    _, localPort, err = net.SplitHostPort(conn.LocalAddr().String())

    if err != nil {
        return "", "", err
    }

    return string(uidBuffer), localPort, nil
}

func (p *UploadPeer) Connect(port string, cert tls.Certificate) error {
    listener, err := net.Listen("tcp", net.JoinHostPort("", port))

    if err != nil {
        return err
    }

    conn, err := listener.Accept()

    if err != nil {
        return err
    }

    p.conn = *tls.Server(conn, &tls.Config{
        Certificates: []tls.Certificate{cert},
    })
    p.inOut = *bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))

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
}

func (p UploadPeer) SendMessage(msg ProtocolMessage) {
    p.conn.Write(messageToBytes(msg))
}

func Download(c *cli.Context) {
    args := DownloadArgs{
        appDir:      c.GlobalString("app-dir"),
        destination: c.String("destination"),
    }

    storage := &Storage{
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

    peer := UploadPeer{}
    uid, localPort, err := peer.PunchHole(*config)

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

    if bytes.Equal(calculateFileChecksum(file), peer.fileInfo.checksum) {
        peer.SendMessage(ChecksumMatch)
        fmt.Println("done")
    } else {
        peer.SendMessage(ChecksumMismatch)
        fmt.Fprintln(os.Stderr, "checksum mismatch")
        os.Exit(1)
    }
}

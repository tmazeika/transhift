package transhift

import (
    "github.com/codegangsta/cli"
    "net"
    "bufio"
    "bytes"
    "fmt"
    "os"
)

type DownloadArgs struct {
    appDir      string
    destination string
}

func (a DownloadArgs) DestinationOrDef(def string) string {
    if a.destination == "" {
        return def
    }
    return a.destination
}

type UploadPeer struct {
    conn     net.Conn
    reader   *bufio.Reader
    writer   *bufio.Writer
    metaInfo *ProtoMetaInfo
}

func (p *UploadPeer) PunchHole(config *Config) (uid, localPort string, err error) {
    conn, err := net.Dial("tcp", net.JoinHostPort(config.PuncherHost, config.PuncherPortStr()))

    if err != nil {
        return "", "", err
    }

    defer conn.Close()
    conn.Write([]byte{byte(ProtoMsgClientTypeDL)})
    uidBuffer := make([]byte, ProtoPeerUIDLen)
    conn.Read(uidBuffer)

    localAddr := conn.LocalAddr().String()
    _, port, err := net.SplitHostPort(localAddr)

    if err != nil {
        return "", "", err
    }

    return string(uidBuffer), port, nil
}

func (p *UploadPeer) Connect(port string) error {
    listener, err := net.Listen("tcp", net.JoinHostPort("", port))

    if err != nil {
        return err
    }

    p.conn, err = listener.Accept()

    if err != nil {
        return err
    }

    p.reader = bufio.NewReader(p.conn)
    p.writer = bufio.NewWriter(p.conn)

    err = checkCompatibility(p.reader, p.writer)

    if err != nil {
        return err
    }

    return nil
}

func (p *UploadPeer) ReceiveMetaInfo() {
    const ExpectedNLCount = 3

    var buffer bytes.Buffer

    for i := 0; i < ExpectedNLCount; i++ {
        line, err := p.reader.ReadBytes('\n')

        if err != nil {
            fmt.Fprintln(os.Stderr, err)
            os.Exit(1)
        }

        buffer.Write(line)
    }

    p.metaInfo = &ProtoMetaInfo{}
    p.metaInfo.Deserialize(buffer.Bytes())
}

func (p *UploadPeer) ReceiveChunks() chan []byte {
    ch := make(chan []byte)
    var bytesRead uint64

    go func() {
        for bytesRead < p.metaInfo.fileSize {
            adjustedChunkSize := uint64Min(p.metaInfo.fileSize - bytesRead, ProtoChunkSize)
            chunkBuffer := make([]byte, adjustedChunkSize)
            chunkBytesRead, _ := p.reader.Read(chunkBuffer)
            bytesRead += uint64(chunkBytesRead)
            ch <- chunkBuffer
        }
    }()

    return ch
}

func (p *UploadPeer) SendMessage(message ProtoMsg) {
    p.conn.Write([]byte{byte(message)})
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

    _, err = storage.PrivKey()

    if err != nil {
        fmt.Fprintln(os.Stderr, err)
        os.Exit(1)
    }

    peer := &UploadPeer{}
    uid, localPort, err := peer.PunchHole(config)

    if err != nil {
        fmt.Fprintln(os.Stderr, "Unable to retrieve UID")
        os.Exit(1)
    }

    fmt.Printf("UID: '%s'\n", uid)
    fmt.Print("Listening for peer... ")

    err = peer.Connect(localPort)

    if err != nil {
        fmt.Fprintln(os.Stderr, err)
        os.Exit(1)
    }

    defer peer.conn.Close()

    fmt.Println("done")
    fmt.Print("Waiting for file info... ")
    peer.ReceiveMetaInfo()
    fmt.Println("Downloading... ")
    file, err := os.Create(args.DestinationOrDef(peer.metaInfo.fileName))
    defer file.Close()

    if err != nil {
        fmt.Fprintln(os.Stderr, err)
        os.Exit(1)
    }

    ch := peer.ReceiveChunks()
    var bytesRead uint64
    progressBar := ProgressBar{
        current: &bytesRead,
        total:   peer.metaInfo.fileSize,
    }
    progressBar.Start()

    for bytesRead < peer.metaInfo.fileSize {
        chunk := <- ch
        file.WriteAt(chunk, int64(bytesRead))
        bytesRead += uint64(len(chunk))
    }

    progressBar.Stop(true)
    fmt.Print("Verifying file... ")

    if bytes.Equal(calculateFileChecksum(file), peer.metaInfo.fileChecksum) {
        peer.SendMessage(ProtoMsgChecksumMatch)
        fmt.Println("done")
    } else {
        peer.SendMessage(ProtoMsgChecksumMismatch)
        fmt.Fprintln(os.Stderr, "checksum mismatch")
        os.Exit(1)
    }
}

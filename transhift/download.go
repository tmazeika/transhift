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
    password    string
    destination string
}

func (a DownloadArgs) PasswordHash() []byte {
    return calculateStringChecksum(a.password)
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

func (p *UploadPeer) Connect(args DownloadArgs) error {
    listener, err := net.Listen("tcp", net.JoinHostPort("", ProtoPortStr))

    if err != nil {
        return err
    }

    p.conn, err = listener.Accept()

    if err != nil {
        return err
    }

    p.reader = bufio.NewReader(p.conn)
    p.writer = bufio.NewWriter(p.conn)
    return nil
}

func (p *UploadPeer) ReceiveMetaInfo() {
    const ExpectedNLCount = 4

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

func (p *UploadPeer) SendMessage(message byte) {
    p.writer.WriteByte(message)
    p.writer.Flush()
}

func Download(c *cli.Context) {
    args := DownloadArgs{
        password:    c.Args()[0],
        destination: c.String("destination"),
    }

    peer := &UploadPeer{}
    fmt.Print("Connecting to peer... ")
    err := peer.Connect(args)

    if err != nil {
        fmt.Fprintln(os.Stderr, err)
        os.Exit(1)
    }

    peer.ReceiveMetaInfo()

    // verify password
    if bytes.Equal(args.PasswordHash(), peer.metaInfo.passwordChecksum) {
        peer.SendMessage(ProtoMsgPasswordMatch)
        fmt.Println("done")
    } else {
        peer.SendMessage(ProtoMsgPasswordMismatch)
        fmt.Fprintln(os.Stderr, "password mismatch")
        os.Exit(1)
    }

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

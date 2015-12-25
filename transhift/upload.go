package transhift

import (
    "github.com/codegangsta/cli"
    "path/filepath"
    "net"
    "bufio"
    "fmt"
    "os"
    "bytes"
)

type UploadArgs struct {
    peerHost string
    password string
    filePath string
}

func (a *UploadArgs) PasswordHash() []byte {
    return calculateStringHash(a.password)
}

func (a *UploadArgs) AbsFilePath() string {
    filePath, _ := filepath.Abs(a.filePath)
    return filePath
}

type DownloadPeer struct {
    conn     net.Conn
    reader   *bufio.Reader
    writer   *bufio.Writer
    metaInfo *ProtoMetaInfo
}

func (p *DownloadPeer) Connect(args UploadArgs) {
    for p.conn == nil {
        p.conn, _ = net.Dial("tcp", net.JoinHostPort(args.peerHost, ProtoPortStr))
    }

    p.reader = bufio.NewReader(p.conn)
    p.writer = bufio.NewWriter(p.conn)
}

func (p *DownloadPeer) ReceiveMetaInfo() {
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

func (p *DownloadPeer) ReceiveChunks() chan *ProtoChunk {
    ch := make(chan *ProtoChunk)
    var bytesRead uint64

    go func() {
        for bytesRead < p.metaInfo.fileSize {
            adjustedChunkSize := uint64Min(p.metaInfo.fileSize - bytesRead, ProtoChunkSize)
            chunkBuffer := make([]byte, adjustedChunkSize)
            var chunkBytesRead uint64

            for chunkBytesRead < adjustedChunkSize {
                subChunkBytesRead, _ := p.reader.Read(chunkBuffer[chunkBytesRead:])
                chunkBytesRead += subChunkBytesRead
            }

            bytesRead += adjustedChunkSize
            chunk := &ProtoChunk{}
            chunk.Deserialize(chunkBuffer)
            ch <- &chunk
        }
    }()

    return ch
}

func Upload(c *cli.Context) {
    args := UploadArgs{
        peerHost: c.Args()[0],
        password: c.Args()[1],
        filePath: c.Args()[2],
    }
}

func uint64Min(x, y uint64) uint64 {
    if x < y {
        return x
    }
    return y
}

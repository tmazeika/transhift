package transhift

import (
    "github.com/codegangsta/cli"
    "fmt"
    "os"
//    "encoding/hex"
    "time"
    "bytes"
)

func Upload(c *cli.Context) {
    host := c.Args()[0]
    pass := c.Args()[1]
    filePath := c.Args()[2]

    downloadPeer := &DownloadPeer{}

    file, err := os.Open(filePath)

    if err != nil {
        fmt.Fprintln(os.Stderr, err)
        os.Exit(1)
    }

    fileInfo, err := file.Stat()

    if err != nil {
        fmt.Fprintln(os.Stderr, err)
        os.Exit(1)
    }

    fmt.Printf("Connecting to %s... ", host)
    downloadPeer.Connect(host, port)

    resCh := downloadPeer.ProtocolResponseChannel()

    downloadPeer.SendMetaInfo(&MetaInfo{
        passHash: stringChecksum(pass),
        fileName: fileInfo.Name(),
        fileSize: uint64(fileInfo.Size()),
        fileHash: fileChecksum(file),
    })

    if <- resCh == PasswordMismatch {
        fmt.Fprintln(os.Stderr, "password mismatch")
        os.Exit(1)
    }

    fmt.Println("done")

    var currentBytes uint64
    totalBytes := uint64(fileInfo.Size())
    progressBarStopCh := showProgressBar(&currentBytes, totalBytes)

    for currentBytes < totalBytes {
        chunkBuff := make([]byte, chunkSize)
        chunkBuffSize, err := file.ReadAt(chunkBuff, int64(totalBytes))

        if err != nil {
            progressBarStopCh <- false
            fmt.Println('\n', err)
            os.Exit(1)
        }

        downloadPeer.SendFileChunk(&FileChunk{
            good: true,
            data: chunkBuff[:chunkBuffSize],
        })

        currentBytes += uint64(chunkBuffSize)
    }

    progressBarStopCh <- true

    fmt.Println("\nUploaded")
    fmt.Println("Awaiting peer response... ")

    if <- resCh == ChecksumMismatch {
        fmt.Fprintln(os.Stderr, "checksum mismatch; the peer's file may be corrupt and/or incomplete")
        os.Exit(1)
    }

    fmt.Println("done")
}

func showProgressBar(current *uint64, total uint64) (stopCh chan bool) {
    stopCh = make(chan bool)
    stop := false

    update := func() {
        var buff bytes.Buffer
        percent := float64(*current) / float64(total)

        buff.WriteString(fmt.Sprintf("\r%.f%% [", percent))

        const BarSize = float64(50)

        for i := float64(0); i < percent * BarSize; i++ {
            buff.WriteRune('=')
        }

        for i := float64(0); i < BarSize - percent * BarSize; i++ {
            buff.WriteRune(' ')
        }

        buff.WriteString(fmt.Sprintf("] %s/%s", formatSize(*current), formatSize(total)))
        fmt.Print(buff.String())
    }

    go func() {
        for ! stop && *current < total {
            update()
            time.Sleep(time.Second)
        }
    }()

    go func() {
        updateAfterStop := <- stopCh
        stop = true

        if updateAfterStop {
            update()
        }
    }()

    return
}

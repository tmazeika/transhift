package transhift

import (
    "github.com/codegangsta/cli"
    "fmt"
    "os"
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

    fmt.Printf("Connecting to %s", host)
    downloadPeer.Connect(host, port)

    resCh := downloadPeer.ProtocolResponseChannel()

    downloadPeer.SendMetaInfo(&MetaInfo{
        passHash: stringChecksum(pass),
        fileName: fileInfo.Name(),
        fileSize: uint64(fileInfo.Size()),
        fileHash: fileChecksum(file),
    })

    if <- resCh == PasswordMismatch {
        fmt.Fprintln(os.Stderr, " password mismatch")
        os.Exit(1)
    }

    fmt.Println(" done")

    var currentBytes uint64
    totalBytes := uint64(fileInfo.Size())
    progressBarStopCh := showProgressBar(&currentBytes, totalBytes)

    for currentBytes < totalBytes {
        adjustedChunkSize := min(totalBytes - currentBytes, chunkSize)

        chunkBuff := make([]byte, adjustedChunkSize)
        chunkBuffSize, _ := file.ReadAt(chunkBuff, int64(currentBytes))

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

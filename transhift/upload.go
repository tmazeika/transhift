package transhift

import (
    "github.com/codegangsta/cli"
    "fmt"
    "os"
//    "encoding/hex"
    "time"
)

func Upload(c *cli.Context) {
    ticker := time.Tick(time.Second)
    for i := 10; i > 0; i-- {
        <-ticker
        fmt.Printf("\r %d/10", i)
    }
    fmt.Println("\nAll is said and done.")

    peerHost := c.Args()[0]
    password := c.Args()[1]
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

    fmt.Printf("Connecting to %s... ", peerHost)
    downloadPeer.Connect(peerHost, port)

    resCh := downloadPeer.ProtocolResponseChannel()

    downloadPeer.SendMetaInfo(&MetaInfo{
        // passHash
        passHash: stringChecksum(password),
        // fileName
        fileName: fileInfo.Name(),
        // fileSize
        fileSize: uint64(fileInfo.Size()),
        // fileHash
        fileHash: fileChecksum(file),
    })

    if <- resCh == PasswordMismatch {
        fmt.Fprintln(os.Stderr, "password mismatch")
        os.Exit(1)
    }

    fmt.Println("done")

    fmt.Println("Sending")
}/*

func upHandleFileChunks(downloadPeer *DownloadPeer, upFileInfo *UpFileInfo) (ok bool) {
    fmt.Printf("Sending file '%s' with a size of %s (SHA-256 %s) to %s\n",
        upFileInfo.file.Name(), formatSize(float64(upFileInfo.fileInfo.Size())), hex.EncodeToString(upFileInfo.checksum),
        downloadPeer.conn.RemoteAddr().String())

    var totalWrote uint64
    updateProgressStopSignal := false

    updateProgress(&totalWrote, uint64(upFileInfo.fileInfo.Size()), &updateProgressStopSignal)

    for totalWrote < uint64(upFileInfo.fileInfo.Size()) {
        fileChunk := make([]byte, chunkSize)
        fileChunkRead, err := upFileInfo.file.ReadAt(fileChunk, int64(totalWrote))

        if err != nil {
            fmt.Println("Error: ", err)
            updateProgressStopSignal = true
            return false
        }

        downloadPeer.SendFileChunk(fileChunk)

        totalWrote += uint64(fileChunkRead)
    }

    fmt.Println("Done!")
    return true
}
*/

package transhift

import (
    "github.com/codegangsta/cli"
    "fmt"
    "os"
    "encoding/hex"
)

func Upload(c *cli.Context) {
    peerHost := c.Args()[0]
    password := c.Args()[1]
    filePath := c.Args()[2]

    downloadPeer := DownloadPeer{}

    upHandleConnect(&downloadPeer, peerHost)
    upHandlePasswordHash(&downloadPeer, password)

    resCh := downloadPeer.ProtocolResponseChannel()

    ok := (<- resCh) == PasswordMatch

    if ok {
        fmt.Println("match")
    } else {
        fmt.Println("mismatch. Retry with the correct password!")
        return
    }

    ok, upFileInfo := upHandleFileInfo(&downloadPeer, filePath)

    if ! ok { return }

    if ok := upHandleFileChunks(&downloadPeer, upFileInfo); ! ok { return }
}

func upHandleConnect(downloadPeer *DownloadPeer, peerHost string) {
    fmt.Printf("Connecting to %s... ", peerHost)
    downloadPeer.Connect(peerHost, port)
    fmt.Println("done")
}

func upHandlePasswordHash(downloadPeer *DownloadPeer, password string) {
    fmt.Print("Sending password... ")
    downloadPeer.SendPasswordHash(password)
}

type UpFileInfo struct {
    file     *os.File
    fileInfo os.FileInfo
    checksum []byte
}

func upHandleFileInfo(downloadPeer *DownloadPeer, filePath string) (ok bool, _ *UpFileInfo) {
    upFileInfo := UpFileInfo{}

    fmt.Print("Sending file info... ")
    var err error
    upFileInfo.file, err = os.Open(filePath)

    if err != nil {
        fmt.Println("error: ", err)
        return false, nil
    }

    upFileInfo.checksum, err = fileChecksum(upFileInfo.file)

    if err != nil {
        fmt.Println("error: ", err)
        return false, nil
    }

    upFileInfo.fileInfo, err = upFileInfo.file.Stat()

    if err != nil {
        fmt.Println("error: ", err)
        return false, nil
    }

    downloadPeer.SendMetaInfo(upFileInfo.file.Name(), uint64(upFileInfo.fileInfo.Size()), upFileInfo.checksum)
    return true, &upFileInfo
}

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

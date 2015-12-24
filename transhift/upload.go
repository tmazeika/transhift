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
    resCh := downloadPeer.ProtocolResponseChannel()

    upHandleConnect(&downloadPeer, peerHost)
    upHandlePasswordHash(&downloadPeer, password)

    ok := (<- resCh) == PasswordMatch

    if ok {
        fmt.Println("match")
    } else {
        fmt.Println("mismatch. Retry with the correct password!")
        return
    }

    ok, file, fileInfo, checksum := upHandleFileInfo(&downloadPeer, filePath)

    if ! ok { return }

    if ok := upHandleFileChunks(&downloadPeer, file, fileInfo, checksum); ! ok { return }
}

func upHandleConnect(downloadPeer *DownloadPeer, peerHost string) {
    fmt.Printf("Connecting to %s... ", peerHost)
    downloadPeer.Connect(peerHost, port)
    fmt.Print("done")
}

func upHandlePasswordHash(downloadPeer *DownloadPeer, password string) {
    fmt.Print("Sending password... ")
    downloadPeer.SendPasswordHash(password)
}

func upHandleFileInfo(downloadPeer *DownloadPeer, filePath string) (ok bool, file os.File, fileInfo os.FileInfo, checksum []byte) {
    fmt.Print("Sending file info... ")
    file, err := os.Open(filePath)

    if err != nil {
        fmt.Println("error: ", err)
        return false
    }

    fileHash, err := fileChecksum(file)

    if err != nil {
        fmt.Println("error: ", err)
        return false
    }

    fileInfo, err = file.Stat()

    if err != nil {
        fmt.Println("error: ", err)
        return false
    }

    downloadPeer.SendFileInfo(fileInfo.Name(), uint64(fileInfo.Size()), fileHash)
    return true, file, fileInfo, fileHash
}

func upHandleFileChunks(downloadPeer *DownloadPeer, file os.File, fileInfo os.FileInfo, checksum []byte) (ok bool) {
    fmt.Printf("Sending file '%s' with a size of %s (SHA-256 %s) to %s\n",
        file.Name(), formatSize(fileInfo.Size()), hex.EncodeToString(checksum),
        downloadPeer.conn.RemoteAddr().String())

    var totalWrote uint64
    updateProgressStopSignal := false

    updateProgress(&float64(totalWrote), float64(fileInfo.Size()), &updateProgressStopSignal)

    for totalWrote < fileInfo.Size() {
        fileChunk := make([]byte, chunkSize)
        fileChunkRead, err := file.ReadAt(fileChunk, int64(totalWrote))

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

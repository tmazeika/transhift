package transhift

import (
    "github.com/codegangsta/cli"
    "net"
    "time"
    "fmt"
    "os"
    "bytes"
    "encoding/binary"
    "path/filepath"
)

func Upload(c *cli.Context) {
    peerHost := c.Args()[0]
    password := c.Args()[1]
    filePath := c.Args()[2]

    downloadPeer := DownloadPeer{}
    resCh := downloadPeer.ProtocolResponseChannel()

    upHandleConnect(downloadPeer, peerHost)
    upHandlePasswordHash(downloadPeer, password)

    ok := (<- resCh) == PasswordMatch

    if ok {
        fmt.Println("match")
    } else {
        fmt.Println("mismatch. Retry with the correct password!")
        return
    }

    if ok := upHandleFileInfo(downloadPeer, filePath); ! ok { return }
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

func upHandleFileInfo(downloadPeer *DownloadPeer, filePath string) (ok bool) {
    fmt.Print("Sending file info... ")
    file, err := os.Open(filePath)

    if err != nil {
        fmt.Println("error: ", err)
        return false
    }

    fileHash, err := fileChecksum(file)

    if err != nil {
        fmt.Println("error: ", err)
    }

    fileInfo, err := file.Stat()

    if err != nil {
        fmt.Println("error: ", err)
    }

    downloadPeer.SendFileInfo(fileInfo.Name(), uint64(fileInfo.Size()), fileHash)
    return true
}

func connect(peer, password, filePath string) {
    var conn net.Conn

    fmt.Print("Dialing peer")

    for {
        // dial the peer
        conn, _ = net.Dial("tcp", net.JoinHostPort(peer, portStr))

        if (conn == nil) {
            fmt.Print(".")
            // wait to try again
            time.Sleep(time.Second * 2)
        } else {
            break
        }
    }

    defer conn.Close()

    fmt.Println("\nConnected to peer")
    send(conn, password, filePath)
}

func send(conn net.Conn, password, filePath string) {
    // write password
    conn.Write(bytesWithLen([]byte(password)))

    passwordResBuffer := make([]byte, 1)
    _, err := conn.Read(passwordResBuffer)
    check(err)

    if (passwordResBuffer[0] == passwordBad) {
        fmt.Println("Sent wrong password")
        conn.Close()
        return
    } else {
        fmt.Println("Password verified")
    }

    // open file
    file, err := os.Open(filePath)
    check(err)
    defer file.Close()

    // stat file
    fileInfo, err := file.Stat()
    check(err)

    fileName := fileInfo.Name()
    var fileSize uint64 = uint64(fileInfo.Size())

    // write fileName data
    conn.Write(bytesWithLen([]byte(fileName)))

    // write fileSize data
    conn.Write(bytesWithLen(uint64ToBytes(uint64(fileSize))))

    // calculate/write SHA256 checksum
    sum := checksum(filePath)
    conn.Write(bytesWithLen(sum))

    path, err := filepath.Abs(fileName)
    check(err)

    fmt.Printf("Uploading file '%s' with a size of %s\n", path, formatSize(fileSize))

    var totalBytesRead uint64
    doProgressPrint := true

    updateProgress(&totalBytesRead, &fileSize, &doProgressPrint)

    // start reading/sending file bytes
    for totalBytesRead < uint64(fileSize) {
        // read a chunk of *up to* chunkSize bytes from file
        fileBuffer := make([]byte, chunkSize)
        fileBytesRead, err := file.ReadAt(fileBuffer, int64(totalBytesRead))

        if len(fileBuffer) == fileBytesRead {
            check(err)
        }

        totalBytesRead += uint64(fileBytesRead)

        // send data
        conn.Write(bytesWithLen(fileBuffer[:fileBytesRead]))
    }

    fmt.Println("File uploaded")
}

func bytesWithLen(b []byte) []byte {
    var buffer bytes.Buffer
    buffer.Write(uint64ToBytes(uint64(len(b))))
    buffer.Write(b)
    return buffer.Bytes()
}

func uint64ToBytes(v uint64) []byte {
    b := make([]byte, sizeOfUint64)
    binary.BigEndian.PutUint64(b, v)
    return b
}

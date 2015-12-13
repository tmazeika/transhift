package transhift

import (
    "github.com/codegangsta/cli"
    "net"
    "fmt"
    "encoding/binary"
    "os"
    "bytes"
    "encoding/hex"
    "path/filepath"
)

const (
    sizeOfUint64 = 8
)

func Download(c *cli.Context) {
    listen(c.Args()[0], c.String("destination"))
}

func listen(password, fileName string) {
    // start the TCP listener
    listener, err := net.Listen("tcp", net.JoinHostPort("", portStr))
    check(err)
    defer listener.Close()

    fmt.Println("Listening...")

    // listen for a peer connection
    conn, err := listener.Accept()
    check(err)
    defer conn.Close()

    fmt.Println("Connected to peer")
    incomingChannel := createIncomingChannel(conn)
    receive(conn, incomingChannel, password, fileName)
}

func createIncomingChannel(conn net.Conn) chan []byte {
    incomingChannel := make(chan []byte)

    go func() {
        for {
            // read sizeOfUint64 bytes from the connection
            buffer := make([]byte, sizeOfUint64)
            bytesRead, err := conn.Read(buffer)

            // EOF
            if err != nil {
                break
            }

            if bytesRead != len(buffer) {
                panic(fmt.Sprint("Didn't read %d bytes", sizeOfUint64))
            }

            dataSize := bytesToUint64(buffer)

            // read dataSize bytes from the connection
            dataBuffer := make([]byte, dataSize)
            dataBytesRead, err := conn.Read(dataBuffer)
            check(err)

            if dataBytesRead != len(dataBuffer) {
                panic(fmt.Sprint("Didn't read %d bytes", dataSize))
            }

            incomingChannel <- dataBuffer
        }
    }()

    return incomingChannel
}

func receive(conn net.Conn, incoming chan []byte, password, fileName string) {
    // wait for password
    incomingPassword := string(<-incoming)

    if (incomingPassword != password) {
        fmt.Println("Peer sent wrong password")
        conn.Write([]byte{password_bad})
        conn.Close()
        return
    } else {
        fmt.Println("Password verified")
        conn.Write([]byte{password_good})
    }

    // wait for fileName
    incomingFileName := string(<-incoming)

    // if the user has not specified a specific file name, use the incoming name
    if (fileName == "") {
        fileName = incomingFileName
    }

    // wait for fileSize
    fileSize := bytesToUint64(<-incoming)

    // wait for checksum
    fileSum := <-incoming

    fmt.Printf("Downloading file '%s' with a size of %s (sha256 %s)\n", incomingFileName, formatSize(fileSize), hex.EncodeToString(fileSum))

    // create temporary file to write to
    tmpFileName := fileName + ".tmp"
    file, err := os.Create(tmpFileName)
    check(err)
    defer file.Close()

    var totalBytesReceived uint64
    var bytesSinceSync uint64
    doProgressPrint := true

    updateProgress(&totalBytesReceived, &fileSize, &doProgressPrint)

    save := func() {
        fmt.Print("Saving... ")
        doProgressPrint = false
        err = file.Sync()
        check(err)
        doProgressPrint = true
        fmt.Println("done")
    }

    // start receiving file bytes
    for totalBytesReceived < fileSize {
        fileBuffer := <-incoming
        fileBytesWritten, err := file.WriteAt(fileBuffer, int64(totalBytesReceived))
        check(err)

        if fileBytesWritten != len(fileBuffer) {
            panic(fmt.Sprint("Didn't write %d bytes", len(fileBuffer)))
        }

        fileBufferLen := uint64(len(fileBuffer))

        totalBytesReceived += fileBufferLen
        bytesSinceSync += fileBufferLen
    }

    // do final save
    save()

    // calculate/verify sha256 checksum
    tmpFileSum := checksum(tmpFileName)

    if bytes.Equal(fileSum, tmpFileSum) {
        fmt.Println("Checksums verified")
    } else {
        panic("Checksum mismatch")
    }

    // rename file to correct name
    err = os.Rename(tmpFileName, fileName)
    check(err)

    path, err := filepath.Abs(fileName)
    check(err)

    fmt.Printf("File downloaded at '%s'\n", path)
}

func bytesToUint64(b []byte) uint64 {
    return binary.BigEndian.Uint64(b)
}
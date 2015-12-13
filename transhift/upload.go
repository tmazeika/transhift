package transhift

import (
    "github.com/codegangsta/cli"
    "net"
    "time"
    "fmt"
    "os"
    "bytes"
    "encoding/binary"
)

func Upload(c *cli.Context) {
    connect(c.Args()[0], c.Args()[1], c.Args()[2])
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

    if (passwordResBuffer[0] == password_bad) {
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

    var totalBytesRead uint64
    doProgressPrint := true

    updateProgress(&totalBytesRead, &fileSize, &doProgressPrint)

    // start reading/sending file bytes
    for totalBytesRead < uint64(fileSize) {
        // read a chunk of *up to* chunkSize bytes from file
        fileBuffer := make([]byte, chunkSize)
        fileBytesRead, err := file.ReadAt(fileBuffer, int64(totalBytesRead))
        check(err)

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

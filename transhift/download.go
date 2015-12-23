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
    "github.com/huin/goupnp/dcps/internetgateway2"
)

const (
    sizeOfUint64 = 8
)

func Download(c *cli.Context) {
    fmt.Print("Preparing for listen... ")

    var ipStr string

    ifaces, err := net.Interfaces()
    check(err)
    for _, i := range ifaces {
        addrs, err := i.Addrs()
        check(err)

        for _, addr := range addrs {
            var ip net.IP

            switch v := addr.(type) {
                case *net.IPNet:
                    ip = v.IP
                case *net.IPAddr:
                    ip = v.IP
            }

            if (! ip.IsLoopback() && ip.To4() != nil) {
                ipStr = ip.String()
            }
        }
    }

    clients, _, err := internetgateway2.NewWANIPConnection1Clients()
    check(err)

    if (len(clients) > 0) {
        err = clients[0].AddPortMapping("", port, "tcp", port, ipStr, true, "Transhift", 0xFFFFFFFF)
        check(err)
        fmt.Println("done")
        defer clients[0].DeletePortMapping("", port, "tcp")
    } else {
        fmt.Println("UPnP is either not needed or is disabled; continuing...")
    }

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
                panic(fmt.Sprintf("Didn't read %d bytes; %d instead", sizeOfUint64, bytesRead))
            }

            dataSize := bytesToUint64(buffer)

            var finalDataBytesRead bytes.Buffer

            for uint64(finalDataBytesRead.Len()) < dataSize {
                // read dataSize bytes from the connection
                dataBuffer := make([]byte, dataSize - uint64(finalDataBytesRead.Len()))
                dataBytesRead, err := conn.Read(dataBuffer)
                finalDataBytesRead.Write(dataBuffer[:dataBytesRead])
                check(err)
            }

            if uint64(finalDataBytesRead.Len()) != dataSize {
                panic(fmt.Sprintf("Didn't read %d bytes; %d instead", dataSize, finalDataBytesRead.Len()))
            }

            incomingChannel <- finalDataBytesRead.Bytes()
        }
    }()

    return incomingChannel
}

func receive(conn net.Conn, incoming chan []byte, password, fileName string) {
    // wait for password
    incomingPassword := string(<-incoming)

    if (incomingPassword != password) {
        fmt.Println("Peer sent wrong password")
        conn.Write([]byte{passwordBad})
        conn.Close()
        return
    } else {
        fmt.Println("Password verified")
        conn.Write([]byte{passwordGood})
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
            panic(fmt.Sprintf("Didn't write %d bytes; %d instead", len(fileBuffer), fileBytesWritten))
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

package transhift

import (
    "github.com/codegangsta/cli"
    "net"
    "fmt"
//    "os"
//    "bytes"
//    "encoding/hex"
//    "path/filepath"
    "github.com/huin/goupnp/dcps/internetgateway2"
    "math"
)

const (
    chunkSize uint64 = 1024
)

type PortMapping struct {
    port   uint16
    added  bool
    client *internetgateway2.WANIPConnection1
}

func (p *PortMapping) Add() error {
    fmt.Printf("Mapping port %d... ", p.port)

    var ipStr string

    interfaces, err := net.Interfaces()

    if err != nil {
        return err
    }

    for _, i := range interfaces {
        addrs, err := i.Addrs()

        if err != nil {
            return err
        }

        for _, addr := range addrs {
            var ip net.IP

            switch v := addr.(type) {
                case *net.IPNet:
                    ip = v.IP
                case *net.IPAddr:
                    ip = v.IP
            }

            if ! ip.IsLoopback() && ip.To4() != nil {
                ipStr = ip.String()
            }
        }
    }

    clients, _, err := internetgateway2.NewWANIPConnection1Clients()

    if err != nil {
        return err
    }

    if (len(clients) > 0) {
        err = clients[0].AddPortMapping("", p.port, "tcp", p.port, ipStr, true, "Transhift", math.MaxUint32)

        if err != nil {
            return err
        }

        p.client = clients[0]

        fmt.Println("done")
        p.added = true
    } else {
        fmt.Println("UPnP is not available; continuing...")
        p.added = false
    }

    return nil
}

func (p *PortMapping) Remove() {
    if p.added {
        p.client.DeletePortMapping("", p.port, "tcp")
        p.added = false
    }
}

func Download(c *cli.Context) {
//    password := c.Args()[0]
//    destination := c.String("destination")
//
//    portMapping := PortMapping{port: port}
//    portMapping.Add()
//    defer portMapping.Remove()
//
//    uploadPeer := UploadPeer{}
//
//    if ok := dlHandleConnect(&uploadPeer, port); ! ok { return }
//
//    if ok := dlHandlePassword(&uploadPeer, password); ! ok { return }
//
//    fileInfo := dlHandleFileInfo(&uploadPeer)
//    ok, file := dlHandleFileChunks(&uploadPeer, destination, fileInfo)
//
//    if ! ok { return }
//
//    if ok := dlHandleVerification(fileInfo, file); ! ok { return }
}

/*
func dlHandleConnect(uploadPeer *UploadPeer, port uint16) (ok bool) {
    fmt.Print("Listening for peer... ")

    if err := uploadPeer.Connect(port); err != nil {
        return false
    }

    fmt.Println("connected")
    return true
}

func dlHandlePassword(uploadPeer *UploadPeer, password string) (ok bool) {
    fmt.Print("Receiving password... ")

    passwordHash := stringChecksum(password)
    peerPasswordHash := uploadPeer.ReceivePasswordHash()

    if bytes.Equal(passwordHash, peerPasswordHash) {
        fmt.Println("match")
        uploadPeer.SendProtocolResponse(PasswordMatch)
        return true
    } else {
        fmt.Fprintln(os.Stderr, "peer sent wrong password")
        uploadPeer.SendProtocolResponse(PasswordMismatch)
//        uploadPeer.Close()
        return false
    }
}

func dlHandleFileInfo(uploadPeer *UploadPeer) *UploadPeerFileInfo {
    fmt.Print("Receiving file info... ")

    info := uploadPeer.ReceiveFileInfo()

    fmt.Println("done")
    return &info
}

func dlHandleFileChunks(uploadPeer *UploadPeer, destination string, fileInfo *UploadPeerFileInfo) (ok bool, file *os.File) {
    if destination == "" {
        destination = fileInfo.name
    }

    file, err := os.Create(destination)

    if err != nil {
        fmt.Fprintln(os.Stderr, "Error: ", err)
        return false, nil
    }

    absPath, err := filepath.Abs(destination)

    if err != nil {
        fmt.Fprintln(os.Stderr, "Error: ", err)
        return false, nil
    }

    fmt.Printf("Downloading file '%s' with a size of %s (SHA-256 %s) to %s\n",
        fileInfo.name, formatSize(float64(fileInfo.size)),
        hex.EncodeToString(fileInfo.checksum), absPath)

    var totalRead uint64

    ch := uploadPeer.ReceiveFileChunks(chunkSize)
    updateProgressStopSignal := true

    updateProgress(&totalRead, fileInfo.size, &updateProgressStopSignal)

    // while the total amount of bytes we've read is less than the file's
    // size...
    for totalRead < fileInfo.size {
        fileChunk := <- ch

        // add to the total amount of bytes read whatever we just read from the
        // peer
        totalRead += uint64(len(fileChunk.data))

        // the peer wants to disconnect
        if ! fileChunk.good {
            fmt.Fprintln(os.Stderr, "Peer stopped sending file, therefore your " +
                "copy cannot be verified and may be corrupt and/or incomplete. " +
                "You should (probably) delete the incomplete file: ", absPath)
            return false, nil
        }

        file.WriteAt(fileChunk.data, int64(totalRead))
    }

    fmt.Printf("Done! Wrote: %s\n", absPath)
    return true, file
}

func dlHandleVerification(fileInfo *UploadPeerFileInfo, file *os.File) (ok bool) {
    fmt.Print("Verifying checksum... ")

    fileHash, err := fileChecksum(file)

    if err != nil {
        fmt.Print("error: ", err)
        return false
    }

    if bytes.Equal(fileHash, fileInfo.checksum) {
        fmt.Println("match")
        return true
    }

    fmt.Println("mismatch. The file may have been corrupted during transport. Try " +
        "asking for the file to be sent again!")

    return false
}
*/

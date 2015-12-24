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
    "os"
    "bytes"
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
    pass := c.Args()[0]
    destination := c.String("destination")

    portMapping := PortMapping{port: port}
    portMapping.Add()
    defer portMapping.Remove()

    uploadPeer := UploadPeer{}

    fmt.Print("Listening for peer... ")

    if err := uploadPeer.Connect(port); err != nil {
        fmt.Fprintln(os.Stderr, err)
        os.Exit(1)
    }

    fileChunkCh := uploadPeer.ReceiveFileChunks(chunkSize)

    uploadPeer.ReceiveMetaInfo()

    passHash := stringChecksum(pass)

    if bytes.Equal(passHash, uploadPeer.metaInfo.passHash) {
        uploadPeer.SendProtocolResponse(PasswordMatch)
    } else {
        uploadPeer.SendProtocolResponse(PasswordMismatch)
        fmt.Fprintln(os.Stderr, "password mismatch")
        os.Exit(1)
    }

    fmt.Println("done")

    if destination == "" {
        destination = uploadPeer.metaInfo.fileName
    }

    file, err := os.Create(destination)

    if err != nil {
        fmt.Fprintln(os.Stderr, err)
        os.Exit(1)
    }

    var currentBytes uint64
    totalBytes := uint64(uploadPeer.metaInfo.fileSize)
    progressBarStopCh := showProgressBar(&currentBytes, totalBytes)

    for currentBytes < totalBytes {
        fileChunk := <- fileChunkCh

        if ! fileChunk.good {
            progressBarStopCh <- false
            fmt.Fprintln(os.Stderr, "Peer stopped sending file, therefore your " +
                "copy cannot be verified and may be corrupt and/or incomplete. " +
                "You should (probably) delete the incomplete file: ", destination)
            os.Exit(1)
        }

        file.WriteAt(fileChunk.data, int64(currentBytes))
        currentBytes += uint64(len(fileChunk.data))
    }

    progressBarStopCh <- true

    fmt.Println("Downloaded")

    fmt.Print("Verifying file... ")
    fileHash := fileChecksum(file)

    if bytes.Equal(fileHash, uploadPeer.metaInfo.fileHash) {
        uploadPeer.SendProtocolResponse(ChecksumMatch)
    } else {
        uploadPeer.SendProtocolResponse(ChecksumMismatch)
        fmt.Fprintln(os.Stderr, "checksum mismatch; the peer's file may be corrupt and/or incomplete")
        os.Exit(1)
    }

    fmt.Println("done")
}

package source

import (
    "github.com/codegangsta/cli"
    "path/filepath"
    "os"
    "crypto/tls"
    "github.com/transhift/transhift/transhift/storage"
    "github.com/transhift/transhift/transhift/tprotocol"
    "github.com/cheggaaa/pb"
    "io"
    "fmt"
)

type args struct {
    id       string
    filePath string
    appDir   string
}

func Start(c *cli.Context) {
    a := args{
        id:       c.Args()[0],
        filePath: c.Args()[1],
        appDir:   c.GlobalString("app-dir"),
    }

    host, port, cert, err := storage.Prepare(a.appDir)
    if err != nil {
        fmt.Fprintln(os.Stderr, "error:", err)
        return
    }

    if err := run(a, host, port, cert); err != nil {
        fmt.Fprintln(os.Stderr, "error:", err)
        return
    }
}

func run(a args, host string, port int, cert tls.Certificate) error {
    fmt.Print("Getting peer address... ")

    // Punch TCP hole.
    targetAddr, err := punchHole(host, port, cert, a.id)
    if err != nil {
        return err
    }

    fmt.Println("done")
    fmt.Print("Connecting... ")

    // Connect to peer.
    peer := tprotocol.NewPeer(targetAddr)
    if err := peer.Connect(); err != nil {
        return err
    }
    defer peer.Close()

    fmt.Println("done")
    fmt.Print("Sending file info... ")

    file, info, err := getFile(a.filePath)
    if err != nil {
        return err
    }
    defer file.Close()

    // Send file info.
    if err := peer.Enc.Encode(info); err != nil {
        return err
    }

    fmt.Println("done")

    absFilePath, err := filepath.Abs(file.Name())
    if err != nil {
        return err
    }

    fmt.Printf("Uploading %s ...\n", absFilePath)

    bar := pb.New64(info.Size).SetUnits(pb.U_BYTES).Format("[=> ]").Start()
    out := io.MultiWriter(peer.Conn, bar)

    if _, err := io.Copy(out, file); err != nil {
        return err
    }

    bar.FinishPrint("Done!")

    fmt.Print("Awaiting verification... ")

    // Expect verification.
    var verified bool
    if err := peer.Dec.Decode(&verified); err != nil {
        return err
    }

    if verified {
        fmt.Println("done")
    } else {
        fmt.Println("failed: the file may have been corrupted in transport")
    }
    return nil
}

package target

import (
    "github.com/codegangsta/cli"
    "path/filepath"
    "os"
    "crypto/tls"
    "github.com/transhift/transhift/transhift/storage"
    "github.com/transhift/transhift/transhift/tprotocol"
    "github.com/cheggaaa/pb"
    "io"
    "github.com/transhift/transhift/common/protocol"
    "bytes"
    "errors"
    "fmt"
)

type args struct {
    dest   string
    appDir string
}

func Start(c *cli.Context) {
    a := args{
        dest:   c.GlobalString("destination"),
        appDir: c.GlobalString("app-dir"),
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

func run(a args, host string, port int, cert tls.Certificate) (err error) {
    fmt.Print("Getting peer address... ")

    // Punch TCP hole.
    laddr, targetAddr, err := punchHole(host, port, cert)
    if err != nil {
        return
    }

    fmt.Print("Connecting... ")

    // Connect to peer.
    peer := tprotocol.NewPeer(targetAddr)
    if err = peer.Connect(laddr); err != nil {
        return
    }
    defer peer.Close()

    fmt.Println("done")
    fmt.Print("Getting file info... ")

    // Expect file info.
    var info protocol.FileInfo
    if err = peer.Dec.Decode(&info); err != nil {
        return
    }

    fmt.Println("done")

    absFilePath, err := filepath.Abs(getPath(a.dest, info.Name))
    if err != nil {
        return
    }

    file, err := os.Create(absFilePath)
    if err != nil {
        return
    }

    fmt.Printf("Downloading %s to %s ...\n", info.Name, absFilePath)

    bar := pb.New64(info.Size).SetUnits(pb.U_BYTES).Format("[=> ]").Start()
    out := io.MultiWriter(file, bar)

    if _, err = io.CopyN(out, peer.Conn, info.Size); err != nil {
        return
    }

    bar.FinishPrint("Done!")

    fmt.Print("Verifying... ")

    hash, err := storage.HashFile(file)
    if err != nil {
        return
    }
    verified := bytes.Equal(hash, info.Hash)

    if verified {
        fmt.Println("done")
    } else {
        err = errors.New("failed: the file may have been corrupted in transport")
    }

    // Send verification.
    if err := peer.Enc.Encode(verified); err != nil {
        return err
    }

    return
}

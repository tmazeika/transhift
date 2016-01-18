package target

import (
    "github.com/codegangsta/cli"
    "path/filepath"
    "os"
    "crypto/tls"
    "log"
    "github.com/transhift/transhift/transhift/storage"
    "github.com/transhift/transhift/transhift/tprotocol"
    "github.com/cheggaaa/pb"
    "io"
    "github.com/transhift/transhift/common/protocol"
    "bytes"
    "errors"
)

type args struct {
    dest   string
    appDir string
}

func Start(c *cli.Context) {
    log.SetOutput(os.Stdout)
    log.SetFlags(0)

    a := args{
        dest:   c.GlobalString("destination"),
        appDir: c.GlobalString("app-dir"),
    }

    host, port, cert, err := storage.Prepare(a.appDir)
    if err != nil {
        log.SetOutput(os.Stderr)
        log.Fatalln("error:", err)
    }

    if err := run(a, host, port, cert); err != nil {
        log.SetOutput(os.Stderr)
        log.Fatalln("error:", err)
    }
}

func run(a args, host string, port int, cert tls.Certificate) (err error) {
    log.Print("Getting peer address... ")

    // Punch TCP hole.
    targetAddr, err := punchHole(host, port, cert)
    if err != nil {
        return
    }

    log.Println("done")
    log.Print("Connecting... ")

    // Connect to peer.
    peer := tprotocol.NewPeer(targetAddr)
    if err = peer.Connect(); err != nil {
        return
    }
    defer peer.Close()

    log.Println("done")
    log.Print("Getting file info... ")

    // Expect file info.
    var info protocol.FileInfo
    if err = peer.Dec.Decode(&info); err != nil {
        return
    }

    log.Println("done")

    absFilePath, err := filepath.Abs(getPath(a.dest, info.Name))
    if err != nil {
        return
    }

    file, err := os.Create(absFilePath)
    if err != nil {
        return
    }

    log.Printf("Downloading %s to %s ...\n", info.Name, absFilePath)

    bar := pb.New64(info.Size).SetUnits(pb.U_BYTES).Format("[=> ]").Start()
    out := io.MultiWriter(file, bar)

    if _, err = io.CopyN(out, peer.Conn, info.Size); err != nil {
        return
    }

    bar.FinishPrint("Done!")

    log.Print("Verifying... ")

    hash, err := storage.HashFile(file)
    if err != nil {
        return
    }
    verified := bytes.Equal(hash, info.Hash)

    if verified {
        log.Println("done")
    } else {
        err = errors.New("failed: the file may have been corrupted in transport")
    }

    // Send verification.
    if err := peer.Enc.Encode(verified); err != nil {
        return err
    }

    return
}

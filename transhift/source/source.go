package source

import (
    "github.com/codegangsta/cli"
    "path/filepath"
    "os"
    "crypto/tls"
    "log"
    "github.com/transhift/transhift/transhift/tstorage"
    "github.com/transhift/transhift/transhift/tprotocol"
    "github.com/cheggaaa/pb"
    "io"
)

type args struct {
    id       string
    filePath string
    appDir   string
}

func Start(c *cli.Context) {
    log.SetOutput(os.Stdout)
    log.SetFlags(0)

    a := args{
        id:       c.Args()[0],
        filePath: c.Args()[1],
        appDir:   c.GlobalString("app-dir"),
    }

    s, err := tstorage.Prepare(a.appDir)
    if err != nil {
        log.SetOutput(os.Stderr)
        log.Fatalln("Error:", err)
    }

    conf, err := s.Config().(tstorage.Config)
    if err != nil {
        log.SetOutput(os.Stderr)
        log.Fatalln("Error:", err)
    }

    cert, err := s.Certificate(tstorage.KeyFileName, tstorage.CertFileName)
    if err != nil {
        log.SetOutput(os.Stderr)
        log.Fatalln("Error:", err)
    }

    if err := run(a, conf, cert); err != nil {
        log.SetOutput(os.Stderr)
        log.Fatalln("error:", err)
    }
}

func run(a args, conf *tstorage.Config, cert *tls.Certificate) error {
    log.Print("Getting peer address... ")

    // Punch TCP hole.
    targetAddr, err := punchHole(conf.Puncher["host"], conf.Puncher["port"], cert, a.id)
    if err != nil {
        return err
    }

    log.Println("done")
    log.Print("Connecting... ")

    // Connect to peer.
    peer := tprotocol.NewPeer(targetAddr)
    if err := peer.Connect(); err != nil {
        return err
    }
    defer peer.Close()

    log.Println("done")
    log.Print("Sending file info... ")

    file, info, err := getFile(args.filePath)
    if err != nil {
        return err
    }
    defer file.Close()

    // Send file info.
    if err := peer.Enc.Encode(info); err != nil {
        return err
    }

    log.Println("done")

    absFilePath, err := filepath.Abs(file.Name())
    if err != nil {
        return err
    }

    log.Printf("Uploading %s ...\n", absFilePath)

    bar := pb.New64(info.Size).SetUnits(pb.U_BYTES).Format("[=> ]").Start()
    out := io.MultiWriter(peer.Conn, bar)

    if _, err := io.Copy(out, file); err != nil {
        return err
    }

    bar.FinishPrint("Done!")

    log.Print("Awaiting verification... ")

    // Expect verification.
    var verified bool
    if err := peer.Dec.Decode(&verified); err != nil {
        return err
    }

    if verified {
        log.Println("done")
    } else {
        log.Println("failed: the file may have been corrupted in transport")
    }
    return nil
}

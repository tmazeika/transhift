package main

import (
    "github.com/bionicrm/transhift/transhift"
    "github.com/codegangsta/cli"
    "os"
)

func main() {
    app := cli.NewApp()

    app.Name = "Transhift"
    app.Usage = "Peer-to-peer file transfers"
    app.Version = "0.1.0"

    app.Commands = []cli.Command{
        {
            Name: "download",
            Aliases: []string{"dl"},
            Usage: "download from a peer",
            ArgsUsage: "<password>",
            Action: transhift.Download,
            Flags: []cli.Flag{
                cli.StringFlag{
                    Name: "destination, d",
                    Value: "",
                    Usage: "destination directory",
                },
            },
        },
        {
            Name: "upload",
            Aliases: []string{"ul"},
            Usage: "Upload to a peer",
            ArgsUsage: "<peer> <password> <file>",
            Action: transhift.Upload,
        },
    }

    app.Run(os.Args)
}

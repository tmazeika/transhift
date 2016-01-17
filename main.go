package main

import (
    "github.com/codegangsta/cli"
    "github.com/transhift/transhift/transhift/source"
    "github.com/transhift/transhift/transhift/target"
    "os"
)

func main() {
    app := cli.NewApp()
    app.Name = "Transhift"
    app.Usage = "Peer-to-peer file transfers"
    app.Version = "0.1.0"
    app.Flags = []cli.Flag{
        cli.StringFlag{
            Name: "app-dir",
            Value: "",
            Usage: "application directory",
        },
    }
    app.Commands = []cli.Command{
        {
            Name: "download",
            Aliases: []string{"dl"},
            Usage: "Download from a peer",
            Action: target.Start,
            Flags: []cli.Flag{
                cli.StringFlag{
                    Name: "destination, d",
                    Value: "",
                    Usage: "destination file",
                },
            },
        },
        {
            Name: "upload",
            Aliases: []string{"ul"},
            Usage: "Upload to a peer",
            ArgsUsage: "<peer> <file>",
            Action: source.Start,
        },
    }
    app.Run(os.Args)
}

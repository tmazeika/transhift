package transhift

import "github.com/codegangsta/cli"

type DownloadArgs struct {
    password    string
    destination string
}

func (a *DownloadArgs) PasswordHash() []byte {
    return calculateStringHash(a.password)
}

func (a *DownloadArgs) DestinationOrDef(def string) string {
    if a.destination == "" {
        return def
    }
    return a.destination
}

func Download(c *cli.Context) {
    args := DownloadArgs{
        password:    c.Args()[0],
        destination: c.String("destination"),
    }
}

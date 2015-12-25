package transhift

import (
    "github.com/codegangsta/cli"
    "path/filepath"
)

type UploadArgs struct {
    peerHost string
    password string
    filePath string
}

func (a *UploadArgs) PasswordHash() []byte {
    return calculateStringHash(a.password)
}

func (a *UploadArgs) AbsFilePath() string {
    filePath, _ := filepath.Abs(a.filePath)
    return filePath
}

func Upload(c *cli.Context) {
    args := UploadArgs{
        peerHost: c.Args()[0],
        password: c.Args()[1],
        filePath: c.Args()[2],
    }
}

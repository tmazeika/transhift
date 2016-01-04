package transhift

import (
    "encoding/binary"
    "fmt"
    "github.com/transhift/common/common"
)

const (
    // Version is the current version of the application.
    Version = "0.2.0"
)

var (
    // compatibility is a map of versions to an array of compatible, older
    // versions.
    compatibility = map[string][]string{
        "0.1.0": []string{"0.1.0"},
        "0.2.0": []string{"0.2.0"},
    }
)

const (
    // ChunkSize is the number of bytes that are read from the file each
    // iteration of the upload loop.
    ChunkSize = 4096
)

type InOut struct {
    in  *common.In
    out *common.Out
}

func CheckCompatibility(inOut *InOut) error {
    // Send version.
    inOut.out.Ch <- common.Message{ common.Version, []byte(Version) }
    <- inOut.out.Done

    if inOut.out.Err != nil {
        return inOut.out.Err
    }

    // Expect version.
    msg, ok := <- inOut.in.Ch

    if ! ok {
        return inOut.in.Err
    }

    remoteVersion := string(msg.Body)
    localCompatible := func() bool {
        for _, v := range compatibility[Version] {
            if v == remoteVersion {
                return true
            }
        }

        return false
    }()

    // Send local compatibility status.
    if localCompatible {
        inOut.out.Ch <- common.Message{ common.Compatible, nil }
    } else {
        inOut.out.Ch <- common.Message{ common.Incompatible, nil }
    }

    <- inOut.out.Done

    if inOut.out.Err != nil {
        return inOut.out.Err
    }

    // Expect remote compatibility status.
    msg, ok = <- inOut.in.Ch

    if ! ok {
        return inOut.in.Err
    }

    switch msg.Packet {
    case common.Compatible:
    case common.Incompatible:
        if ! localCompatible {
            return fmt.Errorf("incompatible versions %s and %s", Version, remoteVersion)
        }
    default:
        return fmt.Errorf("expected compatibility status, got 0x%x", msg.Packet)
    }

    return nil
}

func uint64Min(x, y uint64) uint64 {
    if x < y {
        return x
    }

    return y
}

func uint64ToBytes(i uint64) (b []byte) {
    b = make([]byte, 8)
    binary.BigEndian.PutUint64(b, i)
    return
}

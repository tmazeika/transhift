package transhift

import (
    "bytes"
    "encoding/binary"
    "fmt"
    "bufio"
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

type ProtocolMessage byte

const (
    DownloadClientType ProtocolMessage = 0x00
    UploadClientType   ProtocolMessage = 0x01
    ChecksumMismatch   ProtocolMessage = 0x02
    ChecksumMatch      ProtocolMessage = 0x03
)

type Serializable interface {
    Serialize() []byte

    Deserialize([]byte)
}

func CheckCompatibility(inOut bufio.ReadWriter) error {
    scanner := bufio.NewScanner(inOut.Reader)

    inOut.WriteString(Version)
    inOut.WriteRune('\n')
    inOut.Flush()
    scanner.Scan()

    remoteVersion := scanner.Text()
    var localCompatible bool

    for _, v := range compatibility[Version] {
        if v == remoteVersion {
            localCompatible = true
            break
        }
    }

    inOut.WriteByte(boolToByte(localCompatible))
    inOut.Flush()
    scanner.Split(bufio.ScanBytes)
    scanner.Scan()

    remoteCompatible := byteToBool(scanner.Bytes()[0])

    if ! localCompatible && ! remoteCompatible {
        return fmt.Errorf("incompatible versions %s and %s", Version, remoteVersion)
    }

    return nil
}

type FileInfo struct {
    name     string
    size     uint64
    checksum []byte
}

func (m FileInfo) Serialize() []byte {
    var buffer bytes.Buffer

    // name
    buffer.WriteString(m.name)
    buffer.WriteRune('\n')
    // size
    buffer.Write(uint64ToBytes(m.size))
    buffer.WriteRune('\n')
    // checksum
    buffer.Write(m.checksum)
    buffer.WriteRune('\n')

    return buffer.Bytes()
}

func (m *FileInfo) Deserialize(b []byte) {
    scanner := bufio.NewScanner(bytes.NewReader(b))

    // name
    scanner.Scan()
    m.name = scanner.Text()
    // size
    scanner.Scan()
    m.size = bytesToUint64(scanner.Bytes())
    // checksum
    scanner.Scan()
    m.checksum = scanner.Bytes()
}

func (m *FileInfo) String() string {
    return fmt.Sprintf("name: '%s', size: %s", m.name, formatSize(m.size))
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

func bytesToUint64(b []byte) uint64 {
    return binary.BigEndian.Uint64(b)
}

func boolToByte(b bool) byte {
    if b {
        return 0x01
    }

    return 0x00
}

func byteToBool(b byte) bool {
    return b != 0x00
}

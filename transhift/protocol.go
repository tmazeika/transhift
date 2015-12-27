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

    // UidLength is the length of the UID that the puncher server issues.
    UidLength = 16
)

type ProtocolMessage byte

const (
    DownloadClientType ProtocolMessage = 0x00
    UploadClientType   ProtocolMessage = 0x01
    ChecksumMismatch   ProtocolMessage = 0x02
    ChecksumMatch      ProtocolMessage = 0x03
)

func CheckCompatibility(inOut *bufio.ReadWriter) error {
    if _, err := inOut.WriteString(Version); err != nil {
        return err
    }

    if _, err := inOut.WriteRune('\n'); err != nil {
        return err
    }

    if err := inOut.Flush(); err != nil {
        return err
    }

    scanner := bufio.NewScanner(inOut.Reader)

    if ! scanner.Scan() {
        return scanner.Err()
    }

    remoteVersion := scanner.Text()
    var localCompatible bool

    for _, v := range compatibility[Version] {
        if v == remoteVersion {
            localCompatible = true
            break
        }
    }

    if err := inOut.WriteByte(boolToByte(localCompatible)); err != nil {
        return err
    }

    if err := inOut.Flush(); err != nil {
        return err
    }

    scanner.Split(bufio.ScanBytes)

    if ! scanner.Scan() {
        return scanner.Err()
    }

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

func (m FileInfo) MarshalBinary() (data []byte, err error) {
    var buffer bytes.Buffer

    // name
    if _, err := buffer.WriteString(m.name); err != nil {
        return nil, err
    }

    if _, err := buffer.WriteRune('\n'); err != nil {
        return nil, err
    }

    // size
    if _, err := buffer.Write(uint64ToBytes(m.size)); err != nil {
        return nil, err
    }

    if _, err := buffer.WriteRune('\n'); err != nil {
        return nil, err
    }

    // checksum
    if _, err := buffer.Write(m.checksum); err != nil {
        return nil, err
    }

    if _, err := buffer.WriteRune('\n'); err != nil {
        return nil, err
    }

    return buffer.Bytes(), nil
}

func (m *FileInfo) UnmarshalBinary(data []byte) error {
    scanner := bufio.NewScanner(bytes.NewReader(data))

    // name
    if ! scanner.Scan() {
        return scanner.Err()
    }

    m.name = scanner.Text()
    // size
    if ! scanner.Scan() {
        return scanner.Err()
    }

    m.size = bytesToUint64(scanner.Bytes())
    // checksum
    if ! scanner.Scan() {
        return scanner.Err()
    }

    m.checksum = scanner.Bytes()

    return nil
}

func (m FileInfo) String() string {
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

func messageToBytes(msg ProtocolMessage) []byte {
    return []byte{byte(msg)}
}

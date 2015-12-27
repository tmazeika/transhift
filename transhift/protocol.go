package transhift

import (
    "bytes"
    "encoding/binary"
    "fmt"
    "bufio"
)

// application information
const (
    AppVersion = "0.2.0"
)

// compatibility information
var (
    appCompatibility = map[string][]string{
        "0.1.0": []string{"0.1.0"},
        "0.2.0": []string{"0.2.0"},
    }
)

// puncher information
const (
    ProtoPeerUIDLen    = 16
)

// protocol information
const (
    ProtoChunkSize = 4096
)

type ProtoMsg byte

// protocol messages
const (
    ProtoMsgClientTypeDL     ProtoMsg = 0x00
    ProtoMsgClientTypeUL     ProtoMsg = 0x01
    ProtoMsgChecksumMismatch ProtoMsg = 0x02
    ProtoMsgChecksumMatch    ProtoMsg = 0x03
)

func checkCompatibility(in *bufio.Reader, out *bufio.Writer) error {
    compare := func(v1, v2 string) bool {
        if appCompatibility[v1] != nil {
            for _, v := range appCompatibility[v1] {
                if v == v2 {
                    return true
                }
            }
        }

        return false
    }

    out.WriteString(AppVersion)
    out.WriteRune('\n')
    out.Flush()

    line, err := in.ReadBytes('\n')
    line = line[:len(line) - 1] // trim trailing \n

    if err != nil {
        return err
    }

    remoteVersion := string(line)
    localCompatibility := compare(AppVersion, remoteVersion)
    out.WriteByte(boolToByte(localCompatibility))
    out.Flush()

    lineBuffer := make([]byte, 1)
    _, err = in.Read(lineBuffer)

    if err != nil {
        return err
    }

    if ! localCompatibility && ! byteToBool(lineBuffer[0]) {
        return fmt.Errorf("incompatible versions %s and %s", AppVersion, remoteVersion)
    }

    return nil
}

type Serializable interface {
    Serialize() []byte

    Deserialize([]byte)
}

type ProtoMetaInfo struct {
    fileName      string
    fileSize      uint64
    fileChecksum  []byte
}

func (m *ProtoMetaInfo) Serialize() []byte {
    var buffer bytes.Buffer

    // fileName
    buffer.WriteString(m.fileName)
    buffer.WriteRune('\n')
    // fileSize
    buffer.Write(uint64ToBytes(m.fileSize))
    buffer.WriteRune('\n')
    // fileChecksum
    buffer.Write(m.fileChecksum)
    buffer.WriteRune('\n')

    return buffer.Bytes()
}

func (m *ProtoMetaInfo) Deserialize(b []byte) {
    scanner := bufio.NewScanner(bytes.NewReader(b))

    // fileName
    scanner.Scan()
    m.fileName = scanner.Text()
    // fileSize
    scanner.Scan()
    m.fileSize = bytesToUint64(scanner.Bytes())
    // fileChecksum
    scanner.Scan()
    m.fileChecksum = scanner.Bytes()
}

func (m *ProtoMetaInfo) String() string {
    return fmt.Sprintf("name: '%s', size: %s", m.fileName, formatSize(m.fileSize))
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

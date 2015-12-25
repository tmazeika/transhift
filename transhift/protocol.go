package transhift

import (
    "bytes"
    "encoding/binary"
    "fmt"
)

// protocol information
const (
    ProtoPort uint16 =  50977
    ProtoPortStr     = "50977"
    ProtoChunkSize   = 1024
)

// protocol messages
const (
    ProtoMsgPasswordMismatch = byte(0x00)
    ProtoMsgPasswordMatch    = byte(0x01)
    ProtoMsgChecksumMismatch = byte(0x02)
    ProtoMsgChecksumMatch    = byte(0x03)
)

type Serializable interface {
    Serialize() []byte

    Deserialize([]byte)
}

type ProtoMetaInfo struct {
    passwordChecksum []byte
    fileName         string
    fileSize         uint64
    fileChecksum     []byte
}

func (m *ProtoMetaInfo) Serialize() []byte {
    var buffer bytes.Buffer

    // passwordChecksum
    buffer.Write(m.passwordChecksum)
    buffer.WriteRune('\n')
    // fileName
    buffer.WriteString(m.fileName)
    buffer.WriteRune('\n')
    // fileSize
    fileSizeBuffer := make([]byte, 8)
    binary.BigEndian.PutUint64(fileSizeBuffer, m.fileSize)
    buffer.Write(fileSizeBuffer)
    buffer.WriteRune('\n')
    // fileChecksum
    buffer.Write(m.fileChecksum)
    buffer.WriteRune('\n')

    return buffer.Bytes()
}

func (m *ProtoMetaInfo) Deserialize(b []byte) {
    buffer := bytes.NewBuffer(b)

    // passwordChecksum
    m.passwordChecksum, _ = buffer.ReadBytes('\n')
    m.passwordChecksum = m.passwordChecksum[:len(m.passwordChecksum) - 1] // trim leading \n
    // fileName
    m.fileName, _ = buffer.ReadString('\n')
    m.fileName = m.fileName[:len(m.fileName) - 1] // trim leading \n
    // fileSize
    fileSize, _ := buffer.ReadBytes('\n')
    fileSize = fileSize[:len(fileSize) - 1] // trim leading \n
    m.fileSize = binary.BigEndian.Uint64(fileSize)
    // fileChecksum
    m.fileChecksum, _ = buffer.ReadBytes('\n')
    m.fileChecksum = m.fileChecksum[:len(m.fileChecksum) - 1] // trim leading \n
}

func (m *ProtoMetaInfo) String() string {
    return fmt.Sprintf("{passwordChecksum=%x, fileName=%s, fileSize=%d, fileChecksum=%x}",
        m.passwordChecksum, m.fileName, m.fileSize, m.fileChecksum)
}

type ProtoChunk struct {
    close bool
    data  []byte
    last  bool
}

func (c *ProtoChunk) Serialize() []byte {
    var buffer bytes.Buffer

    // close
    buffer.WriteByte(boolToByte(c.close))
    // data
    buffer.Write(c.data)

    return buffer.Bytes()
}

func (c *ProtoChunk) Deserialize(b []byte) {
    // close
    c.close = byteToBool(b[0])
    // data
    c.data = b[1:]
}

func (c *ProtoChunk) String() string {
    return fmt.Sprintf("{close=%t, last=%t, data=(len)%d}",
        c.close, c.last, len(c.data))
}

func boolToByte(b bool) byte {
    if b {
        return 0x01
    }
    return 0x00
}

func byteToBool(b byte) bool {
    if b == 0x00 {
        return false
    }
    return true
}

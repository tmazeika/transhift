package transhift

import (
    "bytes"
    "encoding/binary"
    "fmt"
)

const (
    ProtoPort uint16 =  50977
    ProtoPortStr     = "50977"
    ProtoChunkSize   = 1024
)

type Serializable interface {
    Serialize() []byte

    Deserialize([]byte)
}

type ProtoMetaInfo struct {
    passwordHash []byte
    fileName     string
    fileSize     uint64
    fileHash     []byte
}

func (m *ProtoMetaInfo) Serialize() []byte {
    var buffer bytes.Buffer

    // passwordHash
    buffer.Write(m.passwordHash)
    buffer.WriteRune('\n')
    // fileName
    buffer.WriteString(m.fileName)
    buffer.WriteRune('\n')
    // fileSize
    fileSizeBuffer := make([]byte, 8)
    binary.BigEndian.PutUint64(fileSizeBuffer, m.fileSize)
    buffer.Write(fileSizeBuffer)
    buffer.WriteRune('\n')
    // fileHash
    buffer.Write(m.fileHash)
    buffer.WriteRune('\n')

    return buffer.Bytes()
}

func (m *ProtoMetaInfo) Deserialize(b []byte) {
    buffer := bytes.NewBuffer(b)

    // passwordHash
    m.passwordHash, _ = buffer.ReadBytes('\n')
    m.passwordHash = m.passwordHash[:len(m.passwordHash) - 1] // trim leading \n
    // fileName
    m.fileName, _ = buffer.ReadString('\n')
    m.fileName = m.fileName[:len(m.fileName) - 1] // trim leading \n
    // fileSize
    fileSize, _ := buffer.ReadBytes('\n')
    fileSize = fileSize[:len(fileSize) - 1] // trim leading \n
    m.fileSize = binary.BigEndian.Uint64(fileSize)
    // fileHash
    m.fileHash, _ = buffer.ReadBytes('\n')
    m.fileHash = m.fileHash[:len(m.fileHash) - 1] // trim leading \n
}

func (m *ProtoMetaInfo) String() string {
    return fmt.Sprintf("{passwordHash=%x, fileName=%s, fileSize=%d, fileHash=%x}",
        m.passwordHash, m.fileName, m.fileSize, m.fileHash)
}

type ProtoChunk struct {
    close bool
    data  []byte
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
    return fmt.Sprintf("{close=%t, data=(len)%d}", c.close, len(c.data))
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

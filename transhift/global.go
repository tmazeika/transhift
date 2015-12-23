package transhift

import (
    "fmt"
    "crypto/sha256"
    "encoding/hex"
    "math"
    "time"
    "os"
    "io"
)

const (
    port          uint16 = 50977
    chunk_size    uint16 = 1024

    password_good byte = 0
    password_bad  byte = 1
)

var (
    portStr = fmt.Sprint(port)
)

func check(e error) {
    if e != nil {
        panic(e)
    }
}

func checksum(filePath string) []byte {
    fmt.Print("Calculating sha256 checksum... ")
    file, err := os.Open(filePath)
    check(err)
    defer file.Close()

    hasher := sha256.New()
    _, err = io.Copy(hasher, file)
    check(err)
    sum := hasher.Sum(nil)
    fmt.Println(hex.EncodeToString(sum))
    return sum;
}

func formatSize(size uint64) string {
    sizeF := float64(size)

    if size < 1000 {
        return fmt.Sprintf("%d B", size)
    } else if sizeF < math.Pow(1000, 2) {
        return fmt.Sprintf("%.2f KB", sizeF / 1000.0)
    } else if sizeF < math.Pow(1000, 3) {
        return fmt.Sprintf("%.2f MB", sizeF / math.Pow(1000, 2))
    } else if sizeF < math.Pow(1000, 4) {
        return fmt.Sprintf("%.2f GB", sizeF / math.Pow(1000, 3))
    } else {
        return fmt.Sprintf("%.2f TB", sizeF / math.Pow(1000, 4))
    }
}

func updateProgress(now* uint64, total* uint64, printCondition* bool) {
    go func() {
        for *now < *total {
            percent := float64(*now) / float64(*total) * 100

            if (*printCondition) {
                fmt.Printf("%.2f%% (%s/%s)\n", percent, formatSize(*now), formatSize(*total))
            }

            time.Sleep(time.Second)
        }
    }()
}

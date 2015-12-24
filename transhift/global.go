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
    port         uint16 = 50977
    chunkSize    uint16 = 1024
)

func formatSize(size float64) string {
    if size < 1000 {
        return fmt.Sprintf("%d B", size)
    } else if size < math.Pow(1000, 2) {
        return fmt.Sprintf("%.2f KB", size / 1000.0)
    } else if size < math.Pow(1000, 3) {
        return fmt.Sprintf("%.2f MB", size / math.Pow(1000, 2))
    } else if size < math.Pow(1000, 4) {
        return fmt.Sprintf("%.2f GB", size / math.Pow(1000, 3))
    } else {
        return fmt.Sprintf("%.2f TB", size / math.Pow(1000, 4))
    }
}

func updateProgress(now* float64, total float64) {
    go func() {
        for *now < total {
            percent := *now / total * 100

            fmt.Printf("%.2f%% (%s/%s)\n", percent, formatSize(*now), formatSize(total))
            time.Sleep(time.Second)
        }
    }()
}

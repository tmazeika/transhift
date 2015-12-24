package transhift

import (
    "fmt"
    "math"
    "time"
)

const (
    port uint16 = 50977
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

func updateProgress(now *uint64, total uint64, stopSignal *bool) {
    go func() {
        for ! *stopSignal && *now < total {
            percent := float64(*now) / float64(total) * 100.0

            fmt.Printf("%.2f%% (%s/%s)\n", percent, formatSize(float64(*now)), formatSize(float64(total)))
            time.Sleep(time.Second)
        }
    }()
}

package transhift

import (
    "fmt"
    "math"
)

const (
    port uint16 = 50977
)

func pow(x, n uint64) uint64 {
    return uint64(math.Pow(float64(x), float64(n)))
}

func formatSize(size uint64) string {
    if size < 1000 {
        return fmt.Sprintf("%d B", size)
    } else if size < pow(1000, 2) {
        return fmt.Sprintf("%.2f KB", size / 1000.0)
    } else if size < pow(1000, 3) {
        return fmt.Sprintf("%.2f MB", size / pow(1000, 2))
    } else if size < pow(1000, 4) {
        return fmt.Sprintf("%.2f GB", size / pow(1000, 3))
    } else {
        return fmt.Sprintf("%.2f TB", size / pow(1000, 4))
    }
}

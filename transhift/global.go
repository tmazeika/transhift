package transhift

import (
    "fmt"
    "math"
    "bytes"
    "time"
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

func showProgressBar(current *uint64, total uint64) (stopCh chan bool) {
    stopCh = make(chan bool)
    stop := false

    update := func() {
        var buff bytes.Buffer
        percent := float64(*current) / float64(total)

        buff.WriteString(fmt.Sprintf("\r%.f%% [", percent))

        const BarSize = float64(50)

        for i := float64(0); i < percent * BarSize; i++ {
            buff.WriteRune('=')
        }

        for i := float64(0); i < BarSize - percent * BarSize; i++ {
            buff.WriteRune(' ')
        }

        buff.WriteString(fmt.Sprintf("] %s/%s", formatSize(*current), formatSize(total)))
        fmt.Print(buff.String())
    }

    go func() {
        for ! stop && *current < total {
            update()
            time.Sleep(time.Second)
        }
    }()

    go func() {
        updateAfterStop := <- stopCh
        stop = true

        if updateAfterStop {
            update()
        }
    }()

    return
}

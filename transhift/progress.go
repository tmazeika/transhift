package transhift

import (
    "bytes"
    "fmt"
    "time"
    "math"
)

func showProgressBar(current *uint64, total uint64) (stopCh chan bool) {
    stopCh = make(chan bool)
    stop := false

    update := func() {
        var buff bytes.Buffer
        percent := float64(*current) / float64(total)

        buff.WriteString(fmt.Sprintf("\r%.f%% [", percent * 100))

        const BarSize = float64(50)

        for i := float64(0); i < percent * BarSize - 1; i++ {
            buff.WriteRune('=')
        }

        buff.WriteRune('>')

        for i := float64(0); i < BarSize - percent * BarSize; i++ {
            buff.WriteRune(' ')
        }

        buff.WriteString(fmt.Sprintf("] %s / %s", formatSize(*current), formatSize(total)))
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

func formatSize(size uint64) string {
    fSize := float64(size)

    switch {
    case fSize < 1000:
        return fmt.Sprintf("%d B", size)
    case fSize < math.Pow(1000, 2):
        return fmt.Sprintf("%.2f KB", fSize / 1000)
    case fSize < math.Pow(1000, 3):
        return fmt.Sprintf("%.2f MB", fSize / math.Pow(1000, 2))
    case fSize < math.Pow(1000, 4):
        return fmt.Sprintf("%.2f GB", fSize / math.Pow(1000, 3))
    default:
        return fmt.Sprintf("%.2f TB", fSize / math.Pow(1000, 4))
    }
}

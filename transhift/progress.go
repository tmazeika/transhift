package transhift

import (
    "bytes"
    "fmt"
    "time"
    "math"
)

type ProgressBar struct {
    stop    bool
    current *uint64
    total   uint64
}

func (p *ProgressBar) Start() {
    go func() {
        for ! p.stop {
            p.Update()
            time.Sleep(time.Second)
        }
    }()
}

func (p *ProgressBar) Update() {
    var buff bytes.Buffer
    percent := float64(*p.current) / float64(p.total)

    buff.WriteString(fmt.Sprintf("%.f%% [", percent * 100))

    const BarSize = float64(50)

    for i := float64(0); i < percent * BarSize - 1; i++ {
        buff.WriteRune('=')
    }

    buff.WriteRune('>')

    for i := float64(0); i < BarSize - percent * BarSize; i++ {
        buff.WriteRune(' ')
    }

    buff.WriteString(fmt.Sprintf("] %s / %s", formatSize(*p.current), formatSize(p.total)))
    fmt.Println(buff.String())
}

func (p *ProgressBar) Stop(forceUpdate bool) {
    p.stop = true

    if forceUpdate {
        p.Update()
    }
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

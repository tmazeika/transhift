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
    const BarSize = float64(50)

    var buff bytes.Buffer
    percent := float64(*p.current) / float64(p.total)
    buff.WriteString(fmt.Sprintf("%3.f%% [", percent * 100))
    var progressItrs uint8

    for i := uint8(0); i < uint8(percent * BarSize); i++ {
        buff.WriteRune('=')
        progressItrs++
    }

    buff.WriteRune('>')

    for i := uint8(0); i < uint8(BarSize) - progressItrs; i++ {
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
        return fmt.Sprintf("%6d B ", size)
    case fSize < math.Pow(1000, 2):
        return fmt.Sprintf("%6.2f KB", fSize / 1000)
    case fSize < math.Pow(1000, 3):
        return fmt.Sprintf("%6.2f MB", fSize / math.Pow(1000, 2))
    case fSize < math.Pow(1000, 4):
        return fmt.Sprintf("%6.2f GB", fSize / math.Pow(1000, 3))
    default:
        return fmt.Sprintf("%6.2f TB", fSize / math.Pow(1000, 4))
    }
}

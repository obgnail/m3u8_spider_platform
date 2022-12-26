package downloader

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	KB = 1024
	MB = 1024 * 1024
	GB = 1024 * 1024 * 1024
)

type ProcessBar struct {
	mu        sync.Mutex // protect current and load()
	graph     string     // 显示符号
	start     int        // 开始的进度位置
	current   int        // 当前的进度位置
	total     int        // 总进度
	bytes     uint64     // 总字节数
	startTime time.Time  // 开始时间
	_once     sync.Once
}

func NewBarWithGraph(start, total int, graph string) *ProcessBar {
	bar := NewBar(start, total)
	bar.graph = graph
	return bar
}

func NewBar(current, total int) *ProcessBar {
	bar := &ProcessBar{graph: "█", total: total}
	bar.reset(current)
	return bar
}

func (b *ProcessBar) Add(bytes uint64) {
	b.withLock(func() {
		b.current += 1
		b.bytes += bytes
		if b.current > b.total {
			b.current = b.total
		}
		b.load()
	})
}

func (b *ProcessBar) Reset(current int) {
	b.withLock(func() {
		b.reset(current)
		b.load()
	})
}

func (b *ProcessBar) Start() {
	b._once.Do(func() {
		go func() {
			for b.current < b.total {
				b.withLock(b.load)
				time.Sleep(time.Second)
			}
		}()
	})
}

func (b *ProcessBar) withLock(f func()) {
	b.mu.Lock()
	f()
	b.mu.Unlock()
}

func (b *ProcessBar) reset(current int) {
	if current > b.total {
		current = b.total
	}
	b.start = current
	b.current = current
	b.bytes = 0
	b.startTime = time.Now()
}

func (b *ProcessBar) getPercent() int {
	return int((float64(b.current) / float64(b.total)) * 100)
}

func toTimeString(second float64) string {
	str := ""
	h := int(second) / 3600
	m := int(second) % 3600 / 60
	if h > 0 {
		str += strconv.Itoa(h) + "h "
	}
	if h > 0 || m > 0 {
		str += strconv.Itoa(m) + "m "
	}
	str += strconv.Itoa(int(second)%60) + "s"
	return str
}

func (b *ProcessBar) toSizeString(fileSize uint64) string {
	switch {
	case fileSize < KB:
		return fmt.Sprintf("%.2fB", float64(fileSize))
	case fileSize < MB:
		return fmt.Sprintf("%.2fKB", float64(fileSize)/KB)
	case fileSize < GB:
		return fmt.Sprintf("%.2fMB", float64(fileSize)/MB)
	default:
		return fmt.Sprintf("%.2fGB", float64(fileSize)/GB)
	}
}

func (b *ProcessBar) getTransferRate() string {
	spent := uint64(time.Now().Sub(b.startTime).Seconds())
	if spent == 0 {
		return "0B/s"
	}
	size := b.toSizeString(b.bytes / spent)
	return fmt.Sprintf("%s/s", size)
}

func (b *ProcessBar) getTotalTransferSize() string {
	return b.toSizeString(b.bytes)
}

func (b *ProcessBar) getSpentTime() string {
	u := time.Now().Sub(b.startTime).Seconds()
	return toTimeString(u)
}

func (b *ProcessBar) getRemainTime() string {
	process := b.current - b.start
	if process == 0 {
		return "INF"
	}
	spent := time.Now().Sub(b.startTime).Seconds()
	remain := b.total - b.current
	u := spent * float64(remain) / float64(process)
	return toTimeString(u)
}

// need lock
func (b *ProcessBar) load() {
	percent := b.getPercent()
	spent := b.getSpentTime()
	remain := b.getRemainTime()
	transferRate := b.getTransferRate()
	totalTransfer := b.getTotalTransferSize()
	bar := strings.Repeat(b.graph, percent/2)

	fmt.Printf("\r[%-50s]% 3d%%(%d/%d)    %2s(%2s)    %2s(+%2s)        ",
		bar, percent, b.current, b.total, totalTransfer, transferRate, spent, remain)

	if b.current == b.total {
		fmt.Println()
	}
}

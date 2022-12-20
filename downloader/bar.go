package downloader

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"
)

type ProcessBar struct {
	mu        sync.Mutex // protect current and load()
	graph     string     // 显示符号
	start     int        // 开始的进度位置
	current   int        // 当前的进度位置
	total     int        // 总进度
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

func (bar *ProcessBar) Add(i int) {
	bar.withLock(func() {
		bar.current += i
		if bar.current > bar.total {
			bar.current = bar.total
		}
		bar.load()
	})
}

func (bar *ProcessBar) Reset(current int) {
	bar.withLock(func() {
		bar.reset(current)
		bar.load()
	})
}

func (bar *ProcessBar) Start() {
	bar._once.Do(func() {
		go func() {
			for bar.current < bar.total {
				bar.withLock(bar.load)
				time.Sleep(time.Second)
			}
		}()
	})
}

func (bar *ProcessBar) withLock(f func()) {
	bar.mu.Lock()
	f()
	bar.mu.Unlock()
}

func (bar *ProcessBar) reset(current int) {
	if current > bar.total {
		current = bar.total
	}
	bar.start = current
	bar.current = current
	bar.startTime = time.Now()
}

func (bar *ProcessBar) getPercent() int {
	return int((float64(bar.current) / float64(bar.total)) * 100)
}

func toString(second float64) string {
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

func (bar *ProcessBar) getSpentTime() string {
	u := time.Now().Sub(bar.startTime).Seconds()
	return toString(u)
}

func (bar *ProcessBar) getRemainTime() string {
	process := bar.current - bar.start
	if process == 0 {
		return "INF"
	}
	spent := time.Now().Sub(bar.startTime).Seconds()
	remain := bar.total - bar.current
	u := spent * float64(remain) / float64(process)
	return toString(u)
}

// need lock
func (bar *ProcessBar) load() {
	percent := bar.getPercent()
	spent := bar.getSpentTime()
	remain := bar.getRemainTime()
	rate := strings.Repeat(bar.graph, percent/2)
	fmt.Printf("\r[%-50s]% 3d%%(%d/%d)    %2s(+%2s)        ",
		rate, percent, bar.current, bar.total, spent, remain)

	if bar.current == bar.total {
		fmt.Println()
	}
}

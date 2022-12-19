package main

import (
	"fmt"
	"strconv"
	"sync"
	"time"
)

type Bar struct {
	mu      sync.Mutex
	graph   string    // 显示符号
	rate    string    // 进度条
	percent int       // 百分比
	current int       // 当前进度位置
	total   int       // 总进度
	start   time.Time // 开始时间

	once sync.Once
}

func NewBar(current, total int) *Bar {
	bar := new(Bar)
	bar.current = current
	bar.total = total
	bar.start = time.Now()
	if bar.graph == "" {
		bar.graph = "█"
	}
	bar.percent = bar.getPercent()
	for i := 0; i < bar.percent; i += 2 {
		bar.rate += bar.graph //初始化进度条位置
	}
	return bar
}

func (bar *Bar) getPercent() int {
	return int((float64(bar.current) / float64(bar.total)) * 100)
}

func calcTime(second float64) (s string) {
	h := int(second) / 3600
	m := int(second) % 3600 / 60
	if h > 0 {
		s += strconv.Itoa(h) + "h "
	}
	if h > 0 || m > 0 {
		s += strconv.Itoa(m) + "m "
	}
	s += strconv.Itoa(int(second)%60) + "s"
	return
}

func (bar *Bar) getSpentTime() (s string) {
	u := time.Now().Sub(bar.start).Seconds()
	return calcTime(u)
}

func (bar *Bar) getRemainTime() (s string) {
	if bar.current == 0 {
		return "INF"
	}
	spent := time.Now().Sub(bar.start).Seconds()
	remain := bar.total - bar.current
	u := spent * float64(remain) / float64(bar.current)
	return calcTime(u)
}

func NewBarWithGraph(start, total int, graph string) *Bar {
	bar := NewBar(start, total)
	bar.graph = graph
	return bar
}

func (bar *Bar) Start() {
	bar.once.Do(func() {
		go func() {
			for bar.current != bar.total {
				bar.load()
				time.Sleep(time.Second)
			}
		}()
	})
}

func (bar *Bar) load() {
	last := bar.percent
	bar.percent = bar.getPercent()
	if bar.percent != last && bar.percent%2 == 0 {
		bar.rate += bar.graph
	}
	fmt.Printf("\r[%-50s]% 3d%%    %2s   %d/%d    +%2s        \r",
		bar.rate, bar.percent, bar.getSpentTime(), bar.current, bar.total, bar.getRemainTime())

	if bar.current == bar.total {
		fmt.Println()
	}
}

func (bar *Bar) Reset(current int) {
	bar.mu.Lock()
	defer bar.mu.Unlock()
	bar.current = current
	bar.load()
}

func (bar *Bar) Add(i int) {
	bar.mu.Lock()
	defer bar.mu.Unlock()
	bar.current += i
	bar.load()
}

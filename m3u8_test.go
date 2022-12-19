package main

import (
	"fmt"
	"github.com/juju/errors"
	"testing"
	"time"
)

func TestM3u8_1(t *testing.T) {
	// nunuyy5.org
	url := "https://b.baobuzz.com/m3u8/569128.m3u8?sign=4d5618aebd4dd9a59b0533e0603922d9"
	downloader := Default(url, "")
	errHandler(downloader.Run())
}

func TestM3u8_2(t *testing.T) {
	d := map[int]string{
		78: "https://b.baobuzz.com/m3u8/569204.m3u8?sign=e99a2a49dc13a136fc2dc2a671dc4a28",
	}

	for idx := 20; idx < 100; idx++ {
		m3u8, ok := d[idx]
		if ok {
			ep := fmt.Sprintf("%02d.ts", idx)
			down := "d:\\tmp\\tmp\\download\\"
			downloader := New(m3u8, ep, down, "e:\\三国演义", true,
				10, 5, nil)
			if err := downloader.Run(); err != nil {
				Logger.Error(errors.ErrorStack(err))
			}
		}
	}
}

func TestBar(t *testing.T) {
	b := NewBar(0, 1000)
	b.Start()
	for i := 0; i < 1000; i++ {
		b.Add(1)
		time.Sleep(time.Second)
	}
}

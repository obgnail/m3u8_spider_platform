package downloader

import (
	"fmt"
	"github.com/juju/errors"
	"net/url"
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
				10, 5, nil, nil, nil)
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

func TestUrlParse(t *testing.T) {
	Url := "https://m3u.haiwaikan.com/xm3u8/6230a02ccf71435937f955e2a0fffb55f65ee324081a33f6852a71657733da119921f11e97d0da21.m3u8"
	r, _ := url.Parse(Url)
	s := fmt.Sprintf("%s://%s", r.Scheme, r.Host)
	t.Log(s)
}

package main

import (
	"github.com/obgnail/m3u8_spider_platform/downloader"
)

func main() {
	url := "https://bf3.sbdm.cc/runtime/Aliyun/9208ddf4d3ad882f80a9fd59860798fc.m3u8"
	downloader.Default(url, "bocchi the rock 11.ts").Crawl()
}

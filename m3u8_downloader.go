package main

import (
	"fmt"
	"github.com/juju/errors"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	Url "net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const (
	UA = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/108.0.0.0 Safari/537.36"
)

var Logger *logrus.Logger

func init() {
	Logger = &logrus.Logger{
		Out:   os.Stderr,
		Level: logrus.DebugLevel,
		Formatter: &logrus.TextFormatter{
			ForceColors:               true,
			EnvironmentOverrideColors: true,
			DisableQuote:              true,
			DisableLevelTruncation:    true,
			FullTimestamp:             true,
			TimestampFormat:           "15:04:05",
		},
	}
}

type WaitGroup struct {
	wg sync.WaitGroup
	p  chan struct{}
}

func NewWaitGroup(parallel int) (w *WaitGroup) {
	w = &WaitGroup{}
	if parallel <= 0 {
		return
	}
	w.p = make(chan struct{}, parallel)
	return
}

func (w *WaitGroup) AddDelta() {
	w.wg.Add(1)
	if w.p == nil {
		return
	}
	w.p <- struct{}{}
}

func (w *WaitGroup) Done() {
	w.wg.Done()
	if w.p == nil {
		return
	}
	<-w.p
}

func (w *WaitGroup) Wait() {
	w.wg.Wait()
}

func (w *WaitGroup) Parallel() int {
	return len(w.p)
}

type M3u8Downloader struct {
	url         string
	saveName    string
	downPath    string
	savePath    string
	clearDebris bool
	threads     int
	maxtry      int
	isShardFunc func(idx int, line string) (skip bool)

	encrypt     bool
	mutex       sync.Mutex // protect checkBitMap
	checkBitMap []byte
}

func New(url, saveName, downPath, savePath string, clearDebris bool, threads, maxtry int,
	isShardFunc func(idx int, line string) (skip bool)) *M3u8Downloader {

	if len(saveName) == 0 {
		u, err := Url.Parse(url)
		if err != nil {
			panic(err)
		}
		result := strings.Split(u.Path, "/")
		saveName = fmt.Sprintf("%s.ts", result[len(result)-1])
	}
	if len(downPath) == 0 {
		downPath = fmt.Sprintf("./Download_%d", time.Now().Unix())
	}
	if len(savePath) == 0 {
		savePath = "./Complete"
	}

	d := &M3u8Downloader{
		url:         url,
		saveName:    saveName,
		downPath:    downPath,
		savePath:    savePath,
		clearDebris: clearDebris,
		threads:     threads,
		maxtry:      maxtry,
		checkBitMap: nil,
		encrypt:     false,
	}

	d.SetIsShardFunc(isShardFunc)
	return d
}

func Default(url, saveName string) *M3u8Downloader {
	return New(url, saveName, "", "", true, 24, 5, nil)
}

func (d *M3u8Downloader) SetIsShardFunc(isShardFunc func(idx int, line string) (skip bool)) {
	f := func(idx int, line string) bool {
		ok := true
		if d.checkBitMap != nil {
			ok = d.checkBitMap[idx] != 0
		}

		if isShardFunc == nil {
			isShardFunc = func(idx int, line string) (skip bool) {
				// return strings.Index(line, ".ts") == -1
				return strings.HasPrefix(line, "http")
			}
		}
		return ok && isShardFunc(idx, line)
	}
	d.isShardFunc = f
}

func (d *M3u8Downloader) initCheckBitMap(shards []string) {
	d.mutex.Lock()
	d.checkBitMap = make([]byte, len(shards))
	d.mutex.Unlock()
}

func (d *M3u8Downloader) shardDone(idx int) {
	d.mutex.Lock()
	d.checkBitMap[idx] = 1
	d.mutex.Unlock()
}

func (d *M3u8Downloader) done() bool {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	for _, ele := range d.checkBitMap {
		if ele == 0 {
			return false
		}
	}
	return true
}

func (d *M3u8Downloader) checkUrl(url string) bool {
	return strings.HasPrefix(url, "http")
}

func (d *M3u8Downloader) request(url string) (body []byte, err error) {
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("user-agent", UA)
	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		return nil, errors.Trace(err)
	}
	defer resp.Body.Close()

	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Trace(err)
	}
	if resp.StatusCode != 200 {
		return body, fmt.Errorf("resp.statusCode == %d", resp.StatusCode)
	}
	return
}

func (d *M3u8Downloader) parse(url string,
	isShardFunc func(idx int, line string) (skip bool)) (shards []string, err error) {
	if isShardFunc == nil {
		panic("isShardFunc == nil")
	}
	resp, err := d.request(url)
	if err != nil {
		return nil, errors.Trace(err)
	}
	response := strings.Split(string(resp), "\n")

	var shardIdx = 0
	for _, line := range response {
		if isShardFunc(shardIdx, line) == true {
			continue
		}

		shards = append(shards, line)
		if strings.HasPrefix(line, "#EXT-X-KEY:") {
			d.encrypt = true
		}

		shardIdx++
	}
	if len(shards) == 0 {
		return nil, fmt.Errorf("len(shards) == 0")
	}
	return
}

func Mkdir(Path string) error {
	if _, err := os.Stat(Path); os.IsNotExist(err) {
		if err = os.MkdirAll(Path, os.ModePerm); err != nil {
			return errors.Trace(err)
		}
	}
	return nil
}

func (d *M3u8Downloader) mkdir() (err error) {
	if err = Mkdir(d.downPath); err != nil {
		return errors.Trace(err)
	}
	if err = Mkdir(d.savePath); err != nil {
		return errors.Trace(err)
	}
	return
}

func writeFile(Path string, content []byte) (err error) {
	file, err := os.OpenFile(Path, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return errors.Trace(err)
	}
	defer file.Close()
	if _, err = file.Write(content); err != nil {
		return errors.Trace(err)
	}
	return
}

func (d *M3u8Downloader) download(group *WaitGroup, url string, idx int, shard, downPath string) (err error) {
	defer group.Done()

	res := strings.Split(url, "/")
	baseUrl := strings.Join(res[:len(res)-1], "/")

	if !strings.HasPrefix(shard, "http") {
		shard = baseUrl + "/" + shard
	}

	debrisName := path.Join(downPath, fmt.Sprintf("%05d.ts", idx))
	if _, err = os.Stat(debrisName); os.IsNotExist(err) {
		resp, err := d.request(shard)
		if err != nil {
			return errors.Trace(err)
		}
		if err = writeFile(debrisName, resp); err != nil {
			return errors.Trace(err)
		}
		d.shardDone(idx)
		Logger.Debugf("finished shard. %d: %s", idx, shard)
	}
	return nil
}

func (d *M3u8Downloader) downloadShards(url string, shards []string, downPath string) {
	Logger.Debug("start download shards...")
	wg := NewWaitGroup(d.threads)
	for idx, shard := range shards {
		wg.AddDelta()
		go func(wg *WaitGroup, url string, idx int, shard, downPath string) {
			if err := d.download(wg, url, idx, shard, downPath); err != nil {
				Logger.Error(errors.ErrorStack(err))
			}
		}(wg, url, idx, shard, downPath)
	}
	wg.Wait()
}

func (d *M3u8Downloader) merge() error {
	save, err := os.Create(filepath.Join(d.savePath, d.saveName))
	if err != nil {
		return errors.Trace(err)
	}
	defer save.Close()

	files, err := ioutil.ReadDir(d.downPath)
	if err != nil {
		return errors.Trace(err)
	}
	for _, f := range files {
		bytes, err := ioutil.ReadFile(filepath.Join(d.downPath, f.Name()))
		if err != nil {
			return errors.Trace(err)
		}
		if _, err := save.Write(bytes); err != nil {
			return errors.Trace(err)
		}
	}
	return nil
}

func (d *M3u8Downloader) ClearDebris() error {
	return os.RemoveAll(d.downPath)
}

func (d *M3u8Downloader) Run() error {
	if !d.checkUrl(d.url) {
		return fmt.Errorf("error url:%s", d.url)
	}

	retry := 0
	for retry == 0 || !d.done() {
		Logger.Debugf("---retry: %d", retry)

		Logger.Debugf("parsing mu38 url: %s", d.url)
		shards, err := d.parse(d.url, d.isShardFunc)
		if err != nil {
			return errors.Trace(err)
		}
		Logger.Debugf("parsed. this m3u8 file has %d shards", len(shards))

		if d.encrypt {
			panic("unsupported encrypt m3u8 file")
		}

		d.initCheckBitMap(shards)

		if err = d.mkdir(); err != nil {
			return errors.Trace(err)
		}

		d.downloadShards(d.url, shards, d.downPath)

		retry++

		if retry == d.maxtry {
			return fmt.Errorf("retry too much")
		}
	}

	Logger.Debug("start merging...")
	if err := d.merge(); err != nil {
		return errors.Trace(err)
	}

	if d.clearDebris {
		Logger.Debug("start Clearing Debris")
		if err := d.ClearDebris(); err != nil {
			return errors.Trace(err)
		}
	}
	Logger.Debug("done")
	return nil
}

func main() {
	// nunuyy5.org
	url := "https://b.baobuzz.com/m3u8/569128.m3u8?sign=4d5618aebd4dd9a59b0533e0603922d9"
	downloader := Default(url, "")
	if err := downloader.Run(); err != nil {
		Logger.Error(errors.ErrorStack(err))
	}
}

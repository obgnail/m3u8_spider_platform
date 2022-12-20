package downloader

import (
	"fmt"
	"github.com/juju/errors"
	"io/ioutil"
	"net/http"
	Url "net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

const (
	ua = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/108.0.0.0 Safari/537.36"

	shardFileFormat = "%05d.ts"
)

type M3u8Downloader struct {
	url         string
	downPath    string
	savePath    string
	saveName    string
	clearDebris bool
	threads     uint
	maxRetry    uint
	isShardFunc func(line string) (need bool)
	_totalShard int
	_encrypt    bool
	_bar        *ProcessBar
}

func Default(url, saveName string) *M3u8Downloader {
	return New(url, saveName, "", "", true, 16, 5, nil)
}

func New(url, saveName, downPath, savePath string, clearDebris bool, threads, maxRetry uint,
	isShardFunc func(line string) (need bool)) *M3u8Downloader {
	if len(saveName) == 0 {
		u, err := Url.Parse(url)
		if err != nil {
			panic(err)
		}
		result := strings.Split(u.Path, "/")
		saveName = fmt.Sprintf("%s.ts", result[len(result)-1])
	}
	if len(downPath) == 0 {
		downPath = "./Download"
	}
	downPath = filepath.Join(downPath, saveName)

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
		maxRetry:    maxRetry,
		_encrypt:    false,
	}
	d.SetIsShardFunc(isShardFunc)
	return d
}

func (d *M3u8Downloader) SetIsShardFunc(isShardFunc func(line string) (need bool)) {
	if isShardFunc == nil {
		isShardFunc = func(line string) (need bool) { return strings.HasPrefix(line, "http") }
	}
	d.isShardFunc = isShardFunc
}

func getFileMap(dirPath string) (map[string]struct{}, error) {
	files, err := ioutil.ReadDir(dirPath)
	if err != nil {
		return nil, errors.Trace(err)
	}
	fileMap := make(map[string]struct{})
	for _, f := range files {
		fileMap[f.Name()] = struct{}{}
	}
	return fileMap, nil
}

func (d *M3u8Downloader) filter(shards map[int]string) (map[int]string, error) {
	fileMap, err := getFileMap(d.downPath)
	if err != nil {
		return nil, errors.Trace(err)
	}
	for idx := 0; idx < d._totalShard; idx++ {
		fileName := fmt.Sprintf(shardFileFormat, idx)
		if _, ok := fileMap[fileName]; ok {
			delete(shards, idx)
		}
	}
	return shards, nil
}

func (d *M3u8Downloader) done() (bool, error) {
	fileMap, err := getFileMap(d.downPath)
	if err != nil {
		return false, errors.Trace(err)
	}
	for idx := 0; idx < d._totalShard; idx++ {
		fileName := fmt.Sprintf(shardFileFormat, idx)
		if _, ok := fileMap[fileName]; !ok {
			return false, nil
		}
	}
	return true, nil
}

func (d *M3u8Downloader) check(url string) bool {
	return strings.HasPrefix(url, "http")
}

func (d *M3u8Downloader) request(url string) (body []byte, err error) {
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("user-agent", ua)
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

func (d *M3u8Downloader) parseM3u8Url(url string, isShardFunc func(line string) (need bool)) (
	shards map[int]string, err error) {
	if isShardFunc == nil {
		panic("isShardFunc == nil")
	}
	resp, err := d.request(url)
	if err != nil {
		return nil, errors.Trace(err)
	}
	response := strings.Split(string(resp), "\n")

	var shardIdx = 0
	shards = make(map[int]string)
	for _, line := range response {
		if isShardFunc(line) == true {
			shards[shardIdx] = line
			shardIdx++
		}
		if strings.HasPrefix(line, "#EXT-X-KEY:") {
			d._encrypt = true
		}
	}
	if len(shards) == 0 {
		return nil, fmt.Errorf("len(shards) == 0")
	}
	d._totalShard = len(shards)
	return
}

func mkdir(Path string) error {
	if _, err := os.Stat(Path); os.IsNotExist(err) {
		if err = os.MkdirAll(Path, os.ModePerm); err != nil {
			return errors.Trace(err)
		}
	}
	return nil
}

func (d *M3u8Downloader) mkdir() (err error) {
	if err = mkdir(d.downPath); err != nil {
		return errors.Trace(err)
	}
	if err = mkdir(d.savePath); err != nil {
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

func (d *M3u8Downloader) downloadShard(wg *WaitGroup, m3u8Url string, shardIdx int, shardUrl, downPath string) error {
	defer wg.Done()

	res := strings.Split(m3u8Url, "/")
	baseUrl := strings.Join(res[:len(res)-1], "/")
	if !strings.HasPrefix(shardUrl, "http") {
		shardUrl = baseUrl + "/" + shardUrl
	}
	debrisName := path.Join(downPath, fmt.Sprintf(shardFileFormat, shardIdx))

	if _, err := os.Stat(debrisName); os.IsNotExist(err) {
		resp, err := d.request(shardUrl)
		if err != nil {
			Logger.Errorf("shard %d failed: %s", shardIdx, shardUrl)
			return errors.Trace(err)
		}
		if err = writeFile(debrisName, resp); err != nil {
			return errors.Trace(err)
		}
		d._bar.Add(1)
	}
	return nil
}

func (d *M3u8Downloader) downloadShards(m3u8Url string, shards map[int]string, downPath string) {
	wg := NewWaitGroup(int(d.threads))
	for shardIdx, shardUrl := range shards {
		wg.AddDelta()
		go ErrHandler(d.downloadShard(wg, m3u8Url, shardIdx, shardUrl, downPath))
	}
	wg.Wait()
}

func (d *M3u8Downloader) retry(maxRetry uint, f func() (stop bool, err error)) error {
	count := uint(0)
	for {
		stop, err := f()
		if stop || err != nil {
			return err
		}
		count++
		if count == maxRetry+1 {
			return fmt.Errorf("retry too much")
		}
		Logger.Warnf("[%d] Time(s) Retry...", count)
	}
}

func (d *M3u8Downloader) prepare() error {
	Logger.Debugf("[STEP0] check url")

	if !d.check(d.url) {
		return fmt.Errorf("error url: %s", d.url)
	}
	if err := d.mkdir(); err != nil {
		return errors.Trace(err)
	}
	return nil
}

func (d *M3u8Downloader) parse() (shards map[int]string, err error) {
	Logger.Debugf("[STEP1] parse m3u8 file: %s", d.url)

	interval := 3 * time.Second
	err = d.retry(d.maxRetry, func() (stop bool, err error) {
		shards, err = d.parseM3u8Url(d.url, d.isShardFunc)
		if err == nil {
			return true, nil
		}
		ErrHandler(err)
		time.Sleep(interval)
		return false, nil
	})
	if err != nil {
		return nil, errors.Trace(err)
	}
	if d._encrypt {
		return nil, fmt.Errorf("unsupported encrypt m3u8 file")
	}
	return shards, nil
}

func (d *M3u8Downloader) download(shards map[int]string) error {
	Logger.Debugf("[STEP2] download [%d] shards", len(shards))

	d._bar = NewBar(0, d._totalShard)
	d._bar.Start()
	err := d.retry(d.maxRetry, func() (stop bool, err error) {
		shards, err = d.filter(shards)
		if err != nil {
			return true, errors.Trace(err)
		}
		d._bar.Reset(d._totalShard - len(shards))
		d.downloadShards(d.url, shards, d.downPath)
		_done, err := d.done()
		if err != nil {
			return true, errors.Trace(err)
		}
		return _done, nil
	})
	return errors.Trace(err)
}

func (d *M3u8Downloader) merge() error {
	Logger.Debug("[STEP3] merge all shards")

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
		content, err := ioutil.ReadFile(filepath.Join(d.downPath, f.Name()))
		if err != nil {
			return errors.Trace(err)
		}
		if _, err := save.Write(content); err != nil {
			return errors.Trace(err)
		}
	}
	return nil
}

func (d *M3u8Downloader) clear() error {
	if d.clearDebris {
		Logger.Debug("[STEP4] clear debris")
		if err := os.RemoveAll(d.downPath); err != nil {
			return errors.Trace(err)
		}
	}
	return nil
}

// prepare -> parse -> download -> merge -> clear
func (d *M3u8Downloader) Run() (err error) {
	Logger.Infof("download: %s\n", d.saveName)

	if err = d.prepare(); err != nil {
		return errors.Trace(err)
	}
	shards, err := d.parse()
	if err != nil {
		return errors.Trace(err)
	}
	if err = d.download(shards); err != nil {
		return errors.Trace(err)
	}
	if err = d.merge(); err != nil {
		return errors.Trace(err)
	}
	if err = d.clear(); err != nil {
		return errors.Trace(err)
	}
	Logger.Infof("fininsh: %s\n", d.saveName)
	return nil
}

func ErrHandler(err error) {
	if err != nil {
		Logger.Error(errors.ErrorStack(err))
	}
}

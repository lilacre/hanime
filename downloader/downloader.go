package downloader

import (
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"sync"

	"github.com/lilacre/hanime/utility"
	"github.com/sirupsen/logrus"
)

const (
	jobUNDO = iota
	jobDONE = iota
)

type Downloader struct {
	Client         *http.Client
	SegmentURL     []string
	StoragePath    string
	FilelistsPath  string
	OutputFilePath string
	TmpPATH        string
	MaxPROCS       int
	IV             string
	Logger         *logrus.Logger
}

// Download
func (downloader *Downloader) Download() {
	if downloader.MaxPROCS == 0 {
		downloader.MaxPROCS = 100
	}

	// Dump to filelists
	dumpFilelists(len(downloader.SegmentURL), downloader.FilelistsPath, downloader.TmpPATH)

	// Download TS
	runtime.GOMAXPROCS(downloader.MaxPROCS)
	wg := sync.WaitGroup{}
	progressBar := utility.ProgressBar{
		MaxLength: len(downloader.SegmentURL),
		Logger:    downloader.Logger,
	}
	for k, v := range downloader.SegmentURL {
		wg.Add(1)
		go func(url string, index int) {
			var jobSTATUS int = jobUNDO
			for jobSTATUS != jobDONE {
				err := downloadTS(url, downloader.Client, filepath.Join(downloader.TmpPATH, strconv.Itoa(index)+".ts"), downloader.IV)
				if err == nil {
					jobSTATUS = jobDONE
					progressBar.Next()
					wg.Done()
				}
			}
		}(v, k)
	}
	wg.Wait()

	// Concat
	stderr, err := concat(downloader.FilelistsPath, downloader.OutputFilePath)
	if err != nil {
		downloader.Logger.Warnln(stderr)
		downloader.Logger.Warnln("concat error")
	} else {
		downloader.Logger.Infoln("concat succeed")
	}

	// Clean
	err = clean(downloader.TmpPATH)
	if err != nil {
		downloader.Logger.Panicln(err)
	}
	downloader.Logger.Infoln("clean succeed")
}

func downloadTS(url string, client *http.Client, path string, iv string) error {
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	content, err := decrypt(body, iv)
	if err != nil {
		return err
	}

	fo, err := os.Create(path)
	if err != nil {
		return err
	}
	defer fo.Close()

	if _, err := fo.Write(content); err != nil {
		return err
	}
	return nil
}

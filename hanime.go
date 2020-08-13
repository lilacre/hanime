package main

import (
	"bytes"
	"flag"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	ss "strings"
	"time"

	"github.com/lilacre/hanime/downloader"
	"github.com/shiena/ansicolor"
	"github.com/sirupsen/logrus"

	m3u8 "github.com/grafov/m3u8"
)

const (
	iv   = "0123456701234567"
	host = "https://hanime.tv"
)

var (
	targetURL      string
	proxyURL       string
	outputFILE     string
	storagePATH    string
	tmpPATH        string
	outputFILEPATH string
	logger         *logrus.Logger
)

func makeDirIfNotExists(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return os.Mkdir(path, os.ModeDir|0755)
	}
	return nil
}

func fileExists(outputPATH string) bool {
	info, err := os.Stat(outputPATH)
	if os.IsNotExist(err) {
		return false
	}

	return !info.IsDir()
}

func init() {
	flag.StringVar(&targetURL, "url", "", "hanime hentai url")
	flag.StringVar(&outputFILE, "output", "", "output file name")
	flag.StringVar(&proxyURL, "proxy", "", "proxy forwarding")
	flag.StringVar(&storagePATH, "dir", "", "dir to store file")
	flag.Parse()

	// Config logger
	logger = logrus.New()
	logger.SetLevel(logrus.InfoLevel)
	logger.SetFormatter(&logrus.TextFormatter{
		ForceColors: true,
	})
	logger.SetOutput(ansicolor.NewAnsiColorWriter(os.Stdout))

	if targetURL == "" {
		logger.Panicln("you must specify a url to download hentai from hanime.tv")
	}

	if storagePATH == "" {
		logger.Panicln("you must specify a directory path to save file")
	}

	if i := strings.Index(targetURL, "?"); i > -1 {
		targetURL = targetURL[:i]
	}

	if outputFILE == "" {
		strLIST := ss.Split(targetURL, "/")
		outputFILE = strLIST[len(strLIST)-1]
	}

	// if outputFILE == "" {
	// 	strLIST := make([]string, 2, 2)
	// 	if i := strings.Index(targetURL, "?"); i > -1 {
	// 		strLIST = ss.Split(targetURL[:i], "/")
	// 	} else {
	// 		strLIST = ss.Split(targetURL, "/")
	// 	}
	// 	outputFILE = strLIST[len(strLIST)-1]
	// }

	outputFILEPATH = filepath.Join(storagePATH, outputFILE+".mp4")
	if i := fileExists(outputFILEPATH); i == true {
		logger.Panicln("file " + outputFILEPATH + " exists")
	}
	tmpPATH = filepath.Join(storagePATH, outputFILE+"_tmp")

	makeDirIfNotExists(storagePATH)
	makeDirIfNotExists(tmpPATH)
}

func getM3U8URL(client *http.Client, targetURL string) (string, error) {
	request, err := http.NewRequest("GET", targetURL, nil)
	request.Header.Add("User-Agent", `Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/84.0.4147.105 Safari/537.36`)
	request.Header.Add("Origin", `https://hanime.tv`)
	request.Header.Add("Accept", `text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9`)
	request.Header.Add("upgrade-insecure-requests", `1`)

	if err != nil {
		return "", err
	}

	resp, err := client.Do(request)
	defer func() {
		if err := resp.Body.Close(); err != nil {
			logger.Panicln(err)
		}
	}()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	re := regexp.MustCompile(`[0-9]+\.m3u8`)
	result := re.FindString(string(body))
	logger.Infoln(string(result))
	return `https://weeb.hanime.tv/weeb-api-cache/api/v8/m3u8s/` + string(result), nil
}

func main() {
	logger.Infoln("start")
	tr := &http.Transport{
		IdleConnTimeout:    30 * time.Second,
		DisableCompression: true,
	}
	if proxyURL != "" {
		proxy, err := url.Parse(proxyURL)
		if err != nil {
			logger.Panicln(err)
		}
		tr.Proxy = http.ProxyURL(proxy)
	}
	client := &http.Client{Transport: tr}

	// get m3u8 url
	m3u8URL, err := getM3U8URL(client, targetURL)
	if err != nil {
		logger.Panicln(err)
	}
	// end get m3u8 url

	// get m3u8
	resp, err := client.Get(m3u8URL)
	if err != nil {
		logger.Panicln(err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Panicln(err)
	}
	resp.Body.Close()
	// end get m3u8

	raw := bytes.NewBuffer(body)
	p, listType, err := m3u8.Decode(*raw, true)
	if err != nil {
		logger.Panicln(err)
	}

	// var playlists []string
	switch listType {
	case m3u8.MEDIA:
		mediapl := p.(*m3u8.MediaPlaylist)
		// log.Printf("%+v\n", mediapl)
		segmentsURLRAW := mediapl.Segments
		segmentsURL := make([]string, 0, len(segmentsURLRAW))
		for _, v := range segmentsURLRAW {
			if v != nil {
				segmentsURL = append(segmentsURL, v.URI)
			}
		}
		filelistsPATH := filepath.Join(tmpPATH, "filelists.txt")
		dl := downloader.Downloader{
			Client:         client,
			SegmentURL:     segmentsURL,
			StoragePath:    storagePATH,
			FilelistsPath:  filelistsPATH,
			OutputFilePath: outputFILEPATH,
			TmpPATH:        tmpPATH,
			MaxPROCS:       100,
			IV:             iv,
			Logger:         logger,
		}
		dl.Download()
	case m3u8.MASTER:
		masterpl := p.(*m3u8.MasterPlaylist)
		logger.Infof("%+v\n", masterpl)
	}
}

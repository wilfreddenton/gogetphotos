package main

import (
	"encoding/json"
	"fmt"
	"gopkg.in/cheggaaa/pb.v1"
	"io/ioutil"
	"math"
	"os"
	"strconv"
	"time"
)

var baseDir string

type Config struct {
	ImgurClientId     string   `json:"imgurClientId"`
	UserAgent         string   `json:"userAgent"`
	BaseDir           string   `json:"baseDir"`
	DefaultSubreddits []string `json:"defaultSubreddits"`
}

func main() {
	start := time.Now()
	goDir := "/src/github.com/wilfreddenton/gogetphotos/"
	dat, err := ioutil.ReadFile(os.Getenv("GOPATH") + goDir + "config.json")
	if err != nil {
		panic(err.Error())
	}
	config := new(Config)
	json.Unmarshal(dat, &config)
	logFile, err := os.Create(os.Getenv("GOPATH") + goDir + "log.txt")
	if err != nil {
		panic(err.Error())
	}
	args := os.Args
	subreddits := config.DefaultSubreddits
	if len(args) > 1 {
		subreddits = os.Args[1:]
	}
	baseDir = config.BaseDir
	counts := make(chan int)
	done := make(chan *Fetcher)
	resources := make(chan *Resource)
	images := make(chan *Resource)
	files := make(chan *Resource)
	resBalancer := NewBalancer("getter", 3, resources, images, files) // numWorkers, in, out, end
	imgBalancer := NewBalancer("saver", 3, images, files, files)
	bar := pb.New(1).Prefix("images")
	bar.Total = int64(math.Inf(1))
	bars := append(resBalancer.bars, bar)
	barPool, err := pb.StartPool(bars...)
	if err != nil {
		panic(err)
	}
	fetchers := make([]*Fetcher, len(subreddits))
	for i, subreddit := range subreddits {
		fetchers[i] = NewFetcher(config.ImgurClientId, config.UserAgent, subreddit)
		go func(i int) {
			fetchers[i].FetchList(resources, counts, done)
		}(i)
	}
	go resBalancer.Balance()
	go imgBalancer.Balance()
	total, byteCount, imageDoneCount, imageErrorCount, fetcherDoneCount := 0, 0.0, 0, 0, 0
	for {
		select {
		case r := <-files:
			if r.Error != nil {
				fmt.Fprintln(logFile, "\nCould not get image after 3 attempts:\nid: "+r.Id+", title: "+r.Title+", subreddit: "+r.Subreddit+", url: "+r.Url+", error: "+r.Error.Error())
				imageErrorCount += 1
			} else {
				imageDoneCount += 1
			}
			bar.Increment()
		case _ = <-done:
			fetcherDoneCount += 1
			if fetcherDoneCount == len(fetchers) {
				bar.Total = int64(total)
			}
		case count := <-counts:
			total += 1
			byteCount += float64(count)
		}
		if imageDoneCount+imageErrorCount == total {
			for _, b := range bars {
				b.Finish()
			}
			break
		}
	}
	barPool.Stop()
	unit := "Bs"
	if byteCount > 1e9 {
		unit = "GBs"
		byteCount /= 1e9
	} else if byteCount > 1e6 {
		unit = "MBs"
		byteCount /= 1e6
	} else if byteCount > 1e3 {
		unit = "KBs"
		byteCount /= 1e3
	}
	elapsed := time.Since(start)
	fmt.Println("Got " + strconv.Itoa(imageDoneCount) + " images (" + strconv.FormatFloat(byteCount, 'f', 2, 64) + " " + unit + ") with " + strconv.Itoa(imageErrorCount) + " errors in " + elapsed.String() + ".")
	fmt.Println("Finished. Enjoy your photos.")
}

package main

import (
	"fmt"
	"gopkg.in/cheggaaa/pb.v1"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"sync"
)

type Resource struct {
	Data      []byte
	Id        string
	Url       string
	Title     string
	Ext       string
	Subreddit string
	Size      int
	attempts  int
	Error     error
}

func (r *Resource) errorHook(err error) (*Resource, error) {
	r.attempts += 1
	r.Error = err
	return r, err
}

func (r *Resource) GetImage(bar *pb.ProgressBar, m *sync.Mutex) (*Resource, error) {
	res, err := http.Get(r.Url)
	if err != nil {
		return r.errorHook(err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return r.errorHook(fmt.Errorf("%s: %d", r.Url, res.StatusCode))
	}
	m.Lock()
	bar.Total += int64(r.Size)
	m.Unlock()
	reader := bar.NewProxyReader(res.Body)
	body, err := ioutil.ReadAll(reader)
	if err != nil {
		return r.errorHook(err)
	}
	r.Data = body
	r.attempts = 0
	r.Error = nil
	return r, nil
}

func (r *Resource) SaveImage() (*Resource, error) {
	title := strings.Replace(r.Title, "/", " ", -1)
	if len(title) > 200 {
		title = title[0:200]
	}
	path := baseDir + "/" + title + "_" + r.Subreddit + "_" + r.Id + r.Ext
	file, err := os.Create(path)
	if err != nil {
		return r.errorHook(err)
	}
	defer file.Close()
	_, err = file.Write(r.Data)
	if err != nil {
		return r.errorHook(err)
	}
	r.attempts = 0
	r.Error = nil
	return r, err
}

func (r *Resource) fn(bar *pb.ProgressBar, m *sync.Mutex) (*Resource, error) {
	if len(r.Data) == 0 {
		return r.GetImage(bar, m)
	} else {
		return r.SaveImage()
	}
}

func NewResource(id, url, title, ext, subreddit string, size int) *Resource {
	return &Resource{make([]byte, 0), id, url, title, ext, subreddit, size, 0, nil}
}

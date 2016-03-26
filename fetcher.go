package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"time"
)

type Image struct {
	Id   string `json:"id"`
	Url  string `json:"link"`
	Size int    `json:"size"`
}

type Album struct {
	Images []Image `json:"data"`
}

type ImageInfo struct {
	Data struct {
		Url  string `json:"link"`
		Size int    `json:"size"`
	} `json:"data"`
}

type List struct {
	Data struct {
		Children []struct {
			Data struct {
				Url   string `json:"url"`
				Title string `json:"title"`
			} `json:"data"`
		} `json:"children"`
	} `json:"data"`
}

type Fetcher struct {
	ImgurBaseUrl  string
	ImgurClientId string
	RedditBaseUrl string
	UserAgent     string
	Subreddit     string
	client        http.Client
}

func (f *Fetcher) FetchImageInfo(id string) (*ImageInfo, error) {
	url := f.ImgurBaseUrl + "image/" + id
	imageInfo := new(ImageInfo)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return imageInfo, err
	}
	req.Header.Set("Authorization", "Client-ID "+f.ImgurClientId)
	res, err := f.client.Do(req)
	if err != nil {
		return imageInfo, err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return imageInfo, fmt.Errorf("%s: %d", url, res.StatusCode)
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return imageInfo, err
	}
	json.Unmarshal(body, &imageInfo)
	return imageInfo, nil
}

func (f *Fetcher) FetchAlbum(id string) (*Album, error) {
	url := f.ImgurBaseUrl + "album/" + id + "/images"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Client-ID "+f.ImgurClientId)
	res, err := f.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return nil, fmt.Errorf("%s: %d", url, res.StatusCode)
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	album := new(Album)
	json.Unmarshal(body, &album)
	return album, nil
}

func (f *Fetcher) FetchList(send chan *Resource, counts chan int, done chan *Fetcher) {
	url := f.RedditBaseUrl + "r/" + f.Subreddit + "/hot.json"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		panic(err.Error())
	}
	req.Header.Set("User-Agent", f.UserAgent)
	res, err := f.client.Do(req)
	if err != nil {
		panic(err.Error())
	}
	if res.StatusCode != 200 {
		panic(err.Error())
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		panic(err.Error())
	}
	res.Body.Close()
	list := new(List)
	json.Unmarshal(body, &list)
	r1 := regexp.MustCompile(`imgur.com`)
	r2 := regexp.MustCompile(`imgur.com/a/`)
	r3 := regexp.MustCompile(`imgur.com\/(a\/)?([a-zA-Z0-9]+)(\.(jpg|gif|png|gifv))?(\?.*)?$`)
	rate := time.Second / 10
	throttle := time.Tick(rate)
	for _, child := range list.Data.Children {
		link := []byte(child.Data.Url)
		matches := r3.FindSubmatch(link)
		if r1.Find(link) != nil && len(matches) > 0 {
			id := string(matches[2])
			if r2.Find(link) != nil {
				album, err := f.FetchAlbum(id)
				if err != nil {
					fmt.Println("Failed to Fetch Album")
				} else {
					for i, image := range album.Images {
						ext := string(r3.FindSubmatch([]byte(image.Url))[3])
						if ext == ".gifv" {
							ext = ext[0 : len(ext)-1]
							image.Url = image.Url[0 : len(image.Url)-1]
						} else if ext == ".gif" {
							image.Url = "http://i.imgur.com/" + image.Id + ext
						}
						<-throttle
						counts <- image.Size
						send <- NewResource(image.Id, image.Url, child.Data.Title+"_"+strconv.Itoa(i), ext, f.Subreddit, image.Size)
					}
				}
			} else {
				imageInfo, err := f.FetchImageInfo(id)
				if err != nil {
					fmt.Println("Failed to fetch image link")
				} else {
					url := imageInfo.Data.Url
					size := imageInfo.Data.Size
					ext := string(r3.FindSubmatch([]byte(url))[3])
					if ext == ".gifv" {
						ext = ext[0 : len(ext)-1]
						url = url[0 : len(url)-1]
					} else if ext == ".gif" {
						url = "http://i.imgur.com/" + id + ext
					}
					counts <- size
					send <- NewResource(id, url, child.Data.Title, ext, f.Subreddit, size)
				}
			}
		}
	}
	done <- f
}

func NewFetcher(imgurClientId, userAgent, subreddit string) *Fetcher {
	client := http.Client{}
	fetcher := Fetcher{"https://api.imgur.com/3/", imgurClientId, "https://api.reddit.com/", userAgent, subreddit, client}
	return &fetcher
}

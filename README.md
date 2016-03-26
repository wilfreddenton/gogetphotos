# gogetphotos

## Installation

1. run `go get github.com/wilfreddenton/gogetphotos`
2. create `config.json` in `$GOPATH/src/github.com/wilfreddenton/gogetphotos/`

example `config.json`

```
{
  "imgurClientId": "", 
  "userAgent": "",
  "baseDir": "",
  "defaultSubreddits": [
    "pics",
    "funny",
    "adviceanimals"
  ]
}
```

**imgurClientId**: the client id of an imgur app (register an app on their site)
**userAgent**: the user agent you want to make requests to reddit (some unique identifying string)
**baseDir**: the directory the downloaded photos are saved to
**defaultSubreddits**: when you run the gogetphotos without arguments these subreddits are downloaded

## Usage

Running just `gogetphotos` downloads the photos from the default subreddits in the `config.json`

Running `gogetphotos anime gaming` will download photos from the front page of r/anime and r/gaming

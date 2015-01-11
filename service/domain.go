package service

import (
    "github.com/jelinden/rssFetcher/rss"
)
var Conf struct {
	MongoUrl string
}

type Result struct {
	Items       []rss.Item
	Description string
	Lang        int
}

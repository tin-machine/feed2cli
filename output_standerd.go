package main

/*
Slackに出力する
*/

import (
	"fmt"
	"log"
	"time"

	"github.com/gorilla/feeds"
	"github.com/mmcdole/gofeed"
)

// OutputStanderd は、フィードを標準出力に出力します。
func OutputStanderd(feed []*gofeed.Feed) {

	for _, f := range feed {
		c1 := f
		now := time.Now()
		output_feed := &feeds.Feed{
			Title:       c1.Title,
			Link:        &feeds.Link{Href: c1.Link},
			Description: c1.Description,
			Created:     now,
		}

		for _, v := range c1.Items {
			item := &feeds.Item{
				Title:       v.Title,
				Link:        &feeds.Link{Href: v.Link},
				Description: v.Description,
				Created:     now,
			}
			output_feed.Add(item)
		}

		rss, err := output_feed.ToRss()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Print(rss)
	}

}

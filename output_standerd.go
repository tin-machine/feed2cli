package main

import (
	"fmt"
	"log"
	"time"

	"github.com/gorilla/feeds"
	"github.com/mmcdole/gofeed"
)

/*
標準出力に出力
*/
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
			// 日付を取得する https://leben.mobi/go/time/go-programming/
			const format1 = "2006-01-02T15:04:05Z"
			t1, _ := time.Parse(format1, v.Published)
			item := &feeds.Item{
				Title:       v.Title,
				Link:        &feeds.Link{Href: v.Link},
				Description: v.Description,
				Created:     t1,
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

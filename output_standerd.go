package main

/*
標準出力に出力する
*/

import (
	"fmt"
	"log"
	"time"

	"github.com/gorilla/feeds"
)

/*
標準出力に出力
*/
func OutputStanderd(feed []sortableFeed) {

	c1 := &feed[0]
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
	// string 型 → []byte 型
	// b := []byte(rss)

	fmt.Print(rss)
	// err2 := ioutil.WriteFile(file, b, 0666)
	// if err2 != nil {
	// 	fmt.Println(os.Stderr, err)
	// 	os.Exit(1)
	// }

}

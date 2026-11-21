package main

import (
	"encoding/json"
	"io"
)

// OutputJSONLTo は記事を一件ずつ JSONL のレコードとして書き出す。
func OutputJSONLTo(w io.Writer, data interface{}) error {
	items, err := outputFeedItems(data)
	if err != nil {
		return err
	}
	encoder := json.NewEncoder(w)
	for _, item := range items {
		if err := encoder.Encode(NewFeedItemJSONLRecord(item)); err != nil {
			return err
		}
	}
	return nil
}

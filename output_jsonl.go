package main

import (
	"encoding/json"
	"io"
)

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

package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
)

// OutputJSON は、フィルタリングされたアイテムのリストを受け取り、
// その内容をJSON形式で標準出力します。デバッグ用途を想定しています。
func OutputJSON(data interface{}) {
	if err := OutputJSONTo(os.Stdout, data); err != nil {
		log.Fatalf("JSONへのシリアライズに失敗しました: %v", err)
	}
}

func OutputJSONTo(w io.Writer, data interface{}) error {
	itemsToProcess := FeedItemsFromData(data)

	jsonData, err := json.MarshalIndent(itemsToProcess, "", "  ")
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(w, string(jsonData))
	return err
}

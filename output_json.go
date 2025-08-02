package main

import (
	"encoding/json"
	"fmt"
	"log"
)

// OutputJSON は、フィルタリングされたアイテムのリストを受け取り、
// その内容をJSON形式で標準出力します。デバッグ用途を想定しています。
func OutputJSON(data interface{}) {
	itemsToProcess := convertToFilteredItems(data)

	// JSONにシリアライズ
	jsonData, err := json.MarshalIndent(itemsToProcess, "", "  ")
	if err != nil {
		log.Fatalf("JSONへのシリアライズに失敗しました: %v", err)
	}

	fmt.Println(string(jsonData))
}

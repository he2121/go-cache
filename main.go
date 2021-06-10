package main

import (
	"fmt"
	"log"
	"net/http"
)

func main()  {
	var db = map[string]string{
		"Tom":  "630",
		"Jack": "589",
		"Sam":  "567",
	}
	_ = NewGroup("score", 2<<10, GetterFunc(
		func(key string) ([]byte, error) {
			log.Println("[SlowDB] search key", key)
			if value, ok := db[key]; ok {
				return []byte(value), nil
			}
			log.Println("record not exist")
			return []byte{}, fmt.Errorf("%s not exist", key)
		},
	))
	addr := "127.0.0.1:9999"
	pool := NewHttpPool(addr)
	http.ListenAndServe(addr, pool)
}

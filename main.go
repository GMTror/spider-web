package main

import (
	"log"
)

var (
	VERSION string = "UNKNOWN"
	HASH    string = "UNKNOWN"
)

func main() {
	log.Printf("VERSION: %s", VERSION)
	log.Printf("HASH: %s", HASH)
}

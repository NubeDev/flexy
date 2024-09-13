package main

import (
	"log"
	"time"
)

func main() {
	for {
		log.Println("hello my-app")
		time.Sleep(10 * time.Second)
	}
}

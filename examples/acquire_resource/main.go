package main

import (
	"fmt"

	"github.com/jpg013/ratelimit"
)

func main() {
	manager, err := ratelimit.NewManager("crawlera")

	if err != nil {
		panic(err)
	}

	fmt.Println("hello there")
}

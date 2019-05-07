package main

import (
	"fmt"
	"time"

	"github.com/kn100/charlotte/sitemap"
)

func main() {
	fmt.Println("Generating sitemap")
	timeout := time.Duration(10 * time.Second)
	sm := sitemap.MakeSiteMap("https://kn100.me/", 5, timeout)

	fmt.Println(sm.String())
	//fmt.Println(sm.JSON())
}

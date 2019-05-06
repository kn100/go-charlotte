package main

import (
	"fmt"

	"github.com/kn100/charlotte/sitemap"
	"github.com/pkg/profile"
)

func main() {
	fmt.Println("Generating sitemap")
	defer profile.Start(profile.MemProfile).Stop()

	sm := sitemap.MakeSiteMap("https://monzo.com/", 2)
	fmt.Println(sm.String())
	//fmt.Println(sm.JSON())
}

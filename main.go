package main

import (
	"os"

	_ "github.com/pirmd/gostore/media/books"
	_ "github.com/pirmd/gostore/module/all"
)

func main() {
	cfg := newConfig()
	app := newApp(cfg)
	app.MustRun(os.Args[1:])
}

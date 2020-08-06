package main

import (
	"os"

	_ "github.com/pirmd/gostore/media/books"

	_ "github.com/pirmd/gostore/modules/checker"
	_ "github.com/pirmd/gostore/modules/dehtmlizer"
	_ "github.com/pirmd/gostore/modules/dupfinder"
	_ "github.com/pirmd/gostore/modules/fetcher"
	_ "github.com/pirmd/gostore/modules/organizer"
	_ "github.com/pirmd/gostore/modules/scrubber"
)

func main() {
	cfg := newConfig()
	app := newApp(cfg)
	app.MustRun(os.Args[1:])
}

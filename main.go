package main

import (
	"os"

	_ "github.com/pirmd/gostore/media/books"
	_ "github.com/pirmd/gostore/modules/checker"
	_ "github.com/pirmd/gostore/modules/dehtmlizer"
	_ "github.com/pirmd/gostore/modules/fetcher"
	_ "github.com/pirmd/gostore/modules/organizer"
	_ "github.com/pirmd/gostore/modules/scrubber"

	_ "github.com/blevesearch/bleve/analysis/lang/en"
	_ "github.com/blevesearch/bleve/analysis/lang/fr"
)

func main() {
	cfg := newConfig()
	app := newApp(cfg)
	app.MustRun(os.Args[1:])
}

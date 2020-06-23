//go:generate go run manpage_generate.go cmd.go gostore.go
package main

import (
	"os"

	_ "github.com/pirmd/gostore/media/books"
	_ "github.com/pirmd/gostore/modules/dehtmlizer"
	_ "github.com/pirmd/gostore/modules/googlebooks"
	_ "github.com/pirmd/gostore/modules/organizer"

	_ "github.com/blevesearch/bleve/analysis/lang/en"
	_ "github.com/blevesearch/bleve/analysis/lang/fr"
)

func main() {
	cfg := newConfig()
	app := newApp(cfg)
	app.MustRun(os.Args[1:])
}

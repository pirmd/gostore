//XXX go:generate go run manpage_generate.go cmd.go config.go ui.go
package main

import (
	_ "github.com/pirmd/gostore/media/books"

	_ "github.com/pirmd/gostore/modules/dehtmlizer"
	_ "github.com/pirmd/gostore/modules/organizer"
)

func main() {
	gostoreApp.MustRun()
}

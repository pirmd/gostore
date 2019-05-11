//go:generate go run manpage_generate.go cmd.go config.go ui.go
package main

import (
	_ "github.com/pirmd/gostore/media/epub"
)

func main() {
	gostore.MustRun()
}

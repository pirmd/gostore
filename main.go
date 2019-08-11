//go:generate go run manpage_generate.go cmd.go config.go ui.go
package main

import (
	_ "github.com/pirmd/gostore/media/books"
)

func main() {
	gostoreApp.MustRun()
}

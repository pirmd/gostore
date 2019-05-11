// +build ignore

//Generate the manpage for 'godex' application
//You have to run it in together with cmd.go and its dependencies, for
//example 'go run manpage_generate.go cmd.go config.go'
package main

import (
	"github.com/pirmd/cli/app"
)

func main() {
	app.GenerateManpage(gostore)
}

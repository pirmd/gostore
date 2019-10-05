// +build ignore

package main

import (
	"github.com/pirmd/cli/app"
)

func main() {
	cfg := newConfig()
	cmd := newApp(cfg, nil)

	app.GenerateManpage(cmd)
	app.GenerateHelpFile(cmd)
}

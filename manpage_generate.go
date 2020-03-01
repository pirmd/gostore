// +build ignore

package main

import (
	"github.com/pirmd/clapp"
)

func main() {
	cfg := newConfig()
	cmd := newApp(cfg)

	clapp.GenerateManpage(cmd)
	clapp.GenerateHelpFile(cmd)
}

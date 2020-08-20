// Package all eases import of all available modules in the main
// application.
package all

import (
	// Silently import all available modules.
	_ "github.com/pirmd/gostore/modules/checker"
	_ "github.com/pirmd/gostore/modules/dehtmlizer"
	_ "github.com/pirmd/gostore/modules/dupfinder"
	_ "github.com/pirmd/gostore/modules/fetcher"
	_ "github.com/pirmd/gostore/modules/organizer"
	_ "github.com/pirmd/gostore/modules/scrubber"
)

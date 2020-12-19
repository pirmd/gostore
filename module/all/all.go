// Package all eases import of all available module in the main
// application.
package all

import (
	// Silently import all available module.
	_ "github.com/pirmd/gostore/module/checker"
	_ "github.com/pirmd/gostore/module/dehtmlizer"
	_ "github.com/pirmd/gostore/module/dupfinder"
	_ "github.com/pirmd/gostore/module/fetcher"
	_ "github.com/pirmd/gostore/module/hasher"
	_ "github.com/pirmd/gostore/module/mdatareader"
	_ "github.com/pirmd/gostore/module/normalizer"
	_ "github.com/pirmd/gostore/module/organizer"
	_ "github.com/pirmd/gostore/module/scrubber"
)

package ui

// UserInterfacer represents any User Interface.
type UserInterfacer interface {
	// Printf displays a message to the user (has same behaviour than fmt.Printf)
	Printf(string, ...interface{})

	// PrettyPrint displays values from the provided map
	PrettyPrint(...map[string]interface{})

	// PrettyDiff displays provided maps, highlighting their differences
	PrettyDiff(map[string]interface{}, map[string]interface{})

	// Edit spawns an editor to modify the given map.
	Edit([]map[string]interface{}) ([]map[string]interface{}, error)

	// Merge spawns an editor that displays two maps and highlight their diff to
	// facilitate change inspection and merge operation between them.
	Merge(map[string]interface{}, map[string]interface{}) (map[string]interface{}, error)
}

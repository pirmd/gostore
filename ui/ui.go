package ui

// UserInterfacer represents any User Interface.
type UserInterfacer interface {
	// Printf displays a message to the user (has same behaviour than fmt.Printf)
	Printf(string, ...interface{})

	// PrettyPrint displays values from the provided map
	PrettyPrint(...map[string]interface{})

	// PrettyDiff displays provided maps, highlighting their differences
	PrettyDiff(map[string]interface{}, map[string]interface{})

	// Edit spawns an editor dialog to modified provided map
	Edit(map[string]interface{}) (map[string]interface{}, error)

	// Merge spawns a dialog to merge two maps into one
	Merge(map[string]interface{}, map[string]interface{}) (map[string]interface{}, error)
}


# NAME
gostore - A command-line minimalist media collection manager.

# SYNOPSIS
**gostore** [--**debug**] [--**root**=**ROOT**] [--**style**=**name|list|full|json**] **help**
**gostore** [--**debug**] [--**root**=**ROOT**] [--**style**=**name|list|full|json**] **version**
**gostore** [--**debug**] [--**root**=**ROOT**] [--**style**=**name|list|full|json**] **import** [--**edit**] [--**dry-run**] **media**
**gostore** [--**debug**] [--**root**=**ROOT**] [--**style**=**name|list|full|json**] **get** [--**from-file**] **name**
**gostore** [--**debug**] [--**root**=**ROOT**] [--**style**=**name|list|full|json**] **list**
**gostore** [--**debug**] [--**root**=**ROOT**] [--**style**=**name|list|full|json**] **search** **query**
**gostore** [--**debug**] [--**root**=**ROOT**] [--**style**=**name|list|full|json**] **update** **name**
**gostore** [--**debug**] [--**root**=**ROOT**] [--**style**=**name|list|full|json**] **delete** **name**
**gostore** [--**debug**] [--**root**=**ROOT**] [--**style**=**name|list|full|json**] **export** **name** *dst*
**gostore** [--**debug**] [--**root**=**ROOT**] [--**style**=**name|list|full|json**] **check**

# DESCRIPTION
A command-line minimalist media collection manager.

# OPTIONS
--**debug**
: Show debug information.

--**root**=**ROOT**
: Path to the root of the collection.

--**style**=**name|list|full|json**
: Style for printing records' details.

# COMMANDS
**help**
: Show usage information.

**version**
: Show version information.

**import** [<flags>] **media**
: Import a new media into the collection.

**get** [<flags>] **name**
: retrieve information about any collection's record.

**list**
: List all the records from the collection.

**search** **query**
: Search the collection.

**update** **name**
: update an existing record from the collection.

**delete** **name**
: delete a record from the collection.

**export** **name** **dst**
: export a media from the collection to the given destination.

**check**
: Verify collection's consistency.

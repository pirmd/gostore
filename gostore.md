
# NAME
gostore - A command-line minimalist media collection manager.

# SYNOPSIS
**gostore** [--**debug**] [--**root**=**ROOT**] [--**style**=**full|json|name|list**] **help**
**gostore** [--**debug**] [--**root**=**ROOT**] [--**style**=**full|json|name|list**] **version**
**gostore** [--**debug**] [--**root**=**ROOT**] [--**style**=**full|json|name|list**] **import** [--**auto**] [--**dry-run**] **media**
**gostore** [--**debug**] [--**root**=**ROOT**] [--**style**=**full|json|name|list**] **get** [--**from-file**] **name**
**gostore** [--**debug**] [--**root**=**ROOT**] [--**style**=**full|json|name|list**] **list**
**gostore** [--**debug**] [--**root**=**ROOT**] [--**style**=**full|json|name|list**] **search** **query**
**gostore** [--**debug**] [--**root**=**ROOT**] [--**style**=**full|json|name|list**] **edit** **name**
**gostore** [--**debug**] [--**root**=**ROOT**] [--**style**=**full|json|name|list**] **delete** **name**
**gostore** [--**debug**] [--**root**=**ROOT**] [--**style**=**full|json|name|list**] **export** **name** *dst*
**gostore** [--**debug**] [--**root**=**ROOT**] [--**style**=**full|json|name|list**] **check**

# DESCRIPTION
A command-line minimalist media collection manager.

# OPTIONS
--**debug**
: Show debug information.

--**root**=**ROOT**
: Path to the root of the collection.

--**style**=**full|json|name|list**
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

**edit** **name**
: edit an existing record from the collection.

**delete** **name**
: delete a record from the collection.

**export** **name** **dst**
: export a media from the collection to the given destination.

**check**
: Verify collection's consistency.

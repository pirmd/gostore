
# NAME
gostore - A command-line minimalist media collection manager.

# SYNOPSIS
**gostore** [--**debug**] [--**root**=**ROOT**] [--**style**=**list|full|json|name**] **help**
**gostore** [--**debug**] [--**root**=**ROOT**] [--**style**=**list|full|json|name**] **version**
**gostore** [--**debug**] [--**root**=**ROOT**] [--**style**=**list|full|json|name**] **import** [--**auto**] [--**dry-run**] **media**
**gostore** [--**debug**] [--**root**=**ROOT**] [--**style**=**list|full|json|name**] **info** [--**from-file**] **name**
**gostore** [--**debug**] [--**root**=**ROOT**] [--**style**=**list|full|json|name**] **list**
**gostore** [--**debug**] [--**root**=**ROOT**] [--**style**=**list|full|json|name**] **search** **query**
**gostore** [--**debug**] [--**root**=**ROOT**] [--**style**=**list|full|json|name**] **edit** **name**
**gostore** [--**debug**] [--**root**=**ROOT**] [--**style**=**list|full|json|name**] **delete** **name**
**gostore** [--**debug**] [--**root**=**ROOT**] [--**style**=**list|full|json|name**] **export** **name** *dst*
**gostore** [--**debug**] [--**root**=**ROOT**] [--**style**=**list|full|json|name**] **check**

# DESCRIPTION
A command-line minimalist media collection manager.

# OPTIONS
--**debug**
: Show debug information.

--**root**=**ROOT**
: Path to the root of the collection.

--**style**=**list|full|json|name**
: Style for printing records' details.

# COMMANDS
**help**
: Show usage information.

**version**
: Show version information.

**import** [<flags>] **media**
: Import a new media into the collection.

**info** [<flags>] **name**
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

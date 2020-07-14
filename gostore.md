# NAME

gostore - A command-line minimalist media collection manager.

# SYNOPSIS

__gostore__ [--__verbose__] [--__debug__] [--__root__=*ROOT*] [--__pretend__] 
[--__auto__] [--__style__=*STYLE*] __help__
__gostore__ [--__verbose__] [--__debug__] [--__root__=*ROOT*] [--__pretend__] 
[--__auto__] [--__style__=*STYLE*] __version__
__gostore__ [--__verbose__] [--__debug__] [--__root__=*ROOT*] [--__pretend__] 
[--__auto__] [--__style__=*STYLE*] __import__ *media* ...
__gostore__ [--__verbose__] [--__debug__] [--__root__=*ROOT*] [--__pretend__] 
[--__auto__] [--__style__=*STYLE*] __list__ [*name* ...]
__gostore__ [--__verbose__] [--__debug__] [--__root__=*ROOT*] [--__pretend__] 
[--__auto__] [--__style__=*STYLE*] __search__ [--__sort__=*SORT*,...,*SORT*] 
*query*
__gostore__ [--__verbose__] [--__debug__] [--__root__=*ROOT*] [--__pretend__] 
[--__auto__] [--__style__=*STYLE*] __edit__ [--__multi-edit__] 
[--__import-orphans__] *name* ...
__gostore__ [--__verbose__] [--__debug__] [--__root__=*ROOT*] [--__pretend__] 
[--__auto__] [--__style__=*STYLE*] __delete__ *name* ...
__gostore__ [--__verbose__] [--__debug__] [--__root__=*ROOT*] [--__pretend__] 
[--__auto__] [--__style__=*STYLE*] __export__ *name* ... [*dst*]
__gostore__ [--__verbose__] [--__debug__] [--__root__=*ROOT*] [--__pretend__] 
[--__auto__] [--__style__=*STYLE*] __check__ [--__delete-ghosts__] 
[--__delete-orphans__] [--__import-orphans__]
__gostore__ [--__verbose__] [--__debug__] [--__root__=*ROOT*] [--__pretend__] 
[--__auto__] [--__style__=*STYLE*] __rebuild-index__

# DESCRIPTION

gostore is a command line tool aiming at providing facilities to manage one or 
more collections of media files, keeping track of their metadata and/or 
additional information the user wants to record.

# OPTIONS

--__debug__
:Show debug information.

--__root__=*ROOT*
:Path to the root of the collection.

--__pretend__
:Operations that modify the collection are simulated.

--__auto__
:Perform operations without manual interaction from the user.

--__style__=*STYLE*
:Style for printing records' details. Available styles are defined in the 
configuration file.

# COMMANDS

__help__
:Show usage information.

__version__
:Show version information.

__import__ *media* ...
:Import a new media into the collection.

__list__ [*name* ...]
:List and retrieve information about collection's records. If no pattern is 
provided, list all records of the collection.

__search__ [<flags>] *query*
:Search the collection's records matching the given query.

__edit__ [<flags>] *name* ...
:Edit an existing record from the collection using user defined's editor. If 
flag '--auto' is used, edition is skipped and nothing happens.

__delete__ *name* ...
:Delete an existing record from the collection.

__export__ *name* ... [*dst*]
:Copy a record's media file from the collection to the given destination.

__check__ [<flags>]
:Verify collection's consistency and repairs or reports found inconsistencies.

__rebuild-index__
:Deletes then rebuild the collection's index from scratch. Useful for example to
implement a new mapping strategy or if things are really going bad.

# FILES

*/etc/manpage_generate/config.yaml*
:System-wide configuration location

*/home/pir/.config/manpage_generate/config.yaml*
:Per-user configuration location

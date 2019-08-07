# NAME

gostore - A command-line minimalist media collection manager.

# SYNOPSIS

__gostore__ [--__debug__] [--__root__=*ROOT*] 
[--__style__=*name|list|full|json*] __help__
__gostore__ [--__debug__] [--__root__=*ROOT*] 
[--__style__=*name|list|full|json*] __version__
__gostore__ [--__debug__] [--__root__=*ROOT*] 
[--__style__=*name|list|full|json*] __import__ [--__auto__] [--__dry-run__] 
*media*
__gostore__ [--__debug__] [--__root__=*ROOT*] 
[--__style__=*name|list|full|json*] __info__ [--__from-file__] *name*
__gostore__ [--__debug__] [--__root__=*ROOT*] 
[--__style__=*name|list|full|json*] __list__ *query*
__gostore__ [--__debug__] [--__root__=*ROOT*] 
[--__style__=*name|list|full|json*] __edit__ *name*
__gostore__ [--__debug__] [--__root__=*ROOT*] 
[--__style__=*name|list|full|json*] __delete__ *name*
__gostore__ [--__debug__] [--__root__=*ROOT*] 
[--__style__=*name|list|full|json*] __export__ *name* *dst*
__gostore__ [--__debug__] [--__root__=*ROOT*] 
[--__style__=*name|list|full|json*] __check__

# DESCRIPTION

A command-line minimalist media collection manager.

# OPTIONS

--__debug__
:Show debug information.

--__root__=*ROOT*
:Path to the root of the collection.

--__style__=*name|list|full|json*
:Style for printing records' details.

# COMMANDS

__help__
:Show usage information.

__version__
:Show version information.

__import__ [<flags>] *media*
:Import a new media into the collection.

__info__ [<flags>] *name*
:retrieve information about any collection's record.

__list__ [*query*]
:List the the records that match the given query. If no query is provied, list 
all records

__edit__ *name*
:edit an existing record from the collection.

__delete__ *name*
:delete a record from the collection.

__export__ *name* *dst*
:export a media from the collection to the given destination.

__check__
:Verify collection's consistency.

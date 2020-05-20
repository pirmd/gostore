# Changelog

## [Unreleased]
## Added
- Add support to import multiple files at once.

## [0.3.4] - 2020.05.17
## Modified
- Google books is converted to module format for better flexibility and later
  extendability
- improve template support to define output format as well as media naming
  schemes. 
- FIX various bugs, spelling errors and extend test coverage.

## [0.3.2] - 2020.05.05
## Added
- Add new template function to only display a metadata if non empty.
## Modified
- FIX 'list' command does not search for the given pattern
- Modify way to display difference between information stored in
  collection and information stored in media file metadata.
- rework template funcmap offered to tweak the way information from the store
  is displayed on screen.
- refactor slightly the way the configuration is handled inside gostrore, store
  and UI structs and modules.

## [0.3.1] - 2020.03.07
### Added
- re-introduce ability to read metadata from files with new
  github.com/pirmd/text/diff hoping better user readability
### Modified
- get rid of old github.com/pirmd/cli dependency
- migrate to new github.com/pirmd/verify version
- migrate to new github.com/blevesearch/bleve and github.com/gabriel-vasile/mimetype
- switch from bolt to go.etcd.io/bbolt
- refactor ui module

## [0.3.0] - 2019.11.19
### Modified
- refactor gostore cli commands definition, separate core functions from ui and
  cli application definition.
- refactor gostore configuration as well as processing modules for hopefully
  better modularity.
### Removed
- disable ability to read metadata from files rather than from the collection's
  database 'until the diff algorithm is fixed)

## [0.2.0] - 2019.08.11
### Added
- add basic support to fetch metadata from google books api.
- add a processing module that cleans epub description from any html formatting
  by converting them into markdown.
- add a processing module that cleans record's name to get reasonable filenames
  (avoid spaces for example)
- generate an help file in markdown format in addition to the manpage.
### Modified
- rename `get` command to `info` and `update` command to `edit`.
- merge `list` and `search` commands.
- change `import` behavior that, by default, asks the user to manually edit
  metadata before storing them in the store. You can have this done
  automatically by using the `--auto`flag.
  I've changed the previous behaviour as the new metadata fetching feature needs
  probably some more love before being blindly trusted.
- Replace the Makefile, which was over-engineered (and probably badly done),
  by a small shell script that wraps go binary to supply version information
  and manpage generation/installation

## [0.1.0] - 2019-05-10
### Added
- basic CRUD operation to manage the collection (import / export / get / search
  / update / delete).
- basic CLI user interface with reasonable customization level of the output
  format.
- support epub files.


[modeline]: # ( vim: set fenc=utf-8 spell spl=en: )

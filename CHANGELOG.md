# Changelog
All notable changes to this project will be documented in this file.

Format is based on [Keep a Changelog] (https://keepachangelog.com/en/1.0.0/).
Versionning adheres to [Semantic Versioning] (https://semver.org/spec/v2.0.0.html)

## [Unreleased]

## [0.2.0] - 2019.08.11œ
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
- Replace the Makefile, which was over-enginneered (and probably badly done),
  by a small shell script that wraps go binary to supply version information
  and manpage generation/installation

## [0.1.0] - 2019-05-10
### Added
- basic CRUD operation to manage the collection (import / export / get / search
  / update / delete).
- basic CLI user interface with reasonable customization level of the output
  format.
- support epub files.

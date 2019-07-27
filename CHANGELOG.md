# Changelog
All notable changes to this project will be documented in this file.

Format is based on [Keep a Changelog] (https://keepachangelog.com/en/1.0.0/).
Versionning adheres to [Semantic Versioning] (https://semver.org/spec/v2.0.0.html)

## [Unreleased]
### Added
- add basic support to fetch metadata from google books api.
- generate an help file in markdown format in addition to the manpage.
- processing module that clean epub description from any html formatting
  by converting them into markdown.
### Modified
- change `import` behavior that, by default, asks the user to manually edit
  metadata before storing them in the store. You can have this done
  automatically by using the `--auto`flag.
  I change the previous behaviour as the new metadata fetching feature need
  probably some more love before being blindly trusted.
- Rename `get` command to `info`and `update` command to `edit`.
- Merge `list` and `search` commands.
- Replace the Makefile, which was over-enginneered (and probably badly done),
  by a small shell script that wraps go binary to supply version information
  and manpage generation/installation

## [0.1.0] - 2019-05-10
### Added
- basic CRUD operation to manage the collection
  (import/export/get/search/pdate/delete).
- basic CLI user interface with reasonable customization level of the output
  format.
- support epub files.

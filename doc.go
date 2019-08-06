// gostore is a command line tool aiming at providing facilities to manage one
// or more collections of media files, keeping track of their metadata and/or
// additional information.
//
// At this point of time, `gostore` is offering a way to manage a personnal ebooks
// collections but should be extendable to accommodate others kind of media files
// (music, images...).
//
// You can think of `gostore` as something close to [beets](http://beets.io/) (but
// less feature-full and mature at this time) but for books.
//
// USAGE
// An up-to-date manpage is provided with the package (use `man -l gostore.1` to
// browse it without installing) or in a [text format](./gostore.md).
//
// To get a flavor of available commmands:
// - `import`: add te given media files to the collection, extracting its
// metadata and optionally process (usually to clean/complete) them. User
// can also manually edit the metadata before saving them in the store;
// - `export`: copy requested record to the given location (usually your
// ebook-reader for eupb)
// - `list`: search the store for existing matching records. Search query is
// based on [bleve](https://blevesearch.com/) and adopt its
// [query](https://blevesearch.com/docs/Query-String-Query/) language
// - `info`: get the information known about the given record
// - `edit`: offer the user to edit information stored about the given
// record
// - `delete`: remove a record from the collection
// - `check`: verify the store's consistency (between file
// sytsems/database/index) and solve or report detected issue
//
// Different output style can be customized/chosen allowing different
// presentation and level of information feedbacked by the tool. You might want
// also to pipe output of the tool to some standard unix tool (like `less`).

package main

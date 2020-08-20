# GOSTORE

[![GoDoc](https://godoc.org/github.com/pirmd/gostore?status.svg)](https://godoc.org/github.com/pirmd/gostore)&nbsp; 
[![Go Report Card](https://goreportcard.com/badge/github.com/pirmd/gostore)](https://goreportcard.com/report/github.com/pirmd/gostore)&nbsp;

`gostore` is a command line tool aiming at providing facilities to manage one
or more collections of media files, keeping track of their metadata and/or
additional information the user wants to record.

At this point of time, `gostore` is offering a way to manage a personnal ebooks
collections but should be extendable to accommodate others kind of media files
(music, images...).

You can think of `gostore` as something close to [beets](http://beets.io/) (but
less feature-full and mature at this time) but for books.

## USAGE
An up-to-date manpage is provided with the package (use `man -l gostore.1` to
browse it without installing) or in a [text format](./gostore.md).

To get a flavor of available commands:
    - `import`: add the given media files to the collection, extracting its
      metadata and optionally process (usually to clean/complete) them. User
      can also manually edit the metadata before saving them in the store;
    - `export`: copy requested record to the given location (usually your
      ebook-reader for eupb);
    - `list`: list records from the store. It accepts wildcards pattern.
    - `search`: search the store for existing matching records. Search query
      is based on [bleve](https://blevesearch.com/) and adopt its
      [query](https://blevesearch.com/docs/Query-String-Query/) language;
    - `info`: get the information known about the given record;
    - `edit`: offer the user to edit information stored about the given
      record;
    - `delete`: remove a record from the collection;
    - `check`: verify the store's consistency (between file
      sytsems/database/index) and solve or report detected issue.

Different output style can be customized/chosen allowing different
presentation and level of information given by the tool. You might want
also to pipe output of the tool to some standard unix tool (like `less`).

## CUSTOMIZATION
Customization can be achieved by using a yaml config files ans putting it in
your usual user configuration folder (like $XDG_CONFIG_HOME/gostore/config.yaml).
An example and commented config file can be found in config.example.yaml.

## INSTALLATION
Run `go get github.com/pirmd/gostore` then `sh ./go install` from inside the
gostore repository.

Note: `make` is a small shell script on top of go binary that incorporates
version information and takes care of manpage generation and installation.
You can either go directly with go standard command line commands.

## MAIN GOALS
Beside bug hunting and improved user experience, main functions planned to be
developed (in no special order):
    - scrapers to retrieve metadata from known remote sites (like goodread);
    - offering more record's metadata processings allowing further cleaning and
      quality of collection content; 
    - allowing syncing file's embedded metatdata with cleaned and completed
      metadata stored in the collection;
    - tweak output template to issue static html description of the collection;
    - improve batch operation (add several media at a time);
    - support more bleve engine features (allowing sorting or folding or
      cleaver stemmers);
    - new media family to be supported (like mp3).

## CONTRIBUTION
If you feel like to contribute, just follow github guidelines on
[forking](https://help.github.com/articles/fork-a-repo/) then [send a pull
request](https://help.github.com/articles/creating-a-pull-request/)


[modeline]: # ( vim: set fenc=utf-8 spell spl=en: )

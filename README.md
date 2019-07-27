# GOSTORE

[![GoDoc](https://godoc.org/github.com/pirmd/gostore?status.svg)](https://godoc.org/github.com/pirmd/gostore)&nbsp; 
[![Go Report Card](https://goreportcard.com/badge/github.com/pirmd/gostore)](https://goreportcard.com/report/github.com/pirmd/gostore)&nbsp;

`gostore` is a command line tool aiming at providing facilities to manage one
or more collections of media files, keeping track of their metadata and/or
additional information.

At this point of time, `gostore` is offering a way to manage a personnal ebooks
collections but should be extendable to accomodate others kind of media files
(music, images...).

You can think of `gostore` as something close to [beets](http://beets.io/) (but
less feature-full and mature at this time) but for books.


## USAGE
An up-to-date manpage is provided with the package (use `man -l gostore.1` to
browse it without installing) or in a [text format](./gostore.md).

To get a flavor of available commmands:
    - `import`: add te given media files to the collection, extracting its
      metadata and optionally process (usually to clean/complete) them. User
      can also manualy edit the metadata before saving them in the store;
    - `export`: copy requested record to the given location (usualy your
      ebook-reader for eupb)
    - `list`: search the store for existing matching records. Search query is
      based on [bleve](https://blevesearch.com/) and adopt its
      [query](https://blevesearch.com/docs/Query-String-Query/) language
    - `info`: get the information known about the given record
    - `edit`: offer the user to edit information stored about the given
      record
    - `delete`: remove a record from the collection
    - `check`: verify the store's consistency (between file
      sytsems/database/index) adn solve or report detected issue

Different output style can be customized/chosen allowing different
presentation and level of information feedbacked by the tool. You might want
also to pipe output of the tool to some standard unix tool (like `less`).

## CUSTOMIZATION
I don't feel the need to have a config file (yet), you'll have to modify the
code for that in the `config.go`, which is hopefully commented enough to have
you play with customizations (even for someone not familiar with golang).
Once done, run `make install` and you're done.

## INSTALLATION
Everything should work fine using go standard commands (`build`, `get`,
`install`...). For simplicity, you can just run `sh ./go install` if you prefer
(supplied o`is a small shell script on top of go binary that incorporates
version information and takes care of manpage generation and installation).

## MAIN GOALS
Beside bug hunting and improved user experience, main functions planned to be
developped (in no special order):
    - scrapers to retrieve metadata from known remote sites (like goodread)
    - offering more record's metadata processings alowing further cleaning and
      quality of collection content 
    - allowing syncing file's embedded metatdata with cleaned and completed
      metadata stored in the collection
    - tweak output template to issue static html description of the collection
    - improve batch operation (add several media at a time)
    - support more bleve engine features (allowing sorting or folding or
      cleaver stemmers)
    - new media familly to be supported (like mp3)

## CONTRIBUTION
If you feel like to contribute, just follow github guidelines on
[forking](https://help.github.com/articles/fork-a-repo/) then [send a pull
request](https://help.github.com/articles/creating-a-pull-request/)

[modeline]: # ( vim: set fenc=utf-8 spell spl=en: )

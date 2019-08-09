package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/pirmd/cli/app"

	"github.com/pirmd/gostore/media"
	"github.com/pirmd/gostore/processing"
	"github.com/pirmd/gostore/store"
)

var (
	gostore = app.New("gostore", "A command-line minimalist media collection manager.")
)

func init() {
	gostore.NewBoolFlagToVar(&cfg.LogDebug, "debug", "Show debug information.")
	gostore.NewStringFlagToVar(&cfg.StoreRoot, "root", "Path to the root of the collection.")
	gostore.NewEnumFlagToVar(&cfg.UIFormatStyle, "style", "Style for printing records' details.", getUIFormatStyles(cfg.UIFormatters))

	importCmd := gostore.NewCommand("import", "Import a new media into the collection.")
	importAuto := importCmd.NewBoolFlag("auto", "Automatically fecth metadata before importing them in the collection.")
	importDryRun := importCmd.NewBoolFlag("dry-run", "Simulate importing a new media in th ecollection (actually retrieveing metadata without inserting them into the store).")
	importMedia := importCmd.NewStringArg("media", "Media to import into the collection.")
	importCmd.Execute = func() error {
		configure()

		f, err := os.Open(*importMedia)
		if err != nil {
			return fmt.Errorf("importing '%s' failed: %s", *importMedia, err)
		}
		defer f.Close()

		mdataFromFile, err := media.GetMetadata(f)
		if err != nil {
			return fmt.Errorf("importing '%s' failed: %s", *importMedia, err)
		}

		mdataFetched, err := media.FetchMetadata(mdataFromFile)
		if err != nil && err != media.ErrNoMetadataFound {
			return fmt.Errorf("importing '%s' failed: %s", *importMedia, err)
		}

		var mdata media.Metadata
		if !*importAuto {
			left, _, err := Merge(mdataFromFile, mdataFetched)
			if err != nil {
				return fmt.Errorf("importing '%s' failed: %s", *importMedia, err)
			}
			mdata = left.(map[string]interface{})
		} else {
			mdata = mdataFetched
			fmt.Printf("Merging metadata read from file and fetched:\n")
			PrettyDiff(mdataFromFile, mdata)
		}
		r := store.NewRecord(*importMedia, mdata)

		if err := processing.ProcessRecord(r); err != nil {
			return fmt.Errorf("importing '%s' failed: %s", *importMedia, err)
		}

		if *importDryRun {
			PrettyPrint(r.Fields())
			return nil
		}

		s, err := openStore()
		if err != nil {
			return fmt.Errorf("importing '%s' failed: %s", *importMedia, err)
		}
		defer s.Close()

		if err := s.Create(r, f); err != nil {
			return fmt.Errorf("importing '%s' failed: %s", *importMedia, err)
		}

		PrettyPrint(r.Fields())
		return nil
	}

	infoCmd := gostore.NewCommand("info", "retrieve information about any collection's record.")
	infoFromFile := infoCmd.NewBoolFlag("from-file", "retrieve information from media file rather than from database.")
	infoKey := infoCmd.NewStringArg("name", "name of the record to get information about.")
	infoCmd.Execute = func() error {
		configure()

		s, err := openStore()
		if err != nil {
			return fmt.Errorf("retrieving information about '%s' failed: %s", *infoKey, err)
		}
		defer s.Close()

		r, err := s.Read(*infoKey)
		if err != nil {
			return fmt.Errorf("retrieving information about '%s' failed: %s", *infoKey, err)
		}

		if *infoFromFile {
			f, err := s.OpenRecord(*infoKey)
			if err != nil {
				return fmt.Errorf("retrieving information about '%s' failed: %s", *infoKey, err)
			}
			defer f.Close()

			mdata, err := media.GetMetadata(f)
			if err != nil {
				return fmt.Errorf("retrieving information about '%s' failed: %s", *infoKey, err)
			}

			PrettyDiff(r.OrigValue(), mdata)
			return nil
		}

		PrettyPrint(r.Fields())
		return nil
	}

	searchCmd := gostore.NewCommand("list", "List the the records that match the given query. If no query is provied, list all records")
	searchQuery := "*"
	searchCmd.NewStringArgToVar(&searchQuery, "query", "Query to look for.")
	searchCmd.CanRunWithoutArg = true
	searchCmd.Execute = func() error {
		configure()

		s, err := openStore()
		if err != nil {
			return fmt.Errorf("searching for '%s' failed: %s", searchQuery, err)
		}
		defer s.Close()

		r, err := s.Search(searchQuery)
		if err != nil {
			return fmt.Errorf("searching for '%s' failed: %s", searchQuery, err)
		}

		PrettyPrint(r.Fields()...)
		return nil
	}

	editCmd := gostore.NewCommand("edit", "edit an existing record from the collection.")
	editKey := editCmd.NewStringArg("name", "name of the record to edit.")
	editCmd.Execute = func() error {
		configure()

		s, err := openStore()
		if err != nil {
			return fmt.Errorf("editing %s failed: %s", *editKey, err)
		}
		defer s.Close()

		r, err := s.Read(*editKey)
		if err != nil {
			return fmt.Errorf("editing %s failed: %s", *editKey, err)
		}

		mdata, err := Edit(r.OrigValue())
		if err != nil {
			return fmt.Errorf("editing %s failed: %s", *editKey, err)
		}
		r.ReplaceValues(mdata.(map[string]interface{}))

		if err := processing.ProcessRecord(r); err != nil {
			return fmt.Errorf("editing %s failed: %s", *editKey, err)
		}

		if err := s.Update(*editKey, r); err != nil {
			return fmt.Errorf("editing %s failed: %s", *editKey, err)
		}

		PrettyPrint(r.Fields())
		return nil
	}

	delCmd := gostore.NewCommand("delete", "delete a record from the collection.")
	delKey := delCmd.NewStringArg("name", "name of the record to delete.")
	delCmd.Execute = func() error {
		configure()

		s, err := openStore()
		if err != nil {
			return fmt.Errorf("deleting '%s' failed: %s", *delKey, err)
		}
		defer s.Close()

		if err := s.Delete(*delKey); err != nil {
			return fmt.Errorf("deleting '%s' failed: %s", *delKey, err)
		}

		return nil
	}

	exportCmd := gostore.NewCommand("export", "export a media from the collection to the given destination.")
	exportKey := exportCmd.NewStringArg("name", "name of the record to export.")
	exportDst := exportCmd.NewStringArg("dst", "destination where the record need to be exported.")
	exportCmd.Execute = func() error {
		configure()

		dst := filepath.Join(*exportDst, *exportKey)

		s, err := openStore()
		if err != nil {
			return fmt.Errorf("exporting '%s' to '%s' failed: %s", *exportKey, *exportDst, err)
		}
		defer s.Close()

		r, err := s.OpenRecord(*exportKey)
		if err != nil {
			return fmt.Errorf("exporting '%s' to '%s' failed: %s", *exportKey, *exportDst, err)
		}
		defer r.Close()

		if err := os.MkdirAll(filepath.Dir(dst), 0777); err != nil {
			return fmt.Errorf("exporting '%s' to '%s' failed: %s", *exportKey, *exportDst, err)
		}

		w, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE, 0666)
		if err != nil {
			return fmt.Errorf("exporting '%s' to '%s' failed: %s", *exportKey, *exportDst, err)
		}
		defer w.Close()

		if _, err := io.Copy(w, r); err != nil {
			_ = os.Remove(dst)
			return fmt.Errorf("exporting '%s' to '%s' failed: %s", *exportKey, *exportDst, err)
		}

		return nil
	}

	checkCmd := gostore.NewCommand("check", "Verify collection's consistency.")
	checkCmd.Execute = func() error {
		configure()

		s, err := openStore()
		if err != nil {
			return fmt.Errorf("checking collection failed: %s", err)
		}
		defer s.Close()

		orphans, err := s.CheckAndRepair()
		if err != nil {
			return fmt.Errorf("checking collection failed: %s", err)
		}

		if len(orphans) > 0 {
			fmt.Printf("Found orphans files in the collection:\n%s\n", strings.Join(orphans, "\n"))
		}
		return nil
	}
}

func openStore() (*store.Store, error) {
	return store.Open(
		cfg.StoreRoot,
		store.UsingLogger(debugger),
		store.UsingTypeField(media.TypeField),
	)
}

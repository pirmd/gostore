package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/kballard/go-shellquote"
)

type editorConfig struct {
	// EditorCmd is the stanza to invoke the editor.
	// The command is specified using fmt.Printf template format, expecting one
	// "%s" argument standing for the filename to edit.
	EditorCmd string

	// MergerCmd is the stanza to invoke the merger.
	// The command is specified using fmt.Printf template format, expecting two
	// "%s" arguments standing for the filenames to merge.
	MergerCmd string

	// Idiom is the format used to edit or merge data using the editor.
	// Supported idioms: json
	// Default to json.
	Idiom string
}

func newEditorConfig() *editorConfig {
	return &editorConfig{
		Idiom: "json",
	}
}

type editor struct {
	editorCmd string
	mergerCmd string

	marshal   func(interface{}) ([]byte, error)
	unmarshal func([]byte, interface{}) error
	// filepattern is a pattern of the temporary file used during edition operation.
	// The pattern follows ioutil.TempFile pattern rule.
	// It is usually helpful to select a meaningful extension to benefit of
	// the editor syntax functions.
	filepattern string
}

func newEditor() *editor {
	return &editor{
		marshal:     func(v interface{}) ([]byte, error) { return json.MarshalIndent(v, "", "  ") },
		unmarshal:   json.Unmarshal,
		filepattern: "*.json",
	}
}

func newEditorFromConfig(cfg *editorConfig) (*editor, error) {
	ed := newEditor()
	ed.editorCmd = cfg.EditorCmd
	ed.mergerCmd = cfg.MergerCmd

	switch cfg.Idiom {
	case "", "json": // default
	default:
		return nil, fmt.Errorf("%s is an unknown edition idiom (support: json)", cfg.Idiom)
	}

	return ed, nil
}

func (ed *editor) Edit(v, edited interface{}) error {
	tmpfile, err := ed.data2file(v)
	if err != nil {
		return err
	}

	if err := run(ed.editorCmd, tmpfile); err != nil {
		return err
	}

	if err := ed.file2data(tmpfile, &edited); err != nil {
		return err
	}

	return nil
}

func (ed *editor) Merge(l, r, merged interface{}) error {
	tmpfileL, err := ed.data2file(l)
	if err != nil {
		return err
	}

	tmpfileR, err := ed.data2file(r)
	if err != nil {
		return err
	}

	if err := run(ed.mergerCmd, tmpfileL, tmpfileR); err != nil {
		return err
	}

	if err := ed.file2data(tmpfileL, &merged); err != nil {
		return err
	}

	return nil
}

// data2file copy the content of data to a temporary text file
func (ed *editor) data2file(v interface{}) (string, error) {
	data, err := ed.marshal(v)
	if err != nil {
		return "", err
	}

	tmpfile, err := ioutil.TempFile("", ed.filepattern)
	if err != nil {
		return "", err
	}
	defer tmpfile.Close()

	n, err := tmpfile.Write(data)
	if err != nil {
		os.Remove(tmpfile.Name())
		return "", err
	}

	if n < len(data) {
		os.Remove(tmpfile.Name())
		return "", io.ErrShortWrite
	}

	return tmpfile.Name(), nil
}

// file2data reads back the content of a temp file and deletes it whatever
// happens. file2data strips comments line.
func (ed *editor) file2data(name string, v interface{}) error {
	defer func() { os.Remove(name) }()

	data, err := ioutil.ReadFile(name)
	if err != nil {
		return err
	}

	if err := ed.unmarshal(data, &v); err != nil {
		return err
	}

	return nil
}

// run executes a command line provided in fmt.Printf format. If the command
// lie is empty, run will not fail and do nothing.
func run(cmdfmt string, args ...interface{}) error {
	if cmdfmt == "" {
		return nil
	}

	cmdline, err := shellquote.Split(fmt.Sprintf(cmdfmt, args...))
	if err != nil {
		return err
	}

	cmd := exec.Command(cmdline[0], cmdline[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

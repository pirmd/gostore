package cli

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/kballard/go-shellquote"
)

// editAsJSON fires-up an editor to modify the provided interface using its JSON
// form.
func editAsJSON(m map[string]interface{}, cmdEditor string) (map[string]interface{}, error) {
	j, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return nil, err
	}

	buf, err := edit(j, cmdEditor)
	if err != nil {
		return nil, err
	}

	var edited map[string]interface{}
	err = json.Unmarshal(buf, &edited)
	if err != nil {
		return nil, err
	}

	return edited, nil
}

// mergeAsJSON fires-up an editor to merge the provided interfaces using its
// JSON form.
func mergeAsJSON(left, right map[string]interface{}, cmdMerger string) (map[string]interface{}, error) {
	l, err := json.MarshalIndent(left, "", "  ")
	if err != nil {
		return nil, err
	}

	r, err := json.MarshalIndent(right, "", "  ")
	if err != nil {
		return nil, err
	}

	bufL, _, err := merge(l, r, cmdMerger)
	if err != nil {
		return nil, err
	}

	var merged map[string]interface{}
	err = json.Unmarshal(bufL, &merged)
	if err != nil {
		return nil, err
	}

	return merged, nil
}

// edit spans an editor to modify the input text and feedbacks the result.
func edit(data []byte, cmdEditor string) ([]byte, error) {
	if len(cmdEditor) == 0 {
		return data, nil
	}

	tmpfile, err := data2file(data)
	if err != nil {
		return nil, err
	}

	cmdArgs, err := parseCmd(cmdEditor, tmpfile)
	if err != nil {
		return nil, fmt.Errorf("cannot parse editor command line '%s': %s", cmdEditor, err)
	}

	ed := exec.Command(cmdArgs[0], cmdArgs[1:]...)
	ed.Stdout = os.Stdout
	ed.Stdin = os.Stdin
	ed.Stderr = os.Stderr
	err = ed.Run()
	if err != nil {
		return nil, err
	}

	body, err := file2data(tmpfile)
	if err != nil {
		return nil, err
	}

	return body, nil
}

// merge spans an editor to merge the input texts and feedbacks the result.
func merge(left, right []byte, cmdMerger string) ([]byte, []byte, error) {
	if len(cmdMerger) == 0 {
		return left, right, nil
	}

	tmpfileL, err := data2file(left)
	if err != nil {
		return nil, nil, err
	}

	tmpfileR, err := data2file(right)
	if err != nil {
		return nil, nil, err
	}

	cmdArgs, err := parseCmd(cmdMerger, tmpfileL, tmpfileR)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot parse merger command line '%s': %s", cmdMerger, err)
	}

	ed := exec.Command(cmdArgs[0], cmdArgs[1:]...)
	ed.Stdout = os.Stdout
	ed.Stdin = os.Stdin
	ed.Stderr = os.Stderr
	err = ed.Run()
	if err != nil {
		return nil, nil, err
	}

	bodyL, err := file2data(tmpfileL)
	if err != nil {
		return nil, nil, err
	}

	bodyR, err := file2data(tmpfileR)
	if err != nil {
		return nil, nil, err
	}

	return bodyL, bodyR, nil
}

// data2file copy the content of data to a temporary text file
func data2file(data []byte) (string, error) {
	tmpfile, err := ioutil.TempFile("", "")
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
func file2data(name string) ([]byte, error) {
	defer func() { os.Remove(name) }()

	f, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	data := []byte{}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Bytes()
		if !bytes.HasPrefix(line, []byte{'#'}) {
			data = append(data, line...)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return data, nil
}

// parseCmd parses a command-line
func parseCmd(cmdline string, args ...interface{}) ([]string, error) {
	c := fmt.Sprintf(cmdline, args...)

	cmd, err := shellquote.Split(c)
	if err != nil {
		return nil, err
	}

	return cmd, nil
}

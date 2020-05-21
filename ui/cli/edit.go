package cli

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
)

// edit spans an editor to modify the input text and feedbacks the result.
func edit(data []byte, cmdEditor []string) ([]byte, error) {
	if len(cmdEditor) == 0 {
		return data, nil
	}

	tmpfile, err := data2file(data)
	if err != nil {
		return nil, err
	}

	cmdArgs := append(cmdEditor, tmpfile)
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

// editAsJSON fires-up an editor to modify the provided interface using its JSON
// form.
func editAsJSON(v interface{}, cmdEditor []string) (interface{}, error) {
	j, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return nil, err
	}

	buf, err := edit(j, cmdEditor)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(buf, &v)
	if err != nil {
		return nil, err
	}

	//I'm not that sure that v needs to be returned as for most of the cases the
	//Unmarshal directive should have already propagated the mods. It happens
	//that it is not working at least for map (that should nee to be
	//reallocated), so result is also returned to the user.
	//
	//TODO(pirmd): it is probably not the right way to do, try harder to find a
	//correct approach
	return v, nil
}

// merge spans an editor to merge the input texts and feedbacks the result.
func merge(left, right []byte, cmdMerger []string) ([]byte, []byte, error) {
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

	cmdArgs := append(cmdMerger, tmpfileL, tmpfileR)
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

// mergeAsJSON fires-up an editor to merge the provided interfaces using its
// JSON form.
func mergeAsJSON(left, right interface{}, cmdMerger []string) (interface{}, interface{}, error) {
	l, err := json.MarshalIndent(left, "", "  ")
	if err != nil {
		return nil, nil, err
	}

	r, err := json.MarshalIndent(right, "", "  ")
	if err != nil {
		return nil, nil, err
	}

	bufL, bufR, err := merge(l, r, cmdMerger)
	if err != nil {
		return nil, nil, err
	}

	err = json.Unmarshal(bufL, &left)
	if err != nil {
		return nil, nil, err
	}

	err = json.Unmarshal(bufR, &right)
	if err != nil {
		return nil, nil, err
	}

	//TODO(pirmd): it is probably not the right way to do, try harder to find a
	//correct approach
	return left, right, nil
}

// data2file copy the content of data to a temporary text file, perfect for
// editing purpose
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

// file2data reads back the content of a temp file and delete it whatever happen
func file2data(name string) ([]byte, error) {
	defer func() { os.Remove(name) }()

	body, err := ioutil.ReadFile(name)
	if err != nil {
		return nil, err
	}

	return body, nil
}

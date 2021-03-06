package media

import (
	"testing"
)

type mockHandler struct {
	typ, mtyp string
}

func (mh *mockHandler) Type() string {
	return mh.typ
}

func (mh *mockHandler) Mimetype() string {
	return mh.mtyp
}

func (mh *mockHandler) ReadMetadata(f File) (Metadata, error) {
	return map[string]interface{}{}, nil
}

func (mh *mockHandler) FetchMetadata(mdata Metadata) ([]Metadata, error) {
	mdata["Fetcher"] = mh.typ
	return []Metadata{mdata}, nil
}

func (mh *mockHandler) CheckMetadata(mdata Metadata) int {
	return 100
}

func (mh *mockHandler) IDCard(mdata Metadata) (exact [][2]string, similar [][2]string) {
	return
}

func TestForType(t *testing.T) {
	testHandlers := Handlers{
		&mockHandler{"test1", "mock/test1"},
	}

	testCases := []struct {
		in   string
		want string
	}{
		{"test1", "test1"},
		{"unknown", "default"},
	}

	if _, err := testHandlers.ForType("unknown"); err != ErrUnknownMediaType {
		t.Errorf("Retrieve handler for unknown type should fail with ErrUnknownMediatype but did not (err = %s)", err)
	}

	testHandlers = append(testHandlers, &mockHandler{"default", DefaultMimetype})
	for _, tc := range testCases {
		mh, err := testHandlers.ForType(tc.in)
		if err != nil {
			t.Errorf("Fail to retrieve handler for %s: %s", tc.in, err)
		}
		if mh.Type() != tc.want {
			t.Errorf("Fail to retrive handler for %s\nWant: %s\nGot : %s", tc.in, tc.want, mh.Type())
		}
	}

}

func TestForMimeType(t *testing.T) {
	testHandlers := Handlers{
		&mockHandler{"test1", "mock/test1"},
	}

	testCases := []struct {
		in   string
		want string
	}{
		{"mock/test1", "test1"},
		{"unknown", "default"},
	}

	if _, err := testHandlers.ForMimetype("unknown"); err != ErrUnknownMediaType {
		t.Errorf("Retrieve handler for unknown mimetype should fail with ErrUnknownMediatype but did not (err = %s)", err)
	}

	testHandlers = append(testHandlers, &mockHandler{"default", DefaultMimetype})
	for _, tc := range testCases {
		mh, err := testHandlers.ForMimetype(tc.in)
		if err != nil {
			t.Errorf("Fail to retrieve handler for %s: %s", tc.in, err)
		}
		if mh.Type() != tc.want {
			t.Errorf("Fail to retrive handler for %s\nWant: %s\nGot : %s", tc.in, tc.want, mh.Type())
		}
	}

}

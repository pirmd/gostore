package hasher

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash"
	"io"

	"github.com/pirmd/gostore/module"
	"github.com/pirmd/gostore/store"
)

const (
	moduleName = "hasher"

	// HashField is the name of the Record's field where to store hash value.
	HashField = "SourceHash"
)

var (
	_ module.Module = (*hasher)(nil) // Makes sure that we implement module.Module interface.
)

type config struct {
	// HashMethod specifies the hash algorithm to use. Available methods are:
	// Available methods are: md5, sha1, sha256. Default to md5
	HashMethod string
}

func newConfig() module.Factory {
	return &config{
		HashMethod: "md5",
	}
}

func (cfg *config) NewModule(env *module.Environment) (module.Module, error) {
	return newHasher(cfg, env)
}

// hasher is a gostore's module that computes a checksum-based signature for
// the record's file and make sure it does not already exist in the store.
type hasher struct {
	*module.Environment
	hash hash.Hash
}

func newHasher(cfg *config, env *module.Environment) (*hasher, error) {
	h := &hasher{
		Environment: env,
	}

	switch cfg.HashMethod {
	case "md5":
		h.hash = md5.New()
	case "sha1":
		h.hash = sha1.New()
	case "sha256":
		h.hash = sha256.New()
	default:
		return nil, fmt.Errorf("module '%s': unknown hash method '%s'. Select md5, sha1 or sha256", moduleName, cfg.HashMethod)
	}

	return h, nil
}

// Process compute a checksum-based uuid for given record and returns an error
// if it already exists in the store.
func (h *hasher) Process(r *store.Record) error {
	if r.File() == nil {
		h.Logger.Printf("Module '%s': no record's file available for %s", moduleName, r.Key())
		return nil
	}

	h.hash.Reset()
	if _, err := io.Copy(h.hash, r.File()); err != nil {
		return fmt.Errorf("module '%s': fail to compute checksum for '%s': %v", moduleName, r.Key(), err)
	}
	checksum := hex.EncodeToString(h.hash.Sum(nil))
	r.Set(HashField, checksum)
	h.Logger.Printf("Module '%s': record hash is: %v", moduleName, checksum)

	matches, err := h.Store.SearchFields(-1, HashField, checksum)
	if err != nil {
		return fmt.Errorf("module '%s': fail to look for duplicate: %v", moduleName, err)
	}
	if len(matches) > 0 {
		return fmt.Errorf("module '%s': possible duplicate(s) of record (%v) found in the database", moduleName, matches)
	}

	return nil
}

func init() {
	module.Register(moduleName, newConfig)
}

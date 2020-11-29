// Package hasher computes a checksum-based signature for the record's file and
// make sure it does not already exist in the store.
package hasher

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"log"

	"github.com/pirmd/gostore/modules"
	"github.com/pirmd/gostore/store"
)

const (
	moduleName = "hasher"

	// HashField is the name of the Record's field where to store hash value.
	HashField = "SourceHash"
)

var (
	_ modules.Module = (*hasher)(nil) // Makes sure that we implement modules.Module interface.
)

// Config defines the different module's options.
type Config struct {
	// HashMethod specifies the hash algorithm to use. Available methods are:
	// Available methods are: md5, sha1, sha256. Default to md5
	HashMethod string
}

func newConfig() *Config {
	return &Config{
		HashMethod: "md5",
	}
}

type hasher struct {
	log   *log.Logger
	store *store.Store
	hash  hash.Hash
}

func newHasher(cfg *Config, logger *log.Logger, store *store.Store) (modules.Module, error) {
	h := &hasher{
		log:   logger,
		store: store,
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

// ProcessRecord compute a checksum-based uuid for given record and returns an
// error if it already exists in the store.
func (h *hasher) ProcessRecord(r *store.Record) error {
	if r.File() == nil {
		h.log.Printf("Module '%s': no record's file available for %s", moduleName, r.Key())
		return nil
	}

	h.hash.Reset()
	if _, err := io.Copy(h.hash, r.File()); err != nil {
		return fmt.Errorf("module '%s': fail to compute checksum for '%s': %v", moduleName, r.Key(), err)
	}
	checksum := hex.EncodeToString(h.hash.Sum(nil))
	r.Set(HashField, checksum)
	h.log.Printf("Module '%s': record hash is: %v", moduleName, checksum)

	matches, err := h.store.SearchFields(-1, HashField, checksum)
	if err != nil {
		return fmt.Errorf("module '%s': fail to look for duplicate: %v", moduleName, err)
	}
	if len(matches) > 0 {
		return fmt.Errorf("module '%s': possible duplicate(s) of record (%v) found in the database", moduleName, matches)
	}

	return nil
}

// NewFromRawConfig creates a new module from a raw configuration.
func NewFromRawConfig(rawcfg modules.Unmarshaler, env *modules.Environment) (modules.Module, error) {
	env.Logger.Printf("Module '%s': new module with config '%v'", moduleName, rawcfg)
	cfg := newConfig()

	if err := rawcfg.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("module '%s': bad configuration: %v", moduleName, err)
	}

	return newHasher(cfg, env.Logger, env.Store)
}

func init() {
	modules.Register(moduleName, NewFromRawConfig)
}

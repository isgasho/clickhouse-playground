package models

import (
	"crypto/sha256"
	"database/sql/driver"
	"encoding/hex"
	"fmt"
)

type Hash [sha256.Size]byte

func (h *Hash) Scan(src interface{}) error {
	// TODO: come up with error string
	uints, ok := src.([]uint8)
	if !ok {
		return fmt.Errorf("fuck off")
	}

	copy(h[:], uints)
	return nil
}

// TODO: fix pointer?
func (h Hash) Value() (driver.Value, error) {
	// TODO: shorten
	arr := [sha256.Size]byte(h)
	return arr[:], nil
}

func (h *Hash) Hex() string {
	return hex.EncodeToString(h[:])
}

type Query struct {
	Str  string `gorm:"UNIQUE"` // TODO: do i want to? (unique)
	Hash Hash   `gorm:"UNIQUE"` // TODO: do i want to? (unique)
}

func NewQuery(str string) *Query {
	return &Query{Str: str, Hash: sha256.Sum256([]byte(str))}
}

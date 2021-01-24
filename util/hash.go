package util

import (
	"bytes"
	"crypto/sha1"
	"fmt"
)

// HashBuffer exported
func HashBuffer(src *bytes.Buffer) string {
	h := sha1.New()
	h.Write(src.Bytes())
	bs := h.Sum(nil)
	return fmt.Sprintf("%x", bs)
}

// HashString exported
func HashString(src string) string {
	h := sha1.New()
	h.Write([]byte(src))
	bs := h.Sum(nil)
	return fmt.Sprintf("%x", bs)
}

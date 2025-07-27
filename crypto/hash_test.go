package crypto

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	text      = "hello, world!\n"
	md5Digest = "910c8bc73110b0cd1bc5d2bcae782511"
)

func TestMd5(t *testing.T) {
	actual := fmt.Sprintf("%x", Md5([]byte(text)))
	assert.Equal(t, md5Digest, actual)
}

func TestMd5Hex(t *testing.T) {
	actual := Md5Hex([]byte(text))
	assert.Equal(t, md5Digest, actual)
}

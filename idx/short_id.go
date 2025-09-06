package idx

import (
	idxv1 "github.com/byteflowing/proto/gen/go/idx/v1"
	"github.com/sqids/sqids-go"
)

// ShortIDGenerator
// 文档：https://sqids.org/go
type ShortIDGenerator struct {
	cli *sqids.Sqids
}

type ShotIDGeneratorOpts struct {
	Alphabet  string
	MinLength uint8
	Blocklist []string
}

func NewShortIdGenerator(opts *idxv1.ShortIdConfig) (generator *ShortIDGenerator, err error) {
	cli, err := sqids.New(sqids.Options{
		Alphabet:  opts.Alphabet,
		MinLength: uint8(opts.MinLength),
		Blocklist: opts.BlockList,
	})
	if err != nil {
		return nil, err
	}
	return &ShortIDGenerator{cli: cli}, nil
}

func (s *ShortIDGenerator) Encode(numbers []uint64) (id string, err error) {
	return s.cli.Encode(numbers)
}

func (s *ShortIDGenerator) Decode(id string) []uint64 {
	return s.cli.Decode(id)
}

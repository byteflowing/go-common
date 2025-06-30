package idx

import "github.com/sqids/sqids-go"

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

func NewShortIdGenerator(opts *ShotIDGeneratorOpts) (generator *ShortIDGenerator, err error) {
	cli, err := sqids.New(sqids.Options{
		Alphabet:  opts.Alphabet,
		MinLength: opts.MinLength,
		Blocklist: opts.Blocklist,
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

package idx

import (
	"time"

	"github.com/sony/sonyflake"
)

// GlobalIDGenerator 雪花算法的sony实现
// 文档：https://github.com/sony/sonyflake
type GlobalIDGenerator struct {
	generator *sonyflake.Sonyflake
}

type GlobalIDOpts struct {
	// 开始时间建议设置为项目开始时间
	// 不要设置未来时间，会导致时间回拨等问题
	// 格式："20060102150405"
	// 时间一旦确定请勿随意更改，建议硬编码到程序代码中
	StartTime string
	// [可选] 默认会使用本机ip的低16位
	// 如果多实例(goroutine)执行建议提供以免id重复
	// 可以使用自增id等等方法
	MachineID func() (uint16, error)
	// [可选] 用于检测machine id是否重复
	CheckMachineID func(uint16) bool
}

func NewGlobalIDGenerator(opts *GlobalIDOpts) (sf *GlobalIDGenerator, err error) {
	t, err := time.Parse("20060102150405", opts.StartTime)
	if err != nil {
		return nil, err
	}
	st := sonyflake.Settings{}
	st.StartTime = t
	if opts.MachineID != nil {
		st.MachineID = opts.MachineID
	}
	if opts.CheckMachineID != nil {
		st.CheckMachineID = opts.CheckMachineID
	}
	flake, err := sonyflake.New(st)
	if err != nil {
		return nil, err
	}
	return &GlobalIDGenerator{generator: flake}, nil
}

// NextID 生成全局ID
// 每 10 毫秒最多可以生成 256 个 ID
func (s *GlobalIDGenerator) NextID() (id uint64, err error) {
	return s.generator.NextID()
}

// Decompose 解析出id的组成部分
// id：原始的完整 ID。
// msb：最高有效位，一般为 0。
// time：时间戳部分，表示从 StartTime 开始到 ID 生成的时间。
// sequence：序列号，用于防止同一时间戳内的 ID 冲突。
// machine-id：机器 ID，标识生成该 ID 的机器。
func (s *GlobalIDGenerator) Decompose(id uint64) map[string]uint64 {
	return sonyflake.Decompose(id)
}

package mail

import (
	"context"
	"crypto/tls"
	"errors"
	"sync"
	"time"

	"github.com/wneessen/go-mail"
)

type ContentType string

const (
	ContentTypeHTML ContentType = "text/html"
	ContentTypeText ContentType = "text/plain"
)

func (m ContentType) toRealType() mail.ContentType {
	switch m {
	case ContentTypeHTML:
		return mail.TypeTextHTML
	case ContentTypeText:
		return mail.TypeTextPlain
	default:
		return mail.TypeTextHTML
	}
}

type Address struct {
	Addr string // 必选 邮件地址
	Name string // 可选 地址对应的名称
}

type Mail struct {
	From        *Address    // 发件人邮箱地址（通常是 SMTP 账号对应的邮箱，例如 "no-reply@example.com"）
	To          []*Address  // 收件人邮箱列表（主要接收者）
	Cc          []*Address  // 抄送邮箱列表（所有收件人都能看到被抄送的地址）
	Bcc         []*Address  // 密送邮箱列表（其他收件人看不到被密送的地址）
	Subject     string      // 邮件主题（标题）
	ContentType ContentType // 邮件正文的纯文本内容（不带格式，兼容性最好）
	Content     string      // 邮件正文的 HTML 格式内容（可以带样式、图片、超链接等）
	Attachments []string    // 附件文件路径列表（本地文件路径，支持多个附件）
}

type SMTP struct {
	cli         *mail.Client
	isConnected bool
	mux         sync.RWMutex
}

type SMTPOpts struct {
	Host     string // 必选，smtp server地址
	Port     int    // 必选，端口
	Username string // 必选，用户名
	Password string // 必选，密码
	SkipTLS  bool   // 可选，仅测试使用
	Timeout  int    // 超时时间,单位s
}

func NewSMTP(opts *SMTPOpts) *SMTP {
	var clientOpts []mail.Option
	clientOpts = append(
		clientOpts,
		mail.WithUsername(opts.Username),
		mail.WithPassword(opts.Password),
		mail.WithPort(opts.Port),
		mail.WithTLSPolicy(mail.TLSOpportunistic),
		mail.WithSMTPAuth(mail.SMTPAuthAutoDiscover),
	)
	if opts.Timeout > 0 {
		clientOpts = append(clientOpts, mail.WithTimeout(time.Duration(opts.Timeout)*time.Second))
	}
	if opts.SkipTLS {
		clientOpts = append(clientOpts, mail.WithTLSConfig(&tls.Config{InsecureSkipVerify: true}))
	}
	cli, err := mail.NewClient(opts.Host, clientOpts...)
	if err != nil {
		panic(err)
	}
	return &SMTP{cli: cli}
}

// DialAndSend 每次都会建立连接发送邮件关闭连接
// 会根据初始化时提供的限流参数进行限流，如果被限流会阻塞，可以通过ctx传入超时控制
func (s *SMTP) DialAndSend(ctx context.Context, mails ...*Mail) error {
	if len(mails) == 0 {
		return errors.New("mail list is empty")
	}
	messages, err := s.convert2Messages(mails...)
	if err != nil {
		return err
	}
	return s.cli.DialAndSendWithContext(ctx, messages...)
}

// Dial 与smtp server建立连接
func (s *SMTP) Dial(ctx context.Context) error {
	s.mux.Lock()
	defer s.mux.Unlock()
	if s.isConnected {
		return nil
	}
	if err := s.cli.DialWithContext(ctx); err != nil {
		return err
	}
	s.isConnected = true
	return nil
}

// Close 断开与smtp server的连接
func (s *SMTP) Close(ctx context.Context) error {
	s.mux.Lock()
	defer s.mux.Unlock()
	if !s.isConnected {
		return nil
	}
	if err := s.cli.Close(); err != nil {
		return err
	}
	s.isConnected = false
	return nil
}

// IsConnected 判断当前是否连接到smtp server
func (s *SMTP) IsConnected() bool {
	s.mux.RLock()
	defer s.mux.RUnlock()
	return s.isConnected
}

// Reset 重置与smtp server的连接状态
func (s *SMTP) Reset() error {
	return s.cli.Reset()
}

// Send 发送邮件，发送前需先确保已经调用了Dial建立与smtp server的连接
// 会根据初始化时提供的限流参数进行限流，如果被限流会阻塞，可以通过ctx传入超时控制
func (s *SMTP) Send(ctx context.Context, mails ...*Mail) error {
	if len(mails) == 0 {
		return errors.New("mail list is empty")
	}
	messages, err := s.convert2Messages(mails...)
	if err != nil {
		return err
	}
	return s.cli.Send(messages...)
}

func (s *SMTP) convert2Messages(mails ...*Mail) ([]*mail.Msg, error) {
	var messages []*mail.Msg
	for _, m := range mails {
		message := mail.NewMsg()
		if m.From != nil {
			if m.From.Name != "" {
				if err := message.FromFormat(m.From.Name, m.From.Addr); err != nil {
					return nil, err
				}
			} else {
				if err := message.From(m.From.Addr); err != nil {
					return nil, err
				}
			}
		}
		for _, to := range m.To {
			if to.Name != "" {
				if err := message.AddToFormat(to.Name, to.Addr); err != nil {
					return nil, err
				}
			} else {
				if err := message.AddTo(to.Addr); err != nil {
					return nil, err
				}
			}
		}
		for _, cc := range m.Cc {
			if cc.Name != "" {
				if err := message.AddCcFormat(cc.Name, cc.Addr); err != nil {
					return nil, err
				}
			} else {
				if err := message.AddCc(cc.Addr); err != nil {
					return nil, err
				}
			}
		}
		for _, bcc := range m.Bcc {
			if bcc.Name != "" {
				if err := message.AddBccFormat(bcc.Name, bcc.Addr); err != nil {
					return nil, err
				}
			} else {
				if err := message.AddBcc(bcc.Addr); err != nil {
					return nil, err
				}
			}
		}
		for _, attachment := range m.Attachments {
			message.AttachFile(attachment)
		}
		message.Subject(m.Subject)
		message.SetBodyString(m.ContentType.toRealType(), m.Content)
		messages = append(messages, message)
	}
	return messages, nil
}

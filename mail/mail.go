package mail

import (
	"context"
	"crypto/tls"
	"errors"
	"sync"

	enumv1 "github.com/byteflowing/proto/gen/go/enums/v1"
	mailv1 "github.com/byteflowing/proto/gen/go/mail/v1"
	"github.com/wneessen/go-mail"
)

type SMTP struct {
	cli         *mail.Client
	isConnected bool
	mux         sync.RWMutex
}

func NewSMTP(opts *mailv1.SMTP) (smtp *SMTP, err error) {
	var clientOpts []mail.Option
	clientOpts = append(
		clientOpts,
		mail.WithUsername(opts.UserName),
		mail.WithPassword(opts.Password),
		mail.WithPort(int(opts.Port)),
		mail.WithTLSPolicy(mail.TLSOpportunistic),
		mail.WithSMTPAuth(mail.SMTPAuthAutoDiscover),
	)
	if opts.Timeout.Seconds > 0 {
		clientOpts = append(clientOpts, mail.WithTimeout(opts.Timeout.AsDuration()))
	}
	if opts.SkipTls {
		clientOpts = append(clientOpts, mail.WithTLSConfig(&tls.Config{InsecureSkipVerify: true}))
	}
	cli, err := mail.NewClient(opts.Host, clientOpts...)
	if err != nil {
		return nil, err
	}
	return &SMTP{cli: cli}, nil
}

// DialAndSend 每次都会建立连接发送邮件关闭连接
// 会根据初始化时提供的限流参数进行限流，如果被限流会阻塞，可以通过ctx传入超时控制
func (s *SMTP) DialAndSend(ctx context.Context, mails ...*mailv1.SendMailReq) error {
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
func (s *SMTP) Send(ctx context.Context, mails ...*mailv1.SendMailReq) error {
	if len(mails) == 0 {
		return errors.New("mail list is empty")
	}
	messages, err := s.convert2Messages(mails...)
	if err != nil {
		return err
	}
	return s.cli.Send(messages...)
}

func (s *SMTP) convert2Messages(mails ...*mailv1.SendMailReq) ([]*mail.Msg, error) {
	var messages []*mail.Msg
	for _, m := range mails {
		message := mail.NewMsg()
		if m.From != nil {
			if m.From.Name != nil && *m.From.Name != "" {
				if err := message.FromFormat(*m.From.Name, m.From.Address); err != nil {
					return nil, err
				}
			} else {
				if err := message.From(m.From.Address); err != nil {
					return nil, err
				}
			}
		}
		for _, to := range m.To {
			if to.Name != nil && *to.Name != "" {
				if err := message.AddToFormat(*to.Name, to.Address); err != nil {
					return nil, err
				}
			} else {
				if err := message.AddTo(to.Address); err != nil {
					return nil, err
				}
			}
		}
		for _, cc := range m.Cc {
			if cc.Name != nil && *cc.Name != "" {
				if err := message.AddCcFormat(*cc.Name, cc.Address); err != nil {
					return nil, err
				}
			} else {
				if err := message.AddCc(cc.Address); err != nil {
					return nil, err
				}
			}
		}
		for _, bcc := range m.Bcc {
			if bcc.Name != nil && *bcc.Name != "" {
				if err := message.AddBccFormat(*bcc.Name, bcc.Address); err != nil {
					return nil, err
				}
			} else {
				if err := message.AddBcc(bcc.Address); err != nil {
					return nil, err
				}
			}
		}
		for _, attachment := range m.Attachments {
			message.AttachFile(attachment)
		}
		message.Subject(m.Subject)
		message.SetBodyString(toRealType(m.ContentType), m.Content)
		messages = append(messages, message)
	}
	return messages, nil
}

func toRealType(t enumv1.MailContentType) mail.ContentType {
	switch t {
	case enumv1.MailContentType_MAIL_CONTENT_TYPE_HTML:
		return mail.TypeTextHTML
	case enumv1.MailContentType_MAIL_CONTENT_TYPE_TEXT:
		return mail.TypeTextPlain
	default:
		return mail.TypeTextHTML
	}
}

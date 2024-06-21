package parse

import (
	"crypto/tls"
	"errors"
	"fmt"

	"gitlab.unjx.de/flohoss/mittag/internal/env"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
)

func NewSMTPParser(env *env.Env, subject string) *SMTPParser {
	s := &SMTPParser{
		Host:     env.SMTPHost,
		Port:     env.SMTPPort,
		Username: env.SMTPUsername,
		Password: env.SMTPPassword,
	}
	if err := s.parse(subject); err != nil {
		return s
	}
	return s
}

type SMTPParser struct {
	Host     string
	Port     int
	Username string
	Password string
}

func (p *SMTPParser) parse(subject string) error {
	c, err := client.DialTLS(fmt.Sprintf("%s:%d", p.Host, p.Port), &tls.Config{})
	if err != nil {
		return err
	}
	defer c.Logout()

	if err := c.Login(p.Username, p.Password); err != nil {
		return err
	}

	_, err = c.Select("Inbox", false)
	if err != nil {
		return err
	}

	criteria := imap.NewSearchCriteria()
	criteria.Header.Add("Subject", subject)
	ids, err := c.Search(criteria)
	if err != nil {
		return err
	}

	if len(ids) == 0 {
		return errors.New("no emails found with the specified subject")
	}

	seqset := new(imap.SeqSet)
	seqset.AddNum(ids...)
	messages := make(chan *imap.Message, 10)
	done := make(chan error, 1)
	go func() {
		done <- c.Fetch(seqset, []imap.FetchItem{imap.FetchEnvelope, imap.FetchRFC822}, messages)
	}()
	return nil
}

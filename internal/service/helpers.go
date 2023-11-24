package service

import (
	"fmt"

	"github.com/emzola/bugtracker/pkg/mailer"
	"go.uber.org/zap"
)

// SendEmail is a helper function which the service layer uses to send emails
// in a background goroutine. It accepts a data map, recipient and template.
func (s *Service) SendEmail(data map[string]string, recipient, template string) {
	go func() {
		defer func() {
			if err := recover(); err != nil {
				s.Logger.Info(fmt.Sprintf("%s", err))
			}
		}()
		mailer := mailer.New(s.Config.Smtp.Host, s.Config.Smtp.Port, s.Config.Smtp.Username, s.Config.Smtp.Password, s.Config.Smtp.Sender)
		err := mailer.Send(recipient, template, data)
		if err != nil {
			s.Logger.Info("Failed to send email", zap.Error(err))
		}
	}()
}

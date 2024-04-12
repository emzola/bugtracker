package issuetracker

import (
	"fmt"

	"github.com/emzola/issuetracker/pkg/mailer"
	"go.uber.org/zap"
)

// SendEmail is a helper function which the service layer uses to send emails
// in a background goroutine. It accepts a data map, recipient and template.
func (c *Controller) SendEmail(data map[string]string, recipient, template string) {
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		defer func() {
			if err := recover(); err != nil {
				c.Logger.Info(fmt.Sprintf("%s", err))
			}
		}()
		mailer := mailer.New(c.Config.Smtp.Host, c.Config.Smtp.Port, c.Config.Smtp.Username, c.Config.Smtp.Password, c.Config.Smtp.Sender)
		err := mailer.Send(recipient, template, data)
		if err != nil {
			c.Logger.Info("failed to send email", zap.Error(err))
		}
	}()
}

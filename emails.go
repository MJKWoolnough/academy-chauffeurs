package main

import (
	"crypto/tls"
	"net"
	"net/smtp"
	"net/url"
	"text/template"
)

const DefaultEmailTemplate = ``

var (
	compiledEmailTemplate *template.Template
	emailServer           string
	emailTLS              bool
	emailAuth             smtp.Auth
)

func setEmailVars(server, username, password, templateT string) error {
	// smtp://host:port or smtps://host:port for TLS
	u, err := url.Parse(server)
	if err != nil {
		return err
	}
	t, err := template.New("email").Parse(templateT)
	if err != nil {
		return err
	}
	compiledEmailTemplate = t
	emailServer = u.Hostname()
	if p := u.Port(); p == "" {
		emailServer = u.Hostname() + ":25"
	} else {
		emailServer = u.Hostname() + ":" + p
	}
	emailTLS = u.Scheme == "smtps"
	emailAuth = smtp.PlainAuth("", username, password, u.Host)
	return nil
}

func (c *Calls) PrepareEmail(eventID int64, md *MessageData) error {
	return c.prepareMessage(compiledEmailTemplate, eventID, md)
}

func (c *Calls) SendEmail(md MessageData, e *string) error {
	var (
		event  Event
		client Client
	)
	err := c.GetEvent(md.ID, &event)
	if err != nil {
		return err
	}
	err = c.GetClient(event.ClientID, &client)
	if err != nil {
		return err
	}
	var conn net.Conn
	if emailTLS {
		conn, err = tls.Dial("tcp", emailServer, nil)
	} else {
		conn, err = net.Dial("tcp", emailServer)
	}
	if err != nil {
		return err
	}
	cl, err := smtp.NewClient(conn, emailServer)
	if err == nil {
		err = cl.Auth(emailAuth)
		if err == nil {
			err = cl.Mail(client.Email)
			if err == nil {
				wr, err := cl.Data()
				if err == nil {
					_, err = wr.Write([]byte(md.Message))
					if err == nil {
						err = wr.Close()
						if err == nil {
							err = cl.Quit()
						}
					}
				}
			}
		}
	}
	if err != nil {
		*e = err.Error()
	} else {
		_, err = c.statements[MessageSent].Exec(md.ID)
		if err != nil {
			return err
		}
	}
	return nil
}

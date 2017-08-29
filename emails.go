package main

import (
	"crypto/tls"
	"io"
	"net"
	"net/smtp"
	"net/url"
	"strings"
	"text/template"
)

const DefaultEmailTemplate = ``

var (
	compiledEmailTemplate *template.Template
	emailServer           string
	emailTLS              bool
	emailFrom             string
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
	emailTLS = u.Scheme == "smtps"
	if p := u.Port(); p == "" {
		if emailTLS {
			emailServer = u.Hostname() + ":465"
		} else {
			emailServer = u.Hostname() + ":25"
		}
	} else {
		emailServer = u.Hostname() + ":" + p
	}
	emailAuth = smtp.PlainAuth("", username, password, emailServer)
	emailFrom = username
	return nil
}

func (c *Calls) PrepareEmail(eventID int64, md *MessageData) error {
	return c.prepareMessage(compiledEmailTemplate, eventID, md)
}

func (c *Calls) SendEmail(md MessageData, e *string) error {
	message := md.Message
	headers := make([]byte, 0, len(message)+3)
	for {
		i := strings.Index(message, "\n")
		if i < 0 {
			break
		}
		line := message[:i]
		message = message[i+1:]
		line = strings.TrimRight(line, "\r")
		if line == "" {
			headers = append(headers, "\r\n"...)
			break
		}
		i = strings.Index(line, ":")
		if i > 0 && len(line) > i+1 && (line[:i] == "Subject" || line[:i] == "MIME-version" || line[:i] == "Content-Type") {
			headers = append(headers, (line[:i] + ": ")...)
			headers = append(headers, strings.TrimSpace(line[i+1:])...)
			headers = append(headers, "\r\n"...)
		} else {
			message = line + "\n" + message
			break
		}
	}

	headers = append(headers, "\r\n"...)

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
			err = cl.Mail(emailFrom)
			if err == nil {
				err = cl.Rcpt(client.Email)
				if err == nil {
					var wr io.WriteCloser
					wr, err = cl.Data()
					if err == nil {
						_, err = wr.Write(append(headers, message...))
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

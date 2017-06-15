package main

import (
	"html/template"
	"net/smtp"
	"net/url"
)

const DefaultEmailTemplate = ``

var (
	compiledEmailTemplate *template.Template
	emailServer           string
	emailTLS              bool
	emailAuth             smtp.Auth
)

func setEmailVars(server, username, password, template string) error {
	// smtp://host:port or smtps://host:port for TLS
	u, err := url.Parse(server)
	if err != nil {
		return err
	}
	t, err := template.New("email").Parse(template)
	if err != nil {
		return err
	}
	compiledEmailTemplate = t
	emailServer = u.Hostname()
	if p := u.Port(); p == "" {
		emailServer = u.Hostname() + ":25"
	} else {
		emailTLS = u.Hostname() + ":" + p
	}
	emailTLS = u.Scheme == "smtps"
	emailAuth = smtp.PlainAuth("", username, password, u.Host)
	return nil
}

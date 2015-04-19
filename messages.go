package main

import (
	"io/ioutil"
	"text/template"
	"time"

	"github.com/MJKWoolnough/memio"
	"github.com/MJKWoolnough/textmagic"
)

const DefaultTemplate = ``

var (
	compiledTemplate *template.Template
	text             textmagic.TextMagic
	fromNumber       bool
	from             string
)

type MessageData struct {
	ID      int64
	Message string
}

type MessageVars struct {
	Start, End                                                     time.Time
	From, To, ClientName, DriverName, DriverPhoneNumber, DriverReg string
}

func setMessageVars(username, password, messageTemplate, fromS string, fromNumberB bool) error {
	t, err := template.New("Message").Parse(messageTemplate)
	if err != nil {
		return err
	}
	// test template

	m := MessageVars{
		time.Now(), time.Now(),
		"", "", "", "", "", "",
	}

	if err := t.Execute(ioutil.Discard, &m); err != nil {
		return err
	}

	compiledTemplate = t
	text = textmagic.New(username, password)
	fromNumber = fromNumberB
	from = fromS
	return nil
}

func (c *Calls) PrepareMessage(eventID int64, m *MessageData) error {
	var (
		event  Event
		driver Driver
		client Client
	)
	err := c.GetEvent(eventID, &event)
	if err != nil {
		return err
	}
	err = c.GetDriver(event.DriverID, &driver)
	if err != nil {
		return err
	}
	err = c.GetClient(event.ClientID, &client)
	if err != nil {
		return err
	}
	var buf []byte
	data := MessageVars{
		time.Unix(event.Start/1000, 0), time.Unix(event.End/1000, 0),
		event.From, event.To,
		client.Name,
		driver.Name,
		driver.PhoneNumber,
		driver.RegistrationNumber,
	}
	err = compiledTemplate.Execute(memio.Create(&buf), data)
	if err != nil {
		return err
	}
	m.ID = eventID
	m.Message = string(buf)
	return nil
}

func (c *Calls) SendMessage(md MessageData, e *string) error {
	var (
		event  Event
		client Client
		fromS  string
	)
	err := c.GetEvent(md.ID, &event)
	if err != nil {
		return err
	}
	err = c.GetClient(event.ClientID, &client)
	if err != nil {
		return err
	}
	if fromNumber {
		var driver Driver
		err = c.GetDriver(event.DriverID, &driver)
		if err != nil {
			return err
		}
		from = driver.PhoneNumber
	} else {
		fromS = from
	}
	_, _, _, err = text.Send(md.Message, []string{client.PhoneNumber}, textmagic.From(fromS))
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

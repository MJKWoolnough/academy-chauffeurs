package main

import (
	"io/ioutil"
	"text/template"
	"time"

	"github.com/MJKWoolnough/memio"
	"github.com/MJKWoolnough/textmagic"
)

const DefaultTemplate = `Hello {{.ClientName}},
I'm your Acadamy Chauffeurs driver, {{.DriverName}}, for your appointment at {{.StartTime}}.`

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
	StartDate, StartTime, EndDate, EndTime, From, To, ClientName, DriverName, DriverPhoneNumber, DriverReg string
}

func setMessageVars(username, password, messageTemplate, fromS string, fromNumberB bool) error {
	t, err := template.New("Message").Parse(messageTemplate)
	if err != nil {
		return err
	}
	// test template

	if err := t.Execute(ioutil.Discard, &MessageVars{}); err != nil {
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
	start := time.Unix(event.Start/1000, 0)
	end := time.Unix(event.End/1000, 0)

	data := MessageVars{
		start.Format("02/01/06"),
		start.Format("15:04"),
		end.Format("02/01/06"),
		end.Format("15:04"),
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
	var phoneNumber = client.PhoneNumber
	if phoneNumber[0] == '0' {
		phoneNumber = "44" + phoneNumber[1:]
	}
	_, _, _, err = text.Send(md.Message, []string{phoneNumber}, textmagic.From(fromS))
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

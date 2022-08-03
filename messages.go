package main

import (
	"io/ioutil"
	"strings"
	"text/template"
	"time"

	"vimagination.zapto.org/memio"
	"vimagination.zapto.org/textmagic"
)

const DefaultTemplate = `Hello {{.ClientName}},
I'm your driver, {{.DriverName}}, for your appointment at {{.StartTime}}.`

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
	ID                                                                                                                                                                     int64
	StartDate, StartTime, EndDate, EndTime, From, To, ClientName, ClientPhoneNumber, ClientFirstName, ClientLastName, DriverName, DriverPhoneNumber, DriverReg, Passengers string
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
	return c.prepareMessage(compiledTemplate, eventID, m)
}

func (c *Calls) prepareMessage(ct *template.Template, eventID int64, m *MessageData) error {
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

	names := strings.Split(client.Name, " ")

	data := MessageVars{
		event.ID,
		start.In(time.UTC).Format("02/01/06"),
		start.In(time.UTC).Format("15:04"),
		end.In(time.UTC).Format("02/01/06"),
		end.In(time.UTC).Format("15:04"),
		event.From, event.To,
		client.Name,
		client.PhoneNumber,
		names[0],
		names[len(names)-1],
		driver.Name,
		driver.PhoneNumber,
		driver.RegistrationNumber,
		event.Other,
	}
	err = ct.Execute(memio.Create(&buf), data)
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
	)
	err := c.GetEvent(md.ID, &event)
	if err != nil {
		return err
	}
	err = c.GetClient(event.ClientID, &client)
	if err != nil {
		return err
	}
	vars := make([]textmagic.Option, 0, 1)
	if fromNumber {
		var driver Driver
		err = c.GetDriver(event.DriverID, &driver)
		if err != nil {
			return err
		}
		vars = append(vars, textmagic.From(driver.PhoneNumber))
	}
	phoneNumber := client.PhoneNumber
	if phoneNumber[0] == '0' {
		phoneNumber = "44" + phoneNumber[1:]
	} else if phoneNumber[0] == '+' {
		phoneNumber = phoneNumber[1:]
	}
	_, _, _, err = text.Send(md.Message, []string{phoneNumber}, vars...)
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

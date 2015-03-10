package main

import "database/sql"

func (c *Calls) GetDriver(id int64, d *Driver) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	err := c.statements[ReadDriver].QueryRow(id).Scan(&(*d).Name, &(*d).RegistrationNumber, &(*d).PhoneNumber)
	if err == sql.ErrNoRows {
		return nil
	}
	(*d).ID = id
	return err
}

func (c *Calls) GetClient(id int64, cl *Client) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	err := c.statements[ReadClient].QueryRow(id).Scan(&(*cl).CompanyID, &(*cl).Name, &(*cl).PhoneNumber, &(*cl).Reference)
	if err == sql.ErrNoRows {
		return nil
	}
	(*cl).ID = id
	return err
}

func (c *Calls) GetCompany(id int64, cy *Company) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	err := c.statements[ReadCompany].QueryRow(id).Scan(&(*cy).Name, &(*cy).Address)
	if err == sql.ErrNoRows {
		return nil
	}
	(*cy).ID = id
	return err
}

func (c *Calls) GetEvent(id int64, e *Event) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	err := c.statements[ReadEvent].QueryRow(id).Scan(&(*e).DriverID, &(*e).ClientID, &(*e).Start, &(*e).End, &(*e).From, &(*e).To)
	if err == sql.ErrNoRows {
		return nil
	}
	(*e).ID = id
	return err
}

type SetDriverResponse struct {
	Errors                          bool
	ID                              int64
	NameError, RegError, PhoneError string
}

func (c *Calls) SetDriver(d Driver, resp *SetDriverResponse) error {
	if d.Name == "" {
		resp.Errors = true
		resp.NameError = "Name Required"
	}
	if d.RegistrationNumber == "" {
		resp.Errors = true
		resp.RegError = "Registration Number Required"
	}
	if !ValidMobileNumber(d.PhoneNumber) {
		resp.Errors = true
		resp.PhoneError = "Valid Mobile Phone Number Required"
	}
	var err error
	if !resp.Errors {
		c.mu.Lock()
		defer c.mu.Unlock()
		if d.ID == 0 {
			r, e := c.statements[CreateDriver].Exec(d.Name, d.RegistrationNumber, d.PhoneNumber)
			if e == nil {
				resp.ID, e = r.LastInsertId()
			}
			err = e
		} else {
			_, err = c.statements[UpdateDriver].Exec(d.Name, d.RegistrationNumber, d.PhoneNumber, d.ID)
		}
	}
	return err
}

type SetClientResponse struct {
	Errors                                              bool
	ID                                                  int64
	NameError, CompanyError, PhoneError, ReferenceError string
}

func (c *Calls) SetClient(cl Client, resp *SetClientResponse) error {
	if cl.Name == "" {
		resp.Errors = true
		resp.NameError = "Name Required"
	}
	if cl.CompanyID == 0 {
		resp.Errors = true
		resp.CompanyError = "Company Required"
	} else {
		var cy Company
		if err := c.GetCompany(cl.CompanyID, &cy); err != nil {
			return err
		}
		if cy.ID == 0 {
			resp.Errors = true
			resp.CompanyError = "Valid Company Required"
		}
	}
	if !ValidMobileNumber(cl.PhoneNumber) {
		resp.Errors = true
		resp.PhoneError = "Valid Mobile Phone Number Required"
	}
	if cl.Reference == "" {
		resp.Errors = true
		resp.ReferenceError = "Reference Required"
	}
	var err error
	if !resp.Errors {
		c.mu.Lock()
		defer c.mu.Unlock()
		if cl.ID == 0 {
			r, e := c.statements[CreateClient].Exec(cl.CompanyID, cl.Name, cl.PhoneNumber, cl.Reference)
			if e == nil {
				resp.ID, e = r.LastInsertId()
			}
			err = e
		} else {
			_, err = c.statements[UpdateClient].Exec(cl.CompanyID, cl.Name, cl.PhoneNumber, cl.Reference, cl.ID)
		}
	}
	return err
}

type SetCompanyResponse struct {
	Errors                  bool
	ID                      int64
	NameError, AddressError string
}

func (c *Calls) SetCompany(cy Company, resp *SetCompanyResponse) error {
	if cy.Name == "" {
		resp.Errors = true
		resp.NameError = "Name Required"
	}
	if cy.Address == "" {
		resp.Errors = true
		resp.AddressError = "Address Required"
	}
	var err error
	if !resp.Errors {
		c.mu.Lock()
		defer c.mu.Unlock()
		if cy.ID == 0 {
			r, e := c.statements[CreateCompany].Exec(cy.Name, cy.Address)
			if e == nil {
				resp.ID, e = r.LastInsertId()
			}
			err = e
		} else {
			_, err = c.statements[UpdateCompany].Exec(cy.Name, cy.Address, cy.ID)
		}
	}
	return err
}

type SetEventResponse struct {
	Errors                                                  bool
	ID                                                      int64
	DriverError, ClientError, TimeError, FromError, ToError string
}

func (c *Calls) SetEvent(e Event, resp *SetEventResponse) error {
	if e.DriverID == 0 {
		resp.Errors = true
		resp.DriverError = "Driver Required"
	} else {
		var d Driver
		if err := c.GetDriver(e.DriverID, &d); err != nil {
			return err
		}
		if d.ID == 0 {
			e.DriverID = 0
			resp.Errors = true
			resp.DriverError = "Valid Driver Required"
		}
	}
	if e.ClientID == 0 {
		resp.Errors = true
		resp.ClientError = "Client Required"
	} else {
		var cl Client
		if err := c.GetClient(e.ClientID, &cl); err != nil {
			return err
		}
		if cl.ID == 0 {
			resp.Errors = true
			resp.ClientError = "Valid Client Required"
		}
	}
	if e.Start.IsZero() || e.End.IsZero() {
		resp.Errors = true
		resp.TimeError = "Invalid Time(s)"
	} else if e.DriverID != 0 {
		var exists int64
		c.mu.Lock()
		err := c.statements[EventOverlap].QueryRow(e.ID, e.DriverID, e.Start, e.End).Scan(&exists)
		c.mu.Unlock()
		if err != nil {
			return err
		}
		if exists != 0 {
			resp.Errors = true
			resp.TimeError = "Times clash with existing event"
		}
	}
	if e.From == "" {
		resp.Errors = true
		resp.FromError = "From/Pickup location required"
	}
	if e.To == "" {
		resp.Errors = true
		resp.FromError = "To/Dropoff location required"
	}
	var err error
	if !resp.Errors {
		c.mu.Lock()
		defer c.mu.Unlock()
		if e.ID == 0 {
			r, er := c.statements[CreateEvent].Exec(e.DriverID, e.ClientID, e.Start, e.End, e.From, e.To)
			if er == nil {
				resp.ID, er = r.LastInsertId()
			}
			err = er
		} else {
			_, err = c.statements[UpdateEvent].Exec(e.DriverID, e.ClientID, e.Start, e.End, e.From, e.To, e.ID)
		}
	}
	return err
}

func (c *Calls) RemoveDriver(id int64, _ *struct{}) error {
	_, err := c.statements[DeleteDriver].Exec(id)
	return err
}

func (c *Calls) RemoveClient(id int64, _ *struct{}) error {
	_, err := c.statements[DeleteClient].Exec(id)
	return err
}

func (c *Calls) RemoveCompany(id int64, _ *struct{}) error {
	_, err := c.statements[DeleteCompany].Exec(id)
	return err
}

func (c *Calls) RemoveEvent(id int64, _ *struct{}) error {
	_, err := c.statements[DeleteEvent].Exec(id)
	return err
}
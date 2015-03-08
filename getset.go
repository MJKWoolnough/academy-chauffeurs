package main

func (c Calls) GetDriver(id int64, d *Driver) error {
	return c.s.Get(d)
}

func (c Calls) GetClient(id int64, cl *Client) error {
	return c.s.Get(cl)
}

func (c Calls) GetCompany(id int64, cy *Company) error {
	return c.s.Get(cy)
}

func (c Calls) GetEvent(id int64, e *Event) error {
	return c.s.Get(e)
}

type SetDriverResponse struct {
	Errors                          bool
	ID                              int64
	NameError, RegError, PhoneError string
}

func (c Calls) SetDriver(d Driver, resp *SetDriverResponse) error {
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
		err = c.s.Set(&d)
		resp.ID = d.ID
	}
	return err
}

type SetClientResponse struct {
	Errors                                              bool
	ID                                                  int64
	NameError, CompanyError, PhoneError, ReferenceError string
}

func (c Calls) SetClient(cl Client, resp *SetClientResponse) error {
	if cl.Name == "" {
		resp.Errors = true
		resp.NameError = "Name Required"
	}
	if cl.CompanyID == 0 {
		resp.Errors = true
		resp.CompanyError = "Company Required"
	} else {
		cy := Company{
			ID: cl.CompanyID,
		}
		if err := c.s.Get(&cy); err != nil {
			return err
		}
		if cl.CompanyID == 0 {
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
		err = c.s.Set(&cl)
		resp.ID = cl.ID
	}
	return err
}

type SetCompanyResponse struct {
	Errors                  bool
	ID                      int64
	NameError, AddressError string
}

func (c Calls) SetCompany(cy Company, resp *SetCompanyResponse) error {
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
		err = c.s.Set(&cy)
		resp.ID = cy.ID
	}
	return err
}

type SetEventResponse struct {
	Errors                                                  bool
	ID                                                      int64
	DriverError, ClientError, TimeError, FromError, ToError string
}

func (c Calls) SetEvent(e Event, resp *SetEventResponse) error {
	if e.DriverID == 0 {
		resp.Errors = true
		resp.DriverError = "Driver Required"
	} else {
		d := Driver{
			ID: e.DriverID,
		}
		if err := c.s.Get(&d); err != nil {
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
		cl := Client{
			ID: e.ClientID,
		}
		if err := c.s.Get(&cl); err != nil {
			resp.Errors = true
			resp.ClientError = "Valid Client Required"
		}
	}
	if e.Start.IsZero() || e.End.IsZero() {
		resp.Errors = true
		resp.TimeError = "Invalid Time(s)"
	} else if e.DriverID != 0 {
		search := c.searches["events"]
		params := search.params.(*EventsFilter)
		params.mu.Lock()
		params.DriverID = e.DriverID
		params.From = e.Start
		params.To = e.End
		params.NotID = e.ID
		if n, err := search.ps.Count(); err != nil {
			return err
		} else if n != 0 {
			resp.Errors = true
			resp.TimeError = "Times clash with existing event"
		}
		params.NotID = 0
		params.mu.Unlock()
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
		err = c.s.Set(&e)
		resp.ID = e.ID
	}
	return err
}

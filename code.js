"use strict";
window.onload = function() {
	var onload = function() {
		rpc.drivers(function(drivers) {
			if (typeof drivers === "undefined" || drivers === null || drivers.length === 0) {
				stack.addLayer("setDriver", onload);
				setDriver();
			} else {
				eventListWithDrivers(new Date(), drivers);
			}
		});
	},
	rpc = new (function(onload){
		var ws = new WebSocket("ws://127.0.0.1:8080/rpc"),
		requests = [],
		nextID = 0,
		request = function (method, params, callback) {
			var msg = {
				"method": "Calls." + method,
				"id": nextID,
				"params": [params],
			};
			requests[nextID] = callback;
			ws.send(JSON.stringify(msg));
			nextID++;
		};
		ws.onmessage = function (event) {
			var data = JSON.parse(event.data),
			req = requests[data.id];
			delete requests[data.id];
			if (typeof req === "undefined") {
				return;
			} else if (data.error !== null) {
				alert(data.error);
				return;
			}
			req(data.result);
		};
		ws.onopen = onload;
		ws.onerror = function(event) {
			document.body.innerHTML = "An error occurred";
		}
		ws.onclose = function(event) {
			if (event.code !== 1000) {
				document.body.innerHTML = "Lost Connection To Server! Code: " + event.code;
			}
		}
		window.addEventListener("beforeunload", function() {
			ws.close();
		});
		this.getDriver     = request.bind(this, "GetDriver");     // id     , callback
		this.getClient     = request.bind(this, "GetClient");     // id     , callback
		this.getCompany    = request.bind(this, "GetCompany");    // id     , callback
		this.getEvent      = request.bind(this, "GetEvent");      // id     , callback
		this.setDriver     = request.bind(this, "SetDriver");     // driver , callback
		this.setClient     = request.bind(this, "SetClient");     // client , callback
		this.setCompany    = request.bind(this, "SetCompany");    // company, callback
		this.setEvent      = request.bind(this, "SetEvent");      // event  , callback
		this.removeDriver  = request.bind(this, "RemoveDriver");  // id     , callback
		this.removeClient  = request.bind(this, "RemoveClient");  // id     , callback
		this.removeCompany = request.bind(this, "RemoveCompany"); // id     , callback
		this.removeEvent   = request.bind(this, "RemoveEvent");   // id     , callback
		this.drivers       = request.bind(this, "Drivers", null); // callback
		this.getEventsWithDriver = function(driverID, start, end, callback) {
			request("Events", {"DriverID": driverID, "Start": start, "End": end}, callback);
		}
		this.autocompleteAddress = function(priority, partial, callback) {
			request("AutocompleteAddress", {"Priority": priority, "Partial": partial}, callback);
		}
	})(onload),
	createElement = (function(){
		var ns = document.getElementsByTagName("html")[0].namespaceURI;
		return function(elementName) {
			return document.createElementNS(ns, elementName);
		};
	}()),
	layer,
	stack = new (function(){
		var stack = [],
		canceler = [],
		body = document.body;
		this.addLayer = function(layerID, callback) {
			if (stack.length == 0) {
				canceler.push(null);
			} else {
				canceler.push(this.removeLayer.bind(this));
			}
			stack.push(callback);
			var outerLayer = createElement("div");
			outerLayer.className = "layer";
			layer = createElement("div");
			layer.setAttribute("id", layerID);
			outerLayer.appendChild(layer);
			body.appendChild(outerLayer);
		};
		this.removeLayer = function() {
			if (stack.length === 0) {
				return;
			}
			var callback = stack.pop();
			canceler.pop();
			if (typeof callback === "function") {
				callback.apply(arguments);
			}
			body.removeChild(body.lastChild);
			layer = body.lastChild.firstChild;
		};
		this.addFragment = function () {
			if (typeof layer == "object" && layer.nodeType !== 11) {
				layer = document.createDocumentFragment();
			}
		};
		this.setFragment = function () {
			if (typeof layer == "object" && layer.nodeType === 11) {
				body.lastChild.firstChild.appendChild(layer);
				layer = body.lastChild.firstChild;
			}
		};
		this.clearLayer = function(callback) {
			return function() {
				while (layer.hasChildNodes()) {
					layer.removeChild(layer.lastChild);
				}
				callback();
			};
		};
		this.setCanceler = function(callback) {
			canceler[canceler.length-1] = callback;
		};
		document.addEventListener("keypress", function(e) {
			if (canceler[canceler.length-1] !== null) {
				e = e || window.event;
				if (e.keyCode === 27) {
					canceler[canceler.length-1]();
				}
			}
		});
	})(),
	timeFormat = function(date) {
		return 
	},
	dateTimeFormat = function(date) {
		return date.toLocaleString('en-GB');
	},
	eventList = function(date) {
		if (arguments.length == 0) {
			date = new Date();
		}
		rpc.drivers(eventListWithDrivers.bind(null, date));
	},
	eventListWithDrivers = function (date, drivers) {
		var f = eventListWithData.bind(null, date, drivers),
		i = 0,
		events = [],
		start, end;
		for (;i < drivers.length; i++) {
			f = function(callback, num) {
				return function() {
					rpc.getEventsWithDriver(drivers[num].ID, start, end, function(e) {
						events[num] = e;
						callback(events);
					});
				};
			} (f, i);
		}
		f();
	},
	eventListWithData = function (date, drivers, events) {
		stack.addFragment();
		// generate dates
		// generate times
		var i = 0,
		ypos = 200,
		layer.appendChild(dateDiv);
		for (; i < drivers.length; i++) {
			var driver = createElement("div"),
			driverName = createElement("div");
			driver.setAttribute("class", "driverName");
			driver.style.top = ypos + "px";
			driver.addEventListener("click", (function(driver) {
				return function() {
					stack.addLayer("viewDriver", stack.clearLayer(eventList));
					viewDriver(driver);
				}
			}(drivers[i])));
			driverName.innerHTML = drivers[i].Name;
			ypos += 100;
			driver.appendChild(driverName);
			layer.appendChild(driver);
		}
		var addDriver = createElement("div"),
		plus = createElement("div");
		addDriver.setAttribute("id", "addDriver");
		addDriver.style.top = ypos + "px";
		plus.innerHTML = "+";
		addDriver.appendChild(plus);
		addDriver.addEventListener("click", function() {
			stack.addLayer("setDriver", stack.clearLayer(eventList));
			setDriver();
		});
		layer.appendChild(addDriver);
		stack.setFragment();
	},
	addTitle = function(id, add, edit) {
		layer.appendChild(createElement("h1")).innerHTML = (id == 0) ? add : edit;
	},
	viewDriver = function(driver) {
		alert(driver.Name);
	},
	addFormElement = function(name, type, id, contents, onBlur) {
		var label = createElement("label"),
		input;
		if (type === "textarea") {
			input = createElement("textarea");
			input.innerHTML = contents;
		} else {
			input = createElement("input");
			input.setAttribute("type", type);
			input.setAttribute("value", contents);
		}
		label.innerHTML = name;
		if (id === "") {
			input.setAttribute("readonly", "readonly");
			layer.appendChild(label);
			layer.appendChild(input);
		} else {
			label.setAttribute("for", id);
			input.setAttribute("id", id);
			if (typeof onBlur === "function") {
				input.addEventListener("blur", onBlur.bind(input));
			}
			var error = createElement("div");
			error.setAttribute("class", "error");
			error.setAttribute("id", "error_"+id);
			layer.appendChild(label);
			layer.appendChild(input);
			layer.appendChild(error);
		}
		layer.appendChild(createElement("br"));
		return input;
	},
	addFormSubmit = function(value, onClick) {
		var button = createElement("input");
		button.setAttribute("value", value);
		button.setAttribute("type", "button");
		button.addEventListener("click", onClick.bind(button));
		return layer.appendChild(button);
	},
	disableElement = function(part) {
		part.setAttribute("disabled", "disabled");
	},
	enableElement = function(part) {
		part.removeAttribute("disabled");
	},
	setDriver = function(id) {
		if (typeof id === "number" && id > 0) {
			rpc.getDriver(id, setDriverWithData);
		} else {
			setDriverWithData({
				"ID": 0,
				"Name": "",
				"RegistrationNumber": "",
				"PhoneNumber": "",
			});
		}
	},
	setDriverWithData = function(driver) {
		stack.addFragment();
		addTitle(driver.ID, "Add Driver", "Edit Driver");
		var driverName = addFormElement("Driver Name", "text", "driver_name", driver.Name, regexpCheck(/.+/, "Please enter a valid name")),
		regNumber = addFormElement("Registration Number", "text", "driver_reg", driver.RegistrationNumber, regexpCheck(/[a-zA-Z0-9 ]+/, "Please enter a valid Vehicle Registration Number")),
		phoneNumber = addFormElement("Phone Number", "text", "driver_phone", driver.PhoneNumber, regexpCheck(/^(0|\+?44)[0-9 ]{10}$/, "Please enter a valid mobile telephone number"));
		addFormSubmit("Add Driver", function() {
			var parts = [this, driverName, regNumber, phoneNumber];
			parts.map(disableElement);
			rpc.setDriver({
				"ID": driver.ID,
				"Name": driverName.value,
				"RegistrationNumber": regNumber.value,
				"PhoneNumber": phoneNumber.value,
			}, function(resp) {
				if (resp.Errors) {
					layer.getElementById("error_driver_name").innerHTML = resp.NameError;
					layer.getElementById("error_driver_reg").innerHTML = resp.RegError;
					layer.getElementById("error_driver_phone").innerHTML = resp.PhoneError;
					parts.map(enableElement);
				} else {
					stack.removeLayer();
				}
			});
		});
		stack.setFragment();
	},
	setClient = function(id) {
		if (typeof id === "number" && id > 0) {
			rpc.getClient(id, function(resp) {
				var client = resp;
				rpc.getCompany(client.CompanyID, function(resp) {
					client.CompanyID = resp.ID;
					client.CompanyName = resp.Name;
					setClientWithData(client);
				});
			});
		} else {
			setClientWithData({
				"ID": 0,
				"Name": "",
				"CompanyName": "",
				"CompanyID": 0,
				"PhoneNumber": "",
				"Reference": "",
			});
		}
	},
	setClientWithData = function(client) {
		stack.addFragment();
		addTitle(client.ID, "Add Client", "Edit Client");
		var clientName = addFormElement("Client Name", "text", "client_name", client.Name, regexpCheck(/.+/, "Please enter a valid name")),
		companyID = addFormElement("", "hidden", "client_company_id", client.CompanyID),
		companyName = addFormElement("Company Name", "text", "client_company_name", client.CompanyName, regexpCheck(/.+/, "Please enter a valid name")),
		clientPhone = addFormElement("Mobile Number", "text", "client_phone", client.PhoneNumber, regexpCheck(/^(0|\+?44)[0-9 ]{10}$/, "Please enter a valid mobile telephone number")),
		clientRef = addFormElement("Client Ref", "text", "client_ref", client.Reference, regexpCheck(/.+/, "Please enter a reference code"));
		autocomplete(rpc.autocompleteCompanyName, companyName, companyID);
		addFormSubmit("Add Client", function() {
			var parts = [this, clientName, companyID, companyName];
			parts.map(disableElement);
			rpc.setClient({
				"ID": client.ID,
				"Name": clientName.value,
				"CompanyID": companyID.value,
				"PhoneNumber": clientPhone.value,
				"Reference": clientRef.value,
			}, function (resp) {
				if (resp.errors) {
					layer.getElementById("error_name").innerHTML = resp.NameError;
					layer.getElementById("error_company_name").innerHTML = resp.CompanyNameError;
					layer.getElementById("error_phone").innerHTML = resp.PhoneError;
					layer.getElementById("error_ref").innerHTML = resp.RefError;
					parts.map(enableElement);
				} else {
					stack.removeLayer(resp.ID, clientName.value);
				}
			});
		});
		stack.addFragment();
	},
	setCompany = function(id) {
		if (typeof id === "number" && id > 0) {
			rpc.getCompany(id, setCompanyWithData);
		} else {
			setCompanyWithData({
				"ID": 0,
				"Name": "",
				"Address": "",
			});
		}
	},
	setCompanyWithData = function(company) {
		stack.addFragment();
		addTitle(company.ID, "Add Company", "Edit Company");
		var companyName = addFormElement("Company Name", "text", "company_name", company.Name, regexpCheck(/.+/, "Please enter a valid name")),
		address = addFormElement("Company Address", "textarea", "company_address", company.Address, regexpCheck(/.+/, "Please enter a valid address"));
		addFormSubmit("Add Company", function() {
			var parts = [this, companyName, address];
			parts.map(disableElement);
			rpc.setCompany({
				"ID": company.ID,
				"Name": companyName.value,
				"Address": address.innerHTML,
			}, function(resp) {
				if (resp.Errors) {
					layer.getElementById("error_company_name").innerHTML = resp.NameError;
					layer.getElementById("error_company_address").innerHTML = resp.AddressError;
					parts.map(enableElement);
				} else {
					stack.removeLayer(resp.ID, companyName.value);
				}
			});
		});
		stack.setFragment();
	},
	setEvent = function(id, startTime, endTime) {
		if (arguments.length > 1) {
			rpc.getDriver(id, function(resp) {
				setEventWithData({
					"ID": 0,
					"Start": startTime,
					"End": endTime,
					"From": "",
					"To": "",
					"ClientID": 0,
					"ClientName": "",
					"DriverID": resp.ID,
					"DriverName": resp.Name,
				});
			});
		} else {
			rpc.getEvent(id, function(resp) {
				var event = resp;
				rpc.getClient(event.ClientID, function(resp) {
					event.ClientID = resp.ID;
					event.ClientName = resp.Name;
					rpc.getDriver(event.DriverID, function(resp) {
						event.DriverID = resp.ID;
						event.DriverName = resp.Name;
						event.Start = new Date(event.Start * 1000); // ms
						event.End = new Date(event.End * 1000); // ms
						setEventWithData(event);
					});
				});
			});
		}
	},
	setEventWithData = (function() {
		var fromAddressRPC = rpc.autocompleteAddress.bind(rpc, 0),
		toAddressRPC = rpc.autocompleteAddress.bind(rpc, 1);
		return function(event) {
			stack.addFragment();
			addTitle(event.ID, "Add Event", "Edit Event");
			addFormElement("Driver", "text", "", event.DriverName);
			addFormElement("Start", "text", "", dateTimeFormat(event.Start));
			addFormElement("End", "text", "", dateTimeFormat(event.End));
			var changeDriverTime = addFormElement("Change Above", "button", "change_driver_time"),
			from = addFormElement("From", "textarea", "from", event.From),
			to = addFormElement("To", "textarea", "to", event.To)
			clientID = addFormElement("", "hidden", "", event.ClientID),
			clientName = addFormElement("Client Name", "text", "client_name", event.ClientName);
			changeDriverTime.addEventListener("click", function() {
				
			}.bind(changeDriverTime));
			autocomplete(fromAddressRPC, from);
			autocomplete(toAddressRPC, to);
			autocomplete(autocompleteClientName, clientName, clientID);
			addFormSubmit("Add Event", function() {
				var parts = [this, changeTime, to, from];
				parts.map(disableElement);
				event.From = from.innerHTML;
				event.To = to.innerHTML;
				rpc.setEvent(event, function(resp) {
					if (resp.errors) {
						layers.getElementById("error_change_driver_time").innerHTML = resp.DriverTimeError;
						layers.getElementById("error_from").innerHTML = resp.FromError;
						layers.getElementById("error_to").innerHTML = resp.ToError;
						parts.map(enableElement);
					} else {
						stack.removeLayer(resp.ID, event.Start / 1000 | 0);
					}
				});
			});
			stack.setFragment();
		}
	}()),
	regexpCheck = function(regexp, error) {
		return function() {
			var errorDiv = document.getElementById("error_" + this.getAttribute("id"));
			if (this.value.match(regexp)) {
				errorDiv.innerHTML = "";
			} else {
				errorDiv.innerHTML = error;
			}
		}
	},
	autocomplete = function(rpcCall, name, id) {
		
	};
	stack.addLayer("eventList");
	Date.prototype.isLeapYear = function() {
		var year = this.getFullYear();
		return year % 4 === 0 && (year % 100 !== 0 || year % 400 === 0);
	}
};

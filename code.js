"use strict";
window.onload = function() {
	var rpc = new (function(onload){
		var ws = new WebSocket("ws://127.0.0.1:8080/rpc", "rpc"),
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
		this.getDriver = function(id, callback) {
			request("GetDriver", id, callback);
		}
		this.getClient = function(id, callback) {
			request("GetClient", client, callback);
		}
		this.getCompany = function(id, callback) {
			request("GetCompany", id, callback);
		}
		this.getEvent = function(id, callback) {
			request("GetEvent", id, callback);
		}
		this.setDriver = function(driver, callback) {
			request("SetDriver", driver, callback);
		}
		this.setClient = function(client, callback) {
			request("SetClient", client, callback);
		}
		this.setCompany = function(company, callback) {
			request("SetCompany", company, callback);
		}
		this.setEvent = function(event, callback) {
			request("SetEvent", event, callback);
		}
		this.removeDriver = function(id, callback) {
			request("RemoveDriver", id, callback);
		}
		this.removeClient = function(id, callback) {
			request("RemoveClient", id, callback);
		}
		this.removeCompany = function(id, callback) {
			request("RemoveCompany", id, callback);
		}
		this.removeEvent = function(id, callback) {
			request("RemoveEvent", id, callback);
		}
		this.drivers = function(callback) {
			request("Drivers", 0, callback);
		}
		this.events = function(driverID, start, end, callback) {
			request("Events", {"DriverID": driverID, "Start": start, "End": end}, callback);
		}
		ws.onopen = onload;
	})(function() {
		eventList();
	}),
	createElement = (function(){
		var ns = document.getElementsByTagName("html")[0].namespaceURI;
		return function(elementName) {
			return createElementNS(ns, elementName);
		};
	}()),
	drivers = [],
	companies = [],
	clients = [],
	events = [],
	layer,
	body = document.body,
	stack = new (function(){
		var stack = [];
		this.addLayer = function(layerID, callback) {
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
			if (typeof callback === "function") {
				callback.apply(arguments);
			}
			body.removeChild(body.lastChild);
			layer = body.lastChild.firstChild;
		};
		this.addFragment() {
			if (typeof layer == "object" && layer.nodeType !== 11) {
				layer = document.createDocumentFragment();
			}
		}
		this.setFragment() {
			if (typeof layer == "object" && layer.nodeType === 11) {
				body.lastChild.firstChild.appendChild(layer);
				layer = body.lastChild.firstChild;
			}
		}
		this.addLayer("eventList");
	})(),
	eventList = function(date) {
		if (arguments.length == 0) {
			date = Date.now()
		}
		rpc.drivers(function(d) {
			drivers = d;
			if (drivers.length === 0) {
				stack.addLayer("addDriver", eventList);
				setDriver();
			} else {
			}
		});
	},
	addTitle = function(id, add, edit) {
		layer.appendChild(createElement("h1")).innerHTML = (id == 0) ? add : edit;
	},
	addFormElement = function(name, type, id, contents, onBlur) {
		var label = createElement("label"),
		error = createElement("div"),
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
		label.setAttribute("for", id);
		input.setAttribute("id", id);
		if (typeof onBlur === "function") {
			input.addEventListener("blur", onBlur.bind(input));
		}
		error.setAttribute("class", "error");
		error.setAttribute("id", "error_"+id);
		layer.appendChild(label);
		layer.appendChild(input);
		layer.appendChild(error);
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
					// add driver to a list?
					stack.removeLayer();
				}
			});
		});
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
		addTitle(client.ID, "Add Client", "Edit Client");
		var clientName = addFormElement("Client Name", "text", "client_name", client.Name, regexpCheck(/.+/, "Please enter a valid name")),
		companyID = addFormElement("", "hidden", "client_company_id", client.CompanyID),
		companyName = addFormElement("Company Name", "text", "client_company_name", client.CompanyName, regexpCheck(/.+/, "Please enter a valid name")),
		clientPhone = addFormElement("Mobile Number", "text", "client_phone", client.PhoneNumber, regexpCheck(/^(0|\+?44)[0-9 ]{10}$/, "Please enter a valid mobile telephone number")),
		clientRef = addFormElement("Client Ref", "text", "client_ref", client.Reference, regexpCheck(/.+/, "Please enter a reference code");
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
		addTitle(company.ID, "Add Company", "Edit Company");
		var companyName = addFormElement("Company Name", "text", "company_name", company.Name, regexpCheck(/.+/, "Please enter a valid name")),
		address = addFormElement("Company Address", "textarea", "company_address", company.Address, regexpCheck(/.+/, "Please enter a valid address"));
		addFormSubmit("Add Company", function() {
			var parts = [this, companyName, address];
			parts.map(disableElement);
			rpc.setCompany({
				"ID": company.ID,
				"Name", companyName.value,
				"Address", address.innerHTML,
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
						setEventWithData(event);
					});
				});
			});
		}
	},
	setEventWithData = function(event) {
		addTitle(event.ID, "Add Event", "Edit Event");

	},
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
};

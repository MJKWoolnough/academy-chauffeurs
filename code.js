"use strict";
window.onload = function() {
	var rpc = new (function(onload){
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
	})(function() {
		events.init();	
	}),
	createElement = (function(){
		var ns = document.getElementsByTagName("html")[0].namespaceURI;
		return function(elementName) {
			return document.createElementNS(ns, elementName);
		};
	}()),
	layer = document.body,
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
	dateTimeFormat = function(date) {
		return date.toLocaleString('en-GB');
	},
	events = new (function() {
		var dateTime,
		    dateShift = (new Date()).getTime(),
		    eventList = createElement("div"),
		    drivers = [],
		    startEnd = [dateShift, dateShift],
		    plusDriver = createElement("div"),
		    nextDriverPos = 100,
		    months = ["Januray", "February", "March", "April", "May", "June", "July", "August", "September", "October", "November", "December"],
		    days = ["Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"],
		    eventClicked = function(driver, time) {
			    
		    },
		    timeToPos = function(date) {
			return ((date.getTime() - dateShift) / 60000) + "px"
		    },
		    update = function(date) {
			if (typeof date === "undefined") {
				date = dateTime;
			} else {
				dateTime = date;
			}
			var unix = date.getTime(),
			    screenWidth = window.innerWidth,
			    mins = (dateShift - unix) / 60000,
			    minOnScreen = unix - ((screenWidth / 2) * 60000),
			    maxOnScreen = unix + ((screenWidth / 2) * 60000),
			    minOnScreenDayStart = minOnScreen - (minOnScreen % 86400000),
			    tDate, year, month, day, t,
			    toCenter = {}, keys, object,
			    newEventListPos = (screenWidth / 2) - mins;
			if (minOnScreenDayStart < startEnd[0]) {
				startEnd[0] = minOnScreenDayStart;
			}
			if (maxOnScreen > startEnd[1]) {
				startEnd[1] = maxOnScreen;
			}
			for (t = startEnd[0]; t < startEnd[1]; t += 86400000) {
				tDate = new Date(t);
				year = tDate.getFullYear();
				month = tDate.getMonth();
				day = tDate.getDate();
				if (addDay(year, month, day)) {
					// TODO: get events
				}
				toCenter["year_" + year] = true;
				toCenter["month_" + year + "_" + month] = true;
				toCenter["day_" + year + "_" + month + "_" + day] = true;
			}
			eventList.style.left = newEventListPos + "px";
			keys = Object.keys(toCenter);
			for (t = 0; t < keys.length; t++) {
				object = document.getElementById(keys[t]);
				if (isOnScreen(object)) {
					var textWidth = object.firstChild.offsetWidth,
					    width = parseInt(object.style.width),
					    left = parseInt(object.style.left, 10) + newEventListPos;
					if (left + (textWidth / 2) > screenWidth / 2) {
						object.firstChild.style.left = "0px";
					} else if (left + width > (screenWidth + textWidth) / 2) {
						object.firstChild.style.left = ((screenWidth - textWidth) / 2) - left + "px"; 
					} else {
						object.firstChild.style.left = width - textWidth + "px";
					}
				}
			}
			startEnd[1] = new Date(t);

		    },
		    addYear = function (year) {
			var yearDate = new Date(year, 0, 1),
			    yearDiv = createElement("div"),
			    textDiv = yearDiv.appendChild(createElement("div"));
			textDiv.innerHTML = year;
			textDiv.setAttribute("class", "slider");
			yearDiv.setAttribute("class", "year");
			yearDiv.setAttribute("id", "year_" + year);
			yearDiv.style.left = timeToPos(yearDate);
			if (yearDate.isLeapYear()) {
				yearDiv.style.width = "527040px";
			} else {
				yearDiv.style.width = "525600px";
			}
			eventList.appendChild(yearDiv);
		    },
		    addMonth = function(year, month) {
			if (document.getElementById("year_" + year) === null) {
				addYear(year);
			}
			var monthDate = new Date(year, month, 1),
			    monthDiv = createElement("div"),
			    textDiv = monthDiv.appendChild(createElement("div"));
			textDiv.innerHTML = months[month];
			textDiv.setAttribute("class", "slider");
			monthDiv.setAttribute("class", "month");
			monthDiv.setAttribute("id", "month_" + year + "_" + month);
			monthDiv.style.left = timeToPos(monthDate);
			monthDiv.style.width = (monthDate.daysInMonth() * 24 * 60) + "px";
			eventList.appendChild(monthDiv);
		    },
		    addDay = function(year, month, day) {
			if (document.getElementById("day_" + year + "_" + month + "_" + day) !== null) {
				return false;
			}
			if (document.getElementById("month_" + year + "_" + month) === null) {
				addMonth(year, month);
			}
			var dayDate = new Date(year, month, day),
			    dayDiv = createElement("div"),
			    textDiv = dayDiv.appendChild(createElement("div")),
			    i = 0;
			textDiv.innerHTML = days[dayDate.getDay()] + ", " + day + dayDate.getOrdinalSuffix();
			textDiv.setAttribute("class", "slider");
			dayDiv.setAttribute("class", "day");
			dayDiv.setAttribute("id", "day_" + year + "_" + month + "_" + day);
			dayDiv.style.left = timeToPos(dayDate);
			dayDiv.style.width = "1440px";
			eventList.appendChild(dayDiv);
			for (; i < 24; i++) {
				addHour(year, month, day, i);
			}
			return true;
		    },
		    addHour = function(year, month, day, hour) {
			var hourDate = new Date(year, month, day, hour),
			    hourDiv = createElement("div");
			hourDiv.setAttribute("class", "hour");
			hourDiv.innerHTML = formatNum(hour);
			hourDiv.style.left = timeToPos(hourDate);
			hourDiv.style.width = "60px";
			eventList.appendChild(hourDiv);
			addFifteen(year, month, day, hour, 0);
			addFifteen(year, month, day, hour, 1);
			addFifteen(year, month, day, hour, 2);
			addFifteen(year, month, day, hour, 3);
		    },
		    addFifteen = function(year, month, day, hour, block) {
			var fifteenDate = new Date(year, month, day, hour, block * 15),
			    fifteenDiv = createElement("div");
			fifteenDiv.setAttribute("class", "minute");
			fifteenDiv.innerHTML = formatNum(block * 15);
			fifteenDiv.style.left = timeToPos(fifteenDate);
			fifteenDiv.style.width = "15px";
			eventList.appendChild(fifteenDiv);
			// TODO: Add driver boxes
		    },
		    isOnScreen = function(div) {
			var offsets = div.getBoundingClientRect();
			return !(offsets.left > window.innerWidth || offsets.right < 0);
		    },
		    formatNum = function(num) {
			if (num < 10) {
				return "0" + num;
			}
			return num;
		    },
		    init = function() {
			init = function() {};
			rpc.drivers(function(ds) {
				for (var i = 0; i < ds.length; i++) {
					this.addDriver(ds[i]);
					drivers[ds[i].ID] = [];
				}
				plusDriver.appendChild(createElement("div")).innerHTML = "+";
				plusDriver.setAttribute("id", "plusDriver");
				plusDriver.addEventListener("click", function() {
					stack.addLayer("addDriver", this.addDriver.bind(this));
					addDriver();
				}.bind(this));
				layer.appendChild(plusDriver);
				layer.appendChild(eventList).setAttribute("class", "events");
				update(new Date());
			}.bind(this));
		    };
		this.init = function() {
			init.call(this);
		};
		this.addDriver = function(d) {
			if (typeof d === "undefined") {
				return;
			}
			drivers[d.ID] = [];
			var dDiv = createElement("div"),
			    t;
			dDiv.appendChild(createElement("div")).innerHTML = d.Name;
			dDiv.setAttribute("class", "driverName");
			dDiv.setAttribute("id", "driver_" + d.ID);
			dDiv.addEventListener("click", function() {
				stack.addLayer("viewDriver");
				viewDriver(d, drivers[d.ID]);
			});
			dDiv.style.top = nextDriverPos + "px";
			nextDriverPos += 100;
			plusDriver.style.top = nextDriverPos + "px";
			layer.appendChild(dDiv);
			for (t = startEnd[0]; t < startEnd[1]; t += 900) {
				// TODO: add time boxes
			}
			// TODO: get events for existing boxes
		};
		this.setTime = function (time) {
			dateTime = time;
			update();
		}
	})(),
	addTitle = function(id, add, edit) {
		layer.appendChild(createElement("h1")).innerHTML = (id == 0) ? add : edit;
	},
	viewDriver = function(driver, events) {
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
	setDriver = function(driver) {
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
	addDriver = setDriver.bind(null, {
		"ID": 0,
		"Name": "",
		"RegistrationNumber": "",
		"PhoneNumber": "",
	}),
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
			    to = addFormElement("To", "textarea", "to", event.To),
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
	stack.addLayer("events");
	Date.prototype.isLeapYear = function() {
		var year = this.getFullYear();
		return year % 4 === 0 && (year % 100 !== 0 || year % 400 === 0);
	}
	Date.prototype.daysInMonth = function() {
		return (new Date(this.getFullYear(), this.getMonth() + 1, 0)).getDate()
	}
	Date.prototype.getOrdinalSuffix = function() {
		var suf = ["th","st","nd","rd"],
		    v = this.getDate() % 100;
		return suf[(v - 20) % 10] || suf[v] || suf[0];
	}
};

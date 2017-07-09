"use strict";
window.addEventListener("load", function(oldDate) {
	var rpc = new (function(onload){
		var ws = new WebSocket("ws://127.0.0.1:" + window.location.port + "/rpc"),
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
		    },
		    closed = false;
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
			document.body.setInnerText("An error occurred");
		}
		ws.onclose = function(event) {
			if (closed === true) {
				return;
			}
			switch(event.code) {
			case 1006:
				document.body.setInnerText("The server unexpectedly closed the connection - this may be an error.");
				break;
			case 4000:
				document.body.setInnerText("The server closed the connection due to another session opening.");
				break;
			default:
				document.body.setInnerText("Lost Connection To Server! Code: " + event.code);
			}
		}
		window.addEventListener("beforeunload", function() {
			closed = true;
			ws.close();
		});
		this.getDriver                 = request.bind(this, "GetDriver");              // id         , callback
		this.getClient                 = request.bind(this, "GetClient");              // id         , callback
		this.getCompany                = request.bind(this, "GetCompany");             // id         , callback
		this.getEvent                  = request.bind(this, "GetEvent");               // id         , callback
		this.getEventFinals            = request.bind(this, "GetEventFinals");         // id         , callback
		this.setDriver                 = request.bind(this, "SetDriver");              // driver     , callback
		this.setClient                 = request.bind(this, "SetClient");              // client     , callback
		this.setCompany                = request.bind(this, "SetCompany");             // company    , callback
		this.setEvent                  = request.bind(this, "SetEvent");               // event      , callback
		this.setEventFinals            = request.bind(this, "SetEventFinals");         // eventFinals
		this.removeDriver              = request.bind(this, "RemoveDriver");           // id         , callback
		this.removeClient              = request.bind(this, "RemoveClient");           // id         , callback
		this.removeCompany             = request.bind(this, "RemoveCompany");          // id         , callback
		this.removeEvent               = request.bind(this, "RemoveEvent");            // id         , callback
		this.getDriverNote             = request.bind(this, "GetDriverNote");          // id         , callback
		this.getClientNote             = request.bind(this, "GetClientNote");          // id         , callback
		this.getCompanyNote            = request.bind(this, "GetCompanyNote");         // id         , callback
		this.getEventNote              = request.bind(this, "GetEventNote") ;          // id         , callback
		this.getNumClients             = request.bind(this, "NumClients");             // id         , callback
		this.getNumClientsForCompanies = request.bind(this, "NumClientsForCompanies"); // [ids]      , callback
		this.getNumEvents              = request.bind(this, "NumEvents");              // id         , callback
		this.getNumEventsCompanies     = request.bind(this, "NumEventsForCompanies");  // [ids]      , callback
		this.getNumEventsClient        = request.bind(this, "NumEventsClient");        // id         , callback
		this.getNumEventsDriver        = request.bind(this, "NumEventsDriver");        // id         , callback
		this.drivers                   = request.bind(this, "Drivers", null);          //              callback
		this.companies                 = request.bind(this, "Companies", null);        //              callback
		this.clients                   = request.bind(this, "Clients", null);          //              callback
		this.unsentMessages            = request.bind(this, "UnsentMessages", null);   //              callback
		this.prepareMessage            = request.bind(this, "PrepareMessage");         // id         , callback
		this.prepareEmail              = request.bind(this, "PrepareEmail");           // id         , callback
		this.sendMessage               = request.bind(this, "SendMessage");            // messageData, callback
		this.clientsForCompany         = request.bind(this, "ClientsForCompany");      // id         , callback
		this.clientsForCompanies       = request.bind(this, "ClientsForCompanies");    // [ids]      , callback
		this.getSettings               = request.bind(this, "GetSettings", null);      //              callback
		this.setSettings               = request.bind(this, "SetSettings");            // settings   , callback
		this.setDriverPosShow          = request.bind(this, "SetDriverPosShow");       // {ID, Pos, Show}, callback
		this.getEventsWithDriver = function(driverID, start, end, callback) {
			request("DriverEvents", {"ID": driverID, "Start": start, "End": end}, callback);
		};
		this.getEventsWithClient = function(clientID, start, end, callback) {
			request("ClientEvents", {"ID": clientID, "Start": start, "End": end}, callback);
		};
		this.getEventsWithCompany = function(companyID, start, end, callback) {
			request("CompanyEvents", {"ID": companyID, "Start": start, "End": end}, callback);
		};
		this.getEventsWithCompanies = function(companyIDs, start, end, callback) {
			request("CompaniesEvents", {"IDs": companyIDs, "Start": start, "End": end}, callback);
		};
		this.setDriverNote = function(id, note, callback) {
			request("SetDriverNote", {ID: id, Note: note}, callback);
		};
		this.setClientNote = function(id, note) {
			request("SetClientNote", {ID: id, Note: note});
		};
		this.setCompanyNote = function(id, note) {
			request("SetCompanyNote", {ID: id, Note: note});
		};
		this.setEventNote = function(id, note) {
			request("SetEventNote", {ID: id, Note: note});
		};
		this.autocompleteAddress = function(priority, clientID, partial, callback) {
			request("AutocompleteAddress", {"ClientID": clientID, "Priority": priority, "Partial": partial}, callback);
		};
		this.autocompleteCompanyName = request.bind(this, "AutocompleteCompanyName"); // partial,  callback
		this.autocompleteClientName = request.bind(this, "AutocompleteClientName");   // partial,  callback
		this.getCompanyColourFromClient = request.bind(this, "CompanyColour");        // clientID, callback
		this.getFirstUnassigned = request.bind(this, "FirstUnassigned", null); //callback
		this.getUnassignedCount = request.bind(this, "UnassignedCount", null); //callback
		this.getUsers = request.bind(this, "GetUsers", null); //callback
		this.setUser = request.bind(this, "SetUser"); // {Username, Password}, callback
		this.removeUser = request.bind(this, "RemoveUser"); // Username, callback
	})(function() {
		events.init();	
	}),
	waitGroup = function(callback) {
		var state = 0;
		this.add = function(amount) {
			amount = amount || 1;
			state += amount;
		};
		this.done = function() {
			state--;
			if (state === 0) {
				callback();
			}
		};
	},
	vatPercent = 20,
	adminPercent = 10,
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
		    body = document.body,
		    oLayer;
		this.addLayer = function(layerID, callback) {
			if (this.layerExists(layerID)) {
				return;
			}
			if (stack.length == 0) {
				canceler.push(null);
			} else {
				canceler.push(this.removeLayer.bind(this));
			}
			stack.push(callback);
			var outerLayer = createElement("div"),
			    cancelButton = createElement("div");
			outerLayer.style.zIndex = stack.length + 1;
			outerLayer.className = "layer";
			layer = createElement("div");
			layer.setAttribute("id", layerID);
			if (stack.length > 1) {
				cancelButton.setAttribute("class", "canceller");
				cancelButton.setInnerText("X");
				cancelButton.addEventListener("click", this.removeLayer.bind(this, undefined));
			}
			layer.appendChild(cancelButton);
			outerLayer.appendChild(layer);
			body.appendChild(outerLayer);
		};
		this.layerExists = function(layerID) {
			return document.getElementById(layerID) !== null;
		}
		this.removeLayer = function() {
			if (stack.length === 0) {
				return;
			}
			body.removeChild(body.lastChild);
			layer = body.lastChild.firstChild;
			var callback = stack.pop();
			canceler.pop();
			if (typeof callback === "function") {
				callback.apply(null, arguments);
			}
		};
		this.addFragment = function () {
			if (typeof layer == "object" && layer.nodeType !== 11) {
				oLayer = layer;
				layer = document.createDocumentFragment();
			}
		};
		this.setFragment = function () {
			if (typeof layer == "object" && layer.nodeType === 11) {
				oLayer.appendChild(layer);
				layer = oLayer;
			}
		};
		this.clearLayer = function(callback) {
			return function() {
				layer.removeChildren()
				callback();
			};
		};
		this.setCanceler = function(callback) {
			canceler[canceler.length-1] = callback;
		};
		document.addEventListener("keydown", function(e) {
			if (canceler[canceler.length-1] !== null) {
				e = e || window.event;
				if (e.keyCode === 27) {
					canceler[canceler.length-1]();
				}
			}
		});
	})(),
	addAdder = function(elementBefore, callback) {
		var adder = createElement("div");
		adder.setInnerText("+");
		adder.addEventListener("click", callback);
		adder.setAttribute("class", "adder");
		layer.insertBefore(adder, elementBefore);
	},
	addLister = function(elementBefore, callback) {
		var adder = createElement("div");
		adder.setInnerText("←");
		adder.addEventListener("click", callback);
		adder.setAttribute("class", "adder");
		elementBefore.parentNode.insertBefore(adder, elementBefore);
	},
	dateTimeFormat = function(date) {
		return (new Date(date)).toLocaleString('en-GB');
	},
	users = function() {
		layer.appendChild(createElement("h1")).setInnerText("Users");
		rpc.getUsers(function(u) {
			for (var i = 0; i < u.length; i++) {
				var username = layer.appendChild(createElement("label").setInnerText(u[i])),
				    password = layer.appendChild(createElement("input"));
				username.setAttribute("for", "user_"+i);
				password.setAttribute("type", "text");
				password.setAttribute("class", "password");
				password.setAttribute("id", "user_"+i);
				addFormSubmit("✔", function(username, password) {
					if (confirm("Are you sure you wish to modify the password for '" + username + "'?")) {
						rpc.setUser({Username:username, Password: password.value}, window.location.reload.bind(window.location));
					}
				}.bind(null, u[i], password));
				if (u[i] !== "admin") {
					addFormSubmit("X", function(username) {
						if (confirm("Are you sure you wish to remove user '" + username + "'?")) {
							rpc.removeUser(username, window.location.reload.bind(window.location));
						}
					}.bind(null, u[i]));
				}
				layer.appendChild(createElement("br"));
				layer.appendChild(createElement("br"));
			}
			addFormSubmit("Add User", function() {
				var username = prompt("Please enter a username"),
				    password = username && prompt("Please enter a password");
				if (typeof username === "string" && typeof password === "string" && username != "" && password != "") {
					rpc.setUser({Username: username, Password: password}, window.location.reload.bind(window.location));
				}
			});
			stack.setFragment();
		});
	},
	settings = function() {
		layer.appendChild(createElement("h1")).setInnerText("Settings");
		rpc.getSettings(function(s) {
			var username = addFormElement("Text Magic Username", "text", "tmusername", s.TMUsername, regexpCheck(/.+/, "Please enter your Text Magic username")),
			    password = addFormElement("Text Magic Password", "text", "tmpassword", s.TMPassword, regexpCheck(/.+/, "Please enter your Text Magic password")),
			    template = addFormElement("Message Template", "textarea", "template", s.TMTemplate, regexpCheck(/.*/, "Please enter a valid message template")),
			    senderID = addFormElement("Sender ID", "text", "senderID", s.TMFrom, regexpCheck(/.+/, "Please enter a sender ID")),
			    useNumber = addFormElement("Driver # as Sender", "checkbox", "useNumber", s.TMUseNumber),
			    vat = addFormElement("VAT (%)", "text", "vat", s.VATPercent, regexpCheck(/^[0-9]+(\.[0-9]+)?$/, "Please enter a valid number")),
			    admin = addFormElement("Admin Cost (%)", "text", "admin", s.AdminPercent, regexpCheck(/^[0-9]+(\.[0-9]+)?$/, "Please enter a valid number")),
			    unass = addFormElement("Unassigned events warning (days)", "text", "uass", s.Unassigned, regexpCheck(/^[0-9]+$/, "Please enter a valid integer")),
			    alarmTime = addFormElement("Calendar Export Alarm Time (m)", "text", "alarmTime", s.AlarmTime, regexpCheck(/^-?[0-9]+$/, "Please enter a valid integer")),
			    serverPort = addFormElement("Server Port", "text", "port", s.Port, regexpCheck(/^[0-9]+$/, "Please enter a valid integer")),
			    invoiceHeader = addFormElement("Invoice Header", "text", "invoiceHeader", s.InvoiceHeader, regexpCheck(/.*/, "Please enter a valid invoice header")),
			    emailSMTP = addFormElement("SMTP Server", "text", "smtpServer", s.EmailSMTP, regexpCheck(/.*/, "Please enter a valid SMTP server")),
			    emailUsername = addFormElement("Email Username", "text", "emailUsername", s.EmailUsername, regexpCheck(/.*/, "Please enter a valid Email Username")),
			    emailPassword = addFormElement("Email Password", "password", "emailPassword", s.EmailPassword, regexpCheck(/.*/, "Please enter a valid Email Password")),
			    emailTemplate = addFormElement("Email Template", "textarea", "emailTemplate", s.EmailTemplate, regexpCheck(/.*/, "Please enter a valid Email Template"));

			useNumber[0].addEventListener("change", function() {
				if (useNumber[0].checked) {
					senderID[0].value = s.TMFrom;
					senderID[0].setAttribute("readonly", "readonly");
				} else {
					senderID[0].removeAttribute("readonly");
				}
			});
			useNumber[0].dispatchEvent(new MouseEvent("change", {"view": window, "bubble": false, "cancelable": true}));
			addFormSubmit("Set Settings", function() {
				var error = false;
				[username, password, template, vat, admin].map(function(i) {
					if (i[1].hasChildNodes()) {
						error = true;
					}
				});
				if (error) {
					return;
				}
				s.TMUsername = username[0].value;
				s.TMPassword = password[0].value;
				s.TMTemplate = template[0].value;
				s.TMFrom = senderID[0].value;
				s.TMUseNumber = useNumber[0].checked;
				s.VATPercent = parseFloat(vat[0].value);
				s.AdminPercent = parseFloat(admin[0].value);
				s.Unassigned = parseInt(unass[0].value);
				s.Port = parseInt(serverPort[0].value);
				s.AlarmTime = parseInt(alarmTime[0].value);
				s.InvoiceHeader = invoiceHeader[0].value;
				s.EmailSMTP = emailSMTP[0].value;
				s.EmailUsername = emailUsername[0].value;
				s.EmailPassword = emailPassword[0].value;
				s.EmailTemplate = emailTemplate[0].value;
				rpc.setSettings(s, function(templateError) {
					if (templateError === "") {
						window.location.search = '';
					} else {
						template[1].setInnerText(templateError);
					}
				});
			});
			stack.setFragment();
		});
	},
	events = new (function() {
		var dateTime,
		    dateShift,
		    driverEvents = createElement("div"),
		    unassignedEvents = [],
		    unassignedNear = 1000 * 3600 * 24 * 7,
		    eventCells = driverEvents.appendChild(createElement("div")),
		    dates = createElement("div"),
		    days = {},
		    drivers = [],
		    startEnd = [dateShift, dateShift],
		    plusDriver = driverEvents.appendChild(createElement("div")),
		    nextDriverPos = 0,
		    timeToPos = function(date) {
			return Math.floor((date.getTime() - dateShift) / 60000) + "px";
		    },
		    update = function(date) {
			if (typeof date === "undefined") {
				date = dateTime;
			} else {
				dateTime = date;
			}
			var unix = date.getTime(),
			    screenWidth = window.innerWidth,
			    mins = (unix - dateShift) / 60000,
			    minOnScreen = unix - ((screenWidth / 2) * 60000),
			    maxOnScreen = unix + ((screenWidth / 2) * 60000),
			    minOnScreenDayStart = minOnScreen - (minOnScreen % 86400000) - 86400000,
			    maxOnScreenDayEnd = maxOnScreen - (maxOnScreen % 86400000) + 2 * 86400000,
			    tDate, year, month, day, t, i,
			    toCenter = {}, keys, object,
			    newEventListPos = (screenWidth / 2) - mins;
			if (minOnScreenDayStart < startEnd[0]) {
				startEnd[0] = minOnScreenDayStart;
			}
			if (maxOnScreenDayEnd > startEnd[1]) {
				startEnd[1] = maxOnScreenDayEnd;
			}
			keys = Object.keys(days);
			for (t = 0; t < keys.length; t++) {
				var node = days[keys[t]][0];
				if (node.parentNode !== null) {
					var parts = keys[t].split("_");
					unix = (new Date(parts[0], parts[1], parts[2])).getTime();
					if (unix < minOnScreenDayStart || unix > maxOnScreenDayEnd) {
						dates.removeChild(days[keys[t]][0]);
						eventCells.removeChild(days[keys[t]][1]);
						eventCells.removeChild(days[keys[t]][2]);
					}
				}
			}
			for (t = minOnScreenDayStart; t < maxOnScreenDayEnd; t += 86400000) {
				tDate = new Date(t);
				year = tDate.getFullYear();
				month = tDate.getMonth();
				day = tDate.getDate();
				if (addDay(year, month, day)) {
					var driverIDs = Object.keys(drivers);
					driverIDs.push(0);
					for (i = 0; i < driverIDs.length; i++) {
						if (driverIDs[i] === 0 || drivers[parseInt(driverIDs[i])].Show !== false) {
							rpc.getEventsWithDriver(parseInt(driverIDs[i]), tDate.getTime(), tDate.getTime() + 86400000, function(events) {
								for(var i = 0; i < events.length; i++) {
									addEventToTable(events[i]);
								}
							});
						}
					}
				}
				toCenter["year_" + year] = true;
				toCenter["month_" + year + "_" + month] = true;
				toCenter["day_" + year + "_" + month + "_" + day] = true;
			}
			eventCells.style.left = newEventListPos + "px";
			dates.style.left = newEventListPos + "px";
			keys = Object.keys(toCenter);
			for (t = 0; t < keys.length; t++) {
				object = document.getElementById(keys[t]);
				if (isOnScreen(object)) {
					var textWidth = object.firstChild.offsetWidth,
					    width = object.offsetWidth,
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
		    },
		    addYear = function (year) {
			var yearDate = new Date(year, 0, 1),
			    yearDiv = createElement("div");
			yearDiv.appendChild(createElement("div")).setInnerText(year).setAttribute("class", "slider");
			yearDiv.setAttribute("class", "year");
			yearDiv.setAttribute("id", "year_" + year);
			yearDiv.style.left = timeToPos(yearDate);
			if (yearDate.isLeapYear()) {
				yearDiv.style.width = "527040px";
			} else {
				yearDiv.style.width = "525600px";
			}
			dates.appendChild(yearDiv);
		    },
		    addMonth = function(year, month) {
			if (document.getElementById("year_" + year) === null) {
				addYear(year);
			}
			var monthDate = new Date(year, month),
			    monthDiv = createElement("div"),
			    monthEnclosure = createElement("div");
			monthDiv.appendChild(createElement("div")).setInnerText(monthDate.getMonthName()).setAttribute("class", "slider");
			monthDiv.setAttribute("class", "month");
			monthDiv.setAttribute("id", "month_" + year + "_" + month);
			monthDiv.style.left = timeToPos(monthDate);
			monthDiv.style.width = (monthDate.daysInMonth() * 24 * 60) + "px";
			dates.appendChild(monthDiv);
		    },
		    addDay = function(year, month, day) {
			if (typeof days[year + "_" + month + "_" + day] !== "undefined") {
				dates.appendChild(days[year + "_" + month + "_" + day][0]);
				eventCells.appendChild(days[year + "_" + month + "_" + day][1]);
				eventCells.appendChild(days[year + "_" + month + "_" + day][2]);
				return;
			} else if (document.getElementById("month_" + year + "_" + month) === null) {
				addMonth(year, month);
			}
			var dayDate = new Date(year, month, day),
			    dayDiv = createElement("div"),
			    dayEnclosure = createElement("div"),
			    i = 0,
			    unassigned = eventCells.appendChild(createElement("div"));
			unassigned.setAttribute("class", "driverUnassigned");
			unassigned.style.top = nextDriverPos + "px";
			dayDiv.appendChild(createElement("div")).setInnerText(dayDate.getDayName() + ", " + day + dayDate.getOrdinalSuffix()).setAttribute("class", "slider");
			dayDiv.setAttribute("class", "day");
			dayDiv.setAttribute("id", "day_" + year + "_" + month + "_" + day);
			dayDiv.style.left = timeToPos(dayDate);
			dayEnclosure.appendChild(dayDiv);
			dayEnclosure.setAttribute("class", "dayEnclosure");

			days[year + "_" + month + "_" + day] = [dayEnclosure, eventCells.appendChild(createElement("div")), unassigned];
			for (; i < 24; i++) {
				addHour(year, month, day, i);
			}
			dates.appendChild(dayEnclosure);
			return true;
		    },
		    addHour = function(year, month, day, hour) {
			var hourDate = new Date(year, month, day, hour),
			    hourDiv = createElement("div");
			hourDiv.setAttribute("class", "hour simpleButton");
			hourDiv.setInnerText(formatNum(hour));
			hourDiv.style.left = timeToPos(hourDate);
			hourDiv.addEventListener("click", update.bind(null, hourDate));
			days[year + "_" + month + "_" + day][0].appendChild(hourDiv);
			addFifteen(year, month, day, hour, 0);
			addFifteen(year, month, day, hour, 1);
			addFifteen(year, month, day, hour, 2);
			addFifteen(year, month, day, hour, 3);
		    },
		    addFifteen = function(year, month, day, hour, block) {
			var fifteenDate = new Date(year, month, day, hour, block * 15),
			    fifteenDiv = createElement("div"),
			    dayDiv = days[year + "_" + month + "_" + day],
			    driverIDs = Object.keys(drivers),
			    cellDiv,
			    leftPos = timeToPos(fifteenDate),
			    vDrivers = 0;
			fifteenDiv.setAttribute("class", "minute");
			fifteenDiv.setAttribute("id", "minute_" + year + "_" + month + "_" + day + "_" + hour + "_" + block);
			fifteenDiv.setInnerText(formatNum(block * 15));
			fifteenDiv.style.left = leftPos;
			dayDiv[0].appendChild(fifteenDiv);
			driverIDs.sort(function(a, b) {
				return drivers[a].Pos - drivers[b].Pos;
			});
			driverIDs.push(0);
			for (var i = 0; i < driverIDs.length; i++) {
				if (driverIDs[i] != 0 && drivers[driverIDs[i]].Show !== true) {
					continue;
				}
				cellDiv = createElement("div");
				cellDiv.setAttribute("class", "eventCell " + (block % 2 === vDrivers % 2 ? "cellOdd" : "cellEven"));
				cellDiv.setAttribute("id", "cell_" + driverIDs[i] + "_" + year + "_" + month + "_" + day + "_" + hour + "_" + block);
				cellDiv.style.left = leftPos;
				cellDiv.addEventListener("mouseover", eventOnMouseOver);
				cellDiv.addEventListener("mouseover", fifteenDiv.setAttribute.bind(fifteenDiv, "class", "minute select"));
				cellDiv.addEventListener("mouseout", eventOnMouseOut);
				cellDiv.addEventListener("mouseout", fifteenDiv.setAttribute.bind(fifteenDiv, "class", "minute"));
				cellDiv.addEventListener("click", eventOnClick);
				if (driverIDs[i] === 0) {
					cellDiv.style.top = "0px";
					dayDiv[2].appendChild(cellDiv);
				} else {
					cellDiv.style.top = drivers[driverIDs[i]].yPos + "px";
					dayDiv[1].appendChild(cellDiv);
				}
				vDrivers++;
			}
		    },
		    isOnScreen = function(div) {
			var left = parseInt(eventCells.style.left, 10) + parseInt(div.style.left, 10),
			    right = left + div.offsetWidth;
			return !(left > window.innerWidth || right < 0);
		    },
		    formatNum = function(num) {
			if (num < 10) {
				return "0" + num;
			}
			return num;
		    },
		    init = function() {
			init = function() {};
			stack.addLayer("events");
			stack.addFragment();
			var n = new Date(),
			    now = new Date(n.getFullYear(), n.getMonth(), n.getDate(), 12),
			    addToBar = function() {
				var topBar = layer.appendChild(createElement("div"));
				topBar.setAttribute("id", "topBar");
				return function(text, callback) {
					var item = topBar.appendChild(createElement("div"));
					item.setInnerText(text);
					item.setAttribute("class", "simpleButton");
					item.addEventListener("click", callback);
					return item;
				};
			    }(),
			    params = window.location.search.substring(1).split("&"), i = 0, paramParts, toLoad = [];
			for (; i < params.length; i++) {
				paramParts = params[i].split("=");
				if (params[0] === "settings") {
					settings();
					return;
				} else if (params[0] === "users") {
					users();
					return;
				} else if (paramParts.length === 2) {
					var id = parseInt(paramParts[1]);
					switch (paramParts[0]){
					case "date":
						now = new Date(parseInt(paramParts[1]));
						break;
					case "driver":
						if (id !== 0) {
							toLoad = [function(id) {
								var driver = drivers[id];
								if (typeof driver !== "undefined") {
									stack.addLayer("showDriver");
									showDriver({
										"ID": driver.ID,
										"Name": driver.Name,
										"PhoneNumber": driver.PhoneNumber,
										"RegistrationNumber": driver.RegistrationNumber
									});
								}
							}.bind(null, id)]
						}
						break;
					case "client":
						toLoad = [function() {
							stack.addLayer("clientList");
							clientList();
						}];
						if (id !== 0) {
							toLoad[1] = rpc.getClient.bind(null, id, function(client) {
								if (client.ID !== 0) {
									rpc.getCompany(client.CompanyID, function(company) {
										if (company.ID !== 0) {
											client.CompanyName = company.Name;
											stack.addLayer("showClient");
											showClient(client);
										}
									});
								}
							});
						}
						break;
					case "company":
						toLoad = [function() {
							stack.addLayer("companyList");
							companyList();
						}];
						if (id !== 0) {
							toLoad[1] = rpc.getCompany.bind(null, id, function(company) {
								if (company.ID !== 0) {
									stack.addLayer("showCompany");
									showCompany(company);
								}
							});
						}
						break;
					case "event":
						if (id !== 0) {
							toLoad = [rpc.getEvent.bind(null, id, function(e) {
								if (e.ID !== 0) {
									stack.addLayer("showEvent");
									showEvent(e);
								}
							})]
						}
						break;
					}
				}
			}
			rpc.getSettings(function (s) {
				unassignedNear = s.Unassigned * 24 * 3600 * 1000;
				vatPercent = s.VATPercent;
				adminPercent = s.AdminPercent;
				addToBar("Companies", function() {
					stack.addLayer("companyList");
					companyList();
				});
				addToBar("Clients", function() {
					stack.addLayer("clientList");
					clientList();
				});
				addToBar("Drivers", function() {
					var ps = Object.keys(drivers),
					    vs = {};
					ps = ps.sort(function(a, b) {
						return drivers[a].Pos - drivers[b].Pos;
					});
					drivers.forEach(function(d) {
						vs[d.ID] = d.Show;
					});
					stack.addLayer("driverList", function() {
						var aps = Object.keys(drivers),
						    wg = new waitGroup(function() {
							this.removeDriver(0);
						    }.bind(this)),
						    wgDone = wg.done.bind(wg);
						aps = aps.sort(function(a, b) {
							return drivers[a].Pos - drivers[b].Pos;
						});
						for (var i = 0; i < aps.length; i++) {
							if (ps[i] != aps[i] || drivers[aps[i]].Show !== vs[drivers[aps[i]].ID]) {
								wg.add();
								var id = parseInt(ps[i]);
								rpc.setDriverPosShow({"ID": id, "Pos": drivers[id].Pos, "Show": drivers[id].Show}, wgDone);
								/*
								rpc.getDriverNote(id, function(id, noteText) {
									var note = noteJSON(noteText);
									note.Pos = drivers[id].Pos;
									note.Show = drivers[id].Show;
									rpc.setDriverNote(id, JSON.stringify(note), function() {
										wg.done();
									});
								}.bind(null, id))
								*/
							}
						}
					}.bind(this));
					driverList(drivers);
				}.bind(this));
				addToBar("Messages", messageList);
				checkUnassigned(addToBar("", goToNextUnassigned));
				dateShift = now.getTime();
				rpc.drivers(function(ds) {
					plusDriver.appendChild(createElement("div")).setInnerText("Add Driver");
					plusDriver.setAttribute("id", "plusDriver");
					plusDriver.setAttribute("class", "simpleButton");
					plusDriver.addEventListener("click", function() {
						stack.addLayer("addDriver", this.addDriver.bind(this));
						addDriver();
					}.bind(this));
					//var wg = new waitGroup(function() {
						ds = ds.sort(function(a, b) {
							return a.Pos - b.Pos;
						});
						for (i = 0; i < ds.length; i++) {
							this.addDriver(ds[i]);
						}
						var eventsDiv = layer.appendChild(createElement("div"));
						eventsDiv.setAttribute("class", "dates");
						driverEvents.setAttribute("class", "driverEvents");
						eventCells.setAttribute("class", "events slider");
						layer.appendChild(dates).setAttribute("class", "dates slider");
						layer.appendChild(driverEvents);
						for (i = 0; i < 10; i++) {
							var div = layer.appendChild(createElement("div"));
							if (i % 2 === 0) {
								div.appendChild(createElement("div")).setInnerText("<");
								div.setAttribute("class", "moveLeft simpleButton");
							} else {
								div.appendChild(createElement("div")).setInnerText(">");
								div.setAttribute("class", "moveRight simpleButton");
							}
							div.style.top = 20 + Math.floor(i / 2) * 20 + "px";
							div.addEventListener("click", moveHandler(i));
							if (i === 4 || i === 5) {
								div.addEventListener("dblclick", moveHandler(i+6));
							}
						}
						stack.setFragment();
						update(now);
						for (i = 0; i < toLoad.length; i++) {
							toLoad[i]();
						}
						window.addEventListener("resize", update.bind(this, undefined));
					/*}.bind(this)) , i;
					for (i = 0; i < ds.length; i++) {
						wg.add();
						rpc.getDriverNote(ds[i].ID, function(i, note)  {
							var n = noteJSON(note);
							if (typeof n.Pos === "undefined") {
								ds[i].Pos = i;
							} else {
								ds[i].Pos = n.Pos
							}
							if (n.Show === false) {
								ds[i].Show = false;
							} else {
								ds[i].Show = true;
							}
							wg.done();
						}.bind(null, i))
					}
					*/
				}.bind(this));
			}.bind(this));
		    },
		    goToNextUnassigned = function() {
			rpc.getFirstUnassigned(function(unix) {
				if (unix > 0) {
					update(new Date(unix));
				}
			});
		    },
		    unassignedCheck,
		    checkUnassigned = function(i) {
			unassignedCheck = rpc.getUnassignedCount.bind(rpc, function(num) {
				if (num > 0) {
					i.setInnerText(num);
					rpc.getFirstUnassigned(function(unix) {
						if (unix - (new Date()).getTime() < unassignedNear) {
							i.setAttribute("class", "simpleButton nearPulse");
						} else {
							i.setAttribute("class", "simpleButton pulse");
						}
					});
				} else {
					i.setInnerText("");
					i.setAttribute("class", "simpleButton");
				}
			});
			window.setInterval(unassignedCheck, 10000);
			unassignedCheck();
		    },
		    moveHandler = function(buttNum) {
			var yearShift = 0,
			    monthShift = 0,
			    dayShift = 0,
			    hourShift = 0,
			    minuteShift = 0;
			switch (buttNum) {
			case 0:
				yearShift = -1;
				break;
			case 1:
				yearShift = 1;
				break;
			case 2:
				monthShift = -1;
				break;
			case 3:
				monthShift = 1;
				break;
			case 4:
				dayShift = -1;
				break;
			case 5:
				dayShift = 1;
				break;
			case 6:
				hourShift = -1;
				break;
			case 7:
				hourShift = 1;
				break;
			case 8:
				minuteShift = -15;
				break;
			case 9:
				minuteShift = 15;
				break;
			case 10:
				dayShift = -5;
				break;
			case 11:
				dayShift = 5;
				break;
			}
			return function() {
				update(new Date(dateTime.getFullYear() + yearShift, dateTime.getMonth() + monthShift, dateTime.getDate() + dayShift, dateTime.getHours() + hourShift, dateTime.getMinutes() + minuteShift));
			};
		    },
		    cellIdToDriver = function(id) {
			var parts = id.split("_");
			return parseInt(parts[1], 10);
		    },
		    cellIdToDate = function(id) {
			var parts = id.split("_");
			return new Date(parts[2], parts[3], parts[4], parts[5], parts[6] * 15).getTime();
		    },
		    getEventsBetween = function(thatID) {
			if (eventSelected === null) {
				return null;
			}
			var thisID = eventSelected.getAttribute("id"),
			    thisDriverID = cellIdToDriver(thisID),
			    thatDriverID = cellIdToDriver(thatID),
			    thisTime = cellIdToDate(thisID),
			    thatTime = cellIdToDate(thatID);
			if (thisDriverID !== thatDriverID || thisTime >= thatTime) {
				return null;
			}
			return eventCellsBetween(thisDriverID, thisTime + 900000, thatTime + 90000);
		    },
		    eventCellsBetween = function(driverID, thisTime, thatTime) {
			var events = [],
			    t;
			for (t = thisTime; t < thatTime; t += 900000) {
				var tDate = new Date(t),
				    year = tDate.getFullYear(),
				    month = tDate.getMonth(),
				    day = tDate.getDate(),
				    hour = tDate.getHours(),
				    block = tDate.getMinutes() / 15,
				    cell;
				if (driverID === 0) {
					cell = days[year + "_" + month + "_" + day][2].getChildElementById("cell_0_" + year + "_" + month + "_" + day + "_" + hour + "_" + block);
				} else {
					cell = days[year + "_" + month + "_" + day][1].getChildElementById("cell_" + driverID + "_" + year + "_" + month + "_" + day + "_" + hour + "_" + block);
				}
				if (cell === null) {
					return null;
				}
				events.push(cell);
			}
			return events;
		    },
		    changeThirdCellClass = function(cell, newClass) {
			var parts = cell.getAttribute("class").split(" ");
			if (parts.length > 2) {
				parts[2] = newClass;
			} else {
				parts.push(newClass);
			}
			cell.setAttribute("class", parts.join(" "));
		    },
		    eventOnMouseOver = function(e) {
			e = e || event;
			if (moverSelected !== null) {
				var thisID = moverSelected.getAttribute("id"),
				    thatID = e.target.getAttribute("id"),
				    driverID = cellIdToDriver(thatID),
				    startDate = cellIdToDate(thatID),
				    endDate = startDate + parseInt(moverSelected.getAttribute("timeWidth")),
				    cells = eventCellsBetween(driverID, startDate, endDate),
				    i = 0;
				if (cells === null) {
					return;
				}
				for (; i < cells.length; i++) {
					changeThirdCellClass(cells[i], "eventsInBetween");
				}
				eventsHighlighted = cells;
				return;
			}
			if (e.target === eventSelected) {
				return;
			}
			if (eventSelected !== null) {
				var cells = getEventsBetween(e.target.getAttribute("id"));
				if (cells === null) {
					return;
				}
				for (var i = 0; i < cells.length; i++) {
					changeThirdCellClass(cells[i], "eventsInBetween");
				}
				eventsHighlighted = cells;
			} else {
				changeThirdCellClass(e.target, "eventHover");
				eventsHighlighted = [e.target];
			}
		    },
		    eventOnMouseOut = function() {
			for (var i = 0; i < eventsHighlighted.length; i++) {
				changeThirdCellClass(eventsHighlighted[i], "");
			}
			eventsHighlighted = [];
		    },
		    eventSelected = null,
		    eventsHighlighted = [],
		    moverSelected = null,
		    moverFifteenSelected = null,
		    eventOnClick = function(e) {
			e = e || event;
			if (moverSelected !== null) {
				var newPlaceID = e.target.getAttribute("id"),
				    newDriverID = cellIdToDriver(newPlaceID),
				    newStartTime = cellIdToDate(newPlaceID);
				rpc.getEvent(parseInt(moverSelected.getAttribute("eventID")), function(ev) {
					if (ev.DriverID === 0) {
						for (var i = 0; i < unassignedEvents.length; i++) {
							if (unassignedEvents[i].ID === ev.ID) {
								unassignedEvents.splice(i, 1);
								break;
							}
						}
					}
					moverSelected.parentNode.removeChild(moverSelected);
					moverSelected = null;
					ev.DriverID = newDriverID;
					ev.End = (ev.End - ev.Start) + newStartTime;
					ev.Start = newStartTime;
					rpc.setEvent(ev, function() {
						addEventToTable(ev);
						unassignedCheck();
						e.target.dispatchEvent(new MouseEvent("mouseout", {"view": window, "bubble": false, "cancelable": true}));
					});
				});
			} else if (e.target === eventSelected) {
				eventSelected = null;
				changeThirdCellClass(e.target, "eventHover");
				eventsHighlighted.push(e.target);
			} else if (eventSelected === null) {
				eventSelected = e.target;
				changeThirdCellClass(e.target, "eventSelected");
				eventsHighlighted = [];
			} else if (getEventsBetween(e.target.getAttribute("id")) !== null){
				eventsHighlighted.push(eventSelected);
				var id = e.target.getAttribute("id"),
				    driverID = cellIdToDriver(id);
				stack.addLayer("addEvent", addEventToTable);
				if (driverID === 0) {
					addEvent({ID: 0, Name: "Unassigned"}, new Date(cellIdToDate(eventSelected.getAttribute("id"))), new Date(cellIdToDate(id) + 900000));
				} else {
					addEvent(drivers[driverID], new Date(cellIdToDate(eventSelected.getAttribute("id"))), new Date(cellIdToDate(id) + 900000));
				}
				eventSelected = null;
			}
		    }.bind(this),
		    moveEvent = function(eventDiv, e) {
			e = e || window.event;
			e.stopPropagation();
			if (eventSelected !== null) {
				eventOnClick({target: eventSelected});
			}
			if (moverSelected === eventDiv) {
				if (moverFifteenSelected !== null) {
					eventDiv.appendChild(moverFifteenSelected);
					moverFifteenSelected = null;
				}
				moverSelected = null;
				e.target.setAttribute("class", "eventMover");
			} else if (moverSelected === null) {
				moverFifteenSelected = eventDiv.querySelector(".eventCell");
				if (moverFifteenSelected !== null) {
					eventDiv.parentNode.appendChild(moverFifteenSelected);
				}
				moverSelected = eventDiv;
				e.target.setAttribute("class", "eventMover selected");
			}
		    },
		    addEventToTable = function(e) {
			if (typeof e === "undefined") {
				return;
			}
			var eventDate = new Date(e.Start),
			    year = eventDate.getFullYear(),
			    month = eventDate.getMonth(),
			    day = eventDate.getDate(),
			    hour = eventDate.getHours(),
			    block = eventDate.getMinutes() / 15,
			    dayStr = year + "_" + month + "_" + day,
			    blockStr = e.DriverID + "_" + dayStr + "_" + hour + "_" + block,
			    eventDiv = createElement("div"),
			    eventMover = eventDiv.appendChild(createElement("div")),
			    eventCell, left, width;
			if (typeof days[dayStr] === "undefined") {
				return;
			}
			if (e.DriverID === 0) {
				eventCell = days[dayStr][2].getElementById("cell_" + blockStr);
			} else {
				eventCell = eventDiv.appendChild(days[dayStr][1].getElementById("cell_" + blockStr));
			}
			var left = eventCell.style.left,
			    width = (e.End - e.Start) / 60000;
			eventDiv.setAttribute("class", "event");
			eventDiv.addEventListener("click", showEvent.bind(null, e));
			eventDiv.style.left = left;
			eventDiv.setAttribute("timeWidth", e.End - e.Start);
			eventDiv.setAttribute("eventID", e.ID);
			eventMover.setAttribute("class", "eventMover");
			eventMover.addEventListener("click", moveEvent.bind(null, eventDiv));
			if (e.DriverID === 0) {
				var blockTop = 0,
				    i = 0;
				Loop:
				while (true) {
					blockTop += 100;
					for (; i < unassignedEvents.length; i++) {
						if (unassignedEvents[i].Top === blockTop && Math.max(unassignedEvents[i].Start, e.Start) < Math.min(unassignedEvents[i].End, e.End)) {
							continue Loop;
						}
					}
					break;
				}
				e.Top = blockTop;
				unassignedEvents.push(e);
				eventDiv.style.top = blockTop + "px";
			} else {
				eventDiv.style.top = eventCell.style.top;
			}
			eventDiv.style.width = width + "px";
			eventDiv.setAttribute("id", "event_" + blockStr);
			rpc.getCompanyColourFromClient(e.ClientID, function(colour) {
				eventDiv.style.backgroundColor = colour.formatColour();
			});
			rpc.getEventNote(e.ID, function(noteText) {
				//noteText = noteJSON(noteText).Note;
				rpc.getClient(e.ClientID, function(c) {
					var name = eventDiv.appendChild(createElement("div")).setInnerText(c.Name),
					    from = eventDiv.appendChild(createElement("div")).setInnerText(e.From),
					    to = eventDiv.appendChild(createElement("div")).setInnerText(e.To),
					    startText = (new Date(e.Start)).toLocaleString(),
					    endText = (new Date(e.End)).toLocaleString(),
					    //startEndText = startText + " -> " + endText,
					    //startEnd = createElement("div").setInnerText(startEndText),
					    notes = createElement("div").setInnerText(noteText),
					    start = createElement("div").setInnerText(startText),
					    end = createElement("div").setInnerText(endText),
					    noteWidth = noteText.getWidth("14px Serif"),
					    nameWidth = c.Name.getWidth("14px Serif"),
					    fromWidth = e.From.getWidth("14px Serif"),
					    toWidth = e.To.getWidth("14px Serif"),
					    //startEndWidth = startEndText.getWidth("14px Serif"),
					    startWidth = startText.getWidth("14px Serif"),
					    endWidth = endText.getWidth("14px Serif"),
					    maxWidth = nameWidth;
					notes.setAttribute("class", "eventNote");
					//startEnd.setAttribute("class", "time");
					start.setAttribute("class", "time");
					end.setAttribute("class", "time");
					name.style.width = nameWidth + "px";
					from.style.width = fromWidth + "px";
					if (noteWidth > maxWidth) {
						maxWidth = noteWidth;
					}
					to.style.width = toWidth + "px";
					if (fromWidth > maxWidth) {
						maxWidth = fromWidth;
					}
					if (toWidth > maxWidth) {
						maxWidth = toWidth;
					}
					eventDiv.setAttribute("class", "event expandable");
					eventDiv.appendChild(notes);
					//eventDiv.appendChild(startEnd);
					eventDiv.appendChild(start);
					eventDiv.appendChild(end);
					eventDiv.appendChild(notes);
					//if (startEndWidth > maxWidth) {
					//	maxWidth = startEndWidth;
					//}
					if (startWidth > maxWidth) {
						maxWidth = startWidth;
					}
					if (endWidth > maxWidth) {
						maxWidth = endWidth;
					}
					var newLeft = Math.floor(parseInt(left) - (((maxWidth + 12) - width) / 2));
					// 1px left border + 5px left padding + 5px right padding + 1px right border
					if (maxWidth + 12 > parseInt(width)) {
						eventDiv.addEventListener("mouseover", function() {
							name.style.marginLeft = (maxWidth - nameWidth) / 2 + "px";
							from.style.marginLeft = (maxWidth - fromWidth) / 2 + "px";
							to.style.marginLeft = (maxWidth - toWidth) / 2 + "px";
							eventDiv.style.width = notes.style.width = maxWidth + 12 + "px";
							eventDiv.style.left = newLeft + "px";
							eventMover.style.left = parseInt(left) - newLeft + "px";
						});
					} else {
						eventDiv.addEventListener("mouseover", function() {
							name.style.marginLeft = (width - nameWidth) / 2 + "px";
							from.style.marginLeft = (width - fromWidth) / 2 + "px";
							to.style.marginLeft = (width - toWidth) / 2 + "px";
						});
					}
					eventDiv.addEventListener("mouseout", function() {
						name.style.marginLeft = "0";
						from.style.marginLeft = "0";
						to.style.marginLeft = "0";
						eventDiv.style.left = left;
						eventDiv.style.width = notes.style.width = width + "px";
						eventMover.style.left = "0px";

					});
				});
			});
			if (e.DriverID === 0) {
				days[dayStr][2].appendChild(eventDiv);
			} else {
				days[dayStr][1].appendChild(eventDiv);
			}
		};
		this.addDriver = function(d) {
			if (typeof d === "undefined") {
				return;
			}
			drivers[d.ID] = d;
			if (d.Show === false) {
				return;
			}
			drivers[d.ID].yPos = nextDriverPos;
			var dDiv = createElement("div"),
			    t;
			drivers[d.ID].driverDiv = dDiv;
			dDiv.appendChild(createElement("div")).setInnerText(d.Name);
			dDiv.setAttribute("class", "driverName simpleButton");
			dDiv.setAttribute("id", "driver_" + d.ID);
			dDiv.addEventListener("click", function() {
				showDriver({
					"ID": d.ID,
					"Name": d.Name,
					"PhoneNumber": d.PhoneNumber,
					"RegistrationNumber": d.RegistrationNumber,
				});
			});
			dDiv.style.top = nextDriverPos + "px";
			nextDriverPos += 100;
			plusDriver.style.top = nextDriverPos + "px";
			driverEvents.appendChild(dDiv);
			var keys = Object.keys(days),
			    oddEven = (nextDriverPos / 100) % 2;
			for (var i = 0; i < keys.length; i++) {
				var parts = keys[i].split("_"),
				    year = parts[0],
				    month = parts[1],
				    day = parts[2],
				    dayDiv = days[keys[i]];
				dayDiv[2].style.top = nextDriverPos + "px";
				for (var hour = 0; hour < 24; hour++) {
					for (var block = 0; block < 4; block++) {
						var cellDiv = createElement("div"),
						    fifteenDiv = dayDiv[0].getElementById("minute_" + year + "_" + month + "_" + day + "_" + hour + "_" + block);
						cellDiv.setAttribute("class", "eventCell " + (block % 2 !== oddEven ? "cellOdd" : "cellEven"));
						cellDiv.setAttribute("id", "cell_" + d.ID + "_" + year + "_" + month + "_" + day + "_" + hour + "_" + block);
						cellDiv.style.left = timeToPos(new Date(year, month, day, hour, block * 15));
						cellDiv.style.top = drivers[d.ID].yPos + "px";
						cellDiv.addEventListener("mouseover", eventOnMouseOver);
						cellDiv.addEventListener("mouseover", fifteenDiv.setAttribute.bind(fifteenDiv, "class", "minute select"));
						cellDiv.addEventListener("mouseout", eventOnMouseOut);
						cellDiv.addEventListener("mouseout", fifteenDiv.setAttribute.bind(fifteenDiv, "class", "minute"));
						cellDiv.addEventListener("click", eventOnClick);
						dayDiv[1].appendChild(cellDiv);
					}
				}
			}
		};
		this.updateDriver = function(d) {
			d.yPos = drivers[d.ID].yPos;
			d.driverDiv = drivers[d.ID].driverDiv;
			d.driverDiv.getElementsByTagName("div")[0].setInnerText(d.Name);
			drivers[d.ID] = d;
		};
		this.reload = function(key, data) {
			var additional = "";
			if (typeof key === "string") {
				additional = "&" + key + "=" + data;
			}
			window.location.search = "?date="+dateTime.getTime() + additional;
		};
		this.removeDriver = function(d) {
			window.location.search = "?date="+dateTime.getTime();
		};
		this.updateEvent = function(e) {
			if (typeof e === "undefined") {
				return;
			}
		};
		this.removeEvent = function (e) {
			
		};
		this.setTime = function (time) {
			dateTime = time;
			update();
		};
		this.init = function() {
			init.call(this);
		};
	})(),
	makeTabs = function() {
		var frag = document.createDocumentFragment(),
		    tabDiv = frag.appendChild(createElement("div")),
		    contentDiv = frag.appendChild(createElement("div")),
		    tabs = new Array(arguments.length);
		tabDiv.setAttribute("class", "tabs");
		contentDiv.setAttribute("class", "content")
		for (var i = 0; i < arguments.length; i++) {
			tabs[i] = tabDiv.appendChild(createElement("div")).setInnerText(arguments[i][0]);
			tabs[i].addEventListener("click", function(tab, callback) {
				if (tab.getAttribute("class") === "selected") {
					return;
				}
				contentDiv.removeChildren();
				tabs.map(function(tab) {tab.removeAttribute("class")});
				tab.setAttribute("class", "selected");
				var tLayer = layer;
				layer = contentDiv;
				callback();
				layer = tLayer;
			}.bind(null, tabs[i], arguments[i][1]));
		}
		tabs[0].dispatchEvent(new MouseEvent("click", {"view": window, "bubble": false, "cancelable": true}));
		return frag;
	},
	makeNote = function(getter, setter) {
		var note = createElement("textarea");
		note.setAttribute("class", "note");
		getter(function(n) {
			note.value = n;
		});
		note.addEventListener("blur", function() {
			setter(note.value);
		});
		return note;
	},
	makeInvoice = function(company, startDate, endDate, events) {
		stack.addLayer("invoice");
		layer.setAttribute("class", "toPrint");
		stack.addFragment();
		var header = layer.appendChild(createElement("div")),
		    topTable = layer.appendChild(createElement("table")),
		    table = layer.appendChild(createElement("table")),
		    costTable = layer.appendChild(createElement("table")),
		    addressDate, eomDate = new Date(0), eomDateElm, invoiceNo, ref, tableTitles, i = 0, totalParking = 0, totalPrice = 0,
		    subTotal, admin, adminPrice, adminTotal, adminTotalPrice, vat, vatPrice, parking, total, finalTotal, lineOne, lineTwo, adminInput, cn, vatEdit;
		header.setAttribute("class", "printOnly");
		cn = header.appendChild(createElement("div"));
		cn.setAttribute("id", "invoiceCompanyName");
		cn.appendChild(createElement("span")).setInnerText("Academy");
		cn.appendChild(createElement("span")).setInnerText("Chauffeurs");
		header.appendChild(createElement("div")).setPreText("83 The Pastures\nStevenage\nHerts SG2 7DF\nVAT No 830 8068 35").setAttribute("id", "invoiceCompanyAddress");
		topTable.setAttribute("class", "invoiceTop");
		topTable.appendChild(createElement("tr")).appendChild(createElement("td")).setInnerText("Invoice to:").setAttribute("colspan", "3");
		addressDate = topTable.appendChild(createElement("tr"));
		addressDate.appendChild(createElement("td")).setPreText(company.Name + "\n" + company.Address).setAttribute("rowspan", "3");
		addressDate.appendChild(createElement("td")).setInnerText("Date :");
		eomDateElm = addressDate.appendChild(createElement("td"))
		invoiceNo = topTable.appendChild(createElement("tr"));
		invoiceNo.appendChild(createElement("td")).setInnerText("Invoice No:");
		invoiceNo.appendChild(createElement("td")).setAttribute("contenteditable", "true");
		ref = topTable.appendChild(createElement("tr"));
		ref.appendChild(createElement("td")).setInnerText("Your Ref:");
		ref.appendChild(createElement("td")).setAttribute("contenteditable", "true");
		table.setAttribute("class", "invoice");
		tableTitles = table.appendChild(createElement("tr"));
		tableTitles.appendChild(createElement("th")).setInnerText("Date");
		tableTitles.appendChild(createElement("th")).setInnerText("Name").setAttribute("colspan", 2);
		tableTitles.appendChild(createElement("th")).setInnerText("Details");
		tableTitles.appendChild(createElement("th")).setInnerText("Additional").setAttribute("contenteditable", "true");
		tableTitles.appendChild(createElement("th")).setInnerText("Parking").setAttribute("colspan", "2");
		tableTitles.appendChild(createElement("th")).setInnerText("").setAttribute("colspan", "2");
		for (; i < events.length; i++) {
			var row = table.appendChild(createElement("tr")), cr, details, extra, thisDate = new Date(events[i].Start);
			if (thisDate.getTime() > eomDate.getTime()) {
				eomDate = thisDate;
			}
			row.appendChild(createElement("td")).setInnerText(thisDate.toDateString());
			cr = row.appendChild(createElement("td"));
			cr.setInnerText(events[i].ClientReference).setAttribute("contenteditable", "true");
			row.appendChild(createElement("td")).setInnerText(events[i].ClientName).setAttribute("contenteditable", "true");
			details = row.appendChild(createElement("td"));
			details.setAttribute("contenteditable", "true");
			extra = row.appendChild(createElement("td"));
			if (events[i].Waiting !== 0) {
				extra.setInnerText(events[i].Waiting + " mins waiting");
			}
			extra.setAttribute("contenteditable", "true");
			cr.setInnerText(events[i].ClientRef);
			extra.setPreText(extra.innerText + "\n" + events[i].InvoiceNote);
			if (events[i].InvoiceFrom === "") {
				details.appendChild(document.createTextNode("From: " + events[i].From));
			} else {
				details.appendChild(document.createTextNode("From: " + events[i].InvoiceFrom));
			}
			details.appendChild(createElement("br"));
			if (events[i].InvoiceTo === "") {
				details.appendChild(document.createTextNode("To: " + events[i].To));
			} else {
				details.appendChild(document.createTextNode("To: " + events[i].InvoiceTo));
			}
			/*
			rpc.getEventNote(events[i].ID, function(cr, details, extra, noteText) {
				var data = noteJSON(noteText);
				if (typeof data.ClientRef !== "undefined") {
					cr.setInnerText(data.ClientRef);
				}
				if (typeof data.InvoiceNote !== "undefined") {
					extra.setPreText(extra.innerText + "\n" + data.InvoiceNote);
				}
				if (typeof data.InvoiceFrom === "undefined") {
					details.appendChild(document.createTextNode("From: " + events[i].From));
				} else {
					details.appendChild(document.createTextNode("From: " + data.InvoiceFrom));
				}
				details.appendChild(createElement("br"));
				if (typeof data.InvoiceTo === "undefined") {
					details.appendChild(document.createTextNode("To: " + events[i].To));
				} else {
					details.appendChild(document.createTextNode("To: " + data.InvoiceTo));
				}
			}.bind(null, cr, details, extra));
			*/
			row.appendChild(createElement("td")).setInnerText("£");
			row.appendChild(createElement("td")).setInnerText((0.01 * events[i].Parking).formatMoney());
			row.appendChild(createElement("td")).setInnerText("£");
			row.appendChild(createElement("td")).setInnerText((0.01 * events[i].Price).formatMoney());
			totalParking += events[i].Parking;
			totalPrice += events[i].Price;
		}
		eomDateElm.setInnerText((new Date(eomDate.getFullYear(), eomDate.getMonth()-1, eomDate.daysInMonth())).toOrdinalDate()).setAttribute("contenteditable", "true");
		costTable.setAttribute("class", "invoiceBottom");
		costTable.setAttribute("id", "invoiceBottom");
		subTotal = costTable.appendChild(createElement("tr"));
		subTotal.setAttribute("class", "totals");
		subTotal.appendChild(createElement("td"));
		subTotal.appendChild(createElement("td")).setInnerText("Sub Total");
		subTotal.appendChild(createElement("td")).setInnerText("£");
		subTotal.appendChild(createElement("td")).setInnerText((totalPrice / 100).formatMoney());
		admin = costTable.appendChild(createElement("tr"));
		adminPrice = totalPrice * adminPercent / 100;
		admin.appendChild(createElement("td"));
		admin.appendChild(createElement("td")).setInnerText("Account Admin");
		admin.appendChild(createElement("td")).setInnerText("£");
		adminInput = admin.appendChild(createElement("td")).setInnerText((adminPrice / 100).formatMoney());
		adminInput.setAttribute("contenteditable", "true");
		adminInput.addEventListener("blur", function() {
			var value = parseFloat(adminInput.textContent);
			adminInput.setInnerText(value.formatMoney());
			value *= 100;
			adminTotalPrice = totalPrice + value;
			vatPrice = adminTotalPrice * vatPercent / 100;
			finalTotal = adminTotalPrice + vatPrice + totalParking;
			adminTotal.lastChild.setInnerText((adminTotalPrice / 100).formatMoney());
			vat.lastChild.setInnerText((vatPrice / 100).formatMoney());
			total.lastChild.setInnerText((finalTotal / 100).formatMoney());
		});
		lineOne = costTable.appendChild(createElement("tr"));
		lineOne.setAttribute("class", "line");
		lineOne.appendChild(createElement("td"));
		lineOne.appendChild(createElement("td")).setAttribute("colspan", "4");
		adminTotal = costTable.appendChild(createElement("tr"));
		adminTotalPrice = totalPrice + adminPrice;
		adminTotal.appendChild(createElement("td"));
		adminTotal.appendChild(createElement("td")).setInnerText("");
		adminTotal.appendChild(createElement("td")).setInnerText("£");
		adminTotal.appendChild(createElement("td")).setInnerText((adminTotalPrice / 100).formatMoney());
		vat = costTable.appendChild(createElement("tr"));
		vat.appendChild(createElement("td"));
		vatPrice = adminTotalPrice * vatPercent / 100;
		vat.appendChild(createElement("td")).setInnerText("Plus VAT @ " + vatPercent + "%");
		vat.appendChild(createElement("td")).setInnerText("£");
		vatEdit = vat.appendChild(createElement("td")).setInnerText((vatPrice / 100).formatMoney());
		vatEdit.setAttribute("contenteditable", "true");
		vatEdit.addEventListener("blur", function() {
			vatPrice = parseFloat(vatEdit.textContent) * 100;
			finalTotal = adminTotalPrice + vatPrice + totalParking;
			total.lastChild.setInnerText((finalTotal / 100).formatMoney());
		});
		parking = costTable.appendChild(createElement("tr"));
		parking.appendChild(createElement("td"));
		parking.appendChild(createElement("td")).setInnerText("Parking");
		parking.appendChild(createElement("td")).setInnerText("£");
		parking.appendChild(createElement("td")).setInnerText((totalParking / 100).formatMoney());
		lineTwo = costTable.appendChild(createElement("tr"));
		lineTwo.setAttribute("class", "doubleLine");
		lineTwo.appendChild(createElement("td"));
		lineTwo.appendChild(createElement("td")).setAttribute("colspan", "4");
		total = costTable.appendChild(createElement("tr"));
		finalTotal = adminTotalPrice + vatPrice + totalParking;
		total.appendChild(createElement("td"));
		total.appendChild(createElement("td")).setInnerText("Total");
		total.appendChild(createElement("td")).setInnerText("£");
		total.appendChild(createElement("td")).setInnerText((finalTotal / 100).formatMoney());
		stack.setFragment();
	},
	showCompany = function(company) {
		stack.addFragment();
		layer.appendChild(createElement("h1")).setInnerText(company.Name);
		layer.appendChild(makeTabs(
			[ "Details", function() {
				var toPrint = layer.appendChild(createElement("div"));
				toPrint.setAttribute("class", "toPrint");
				toPrint.appendChild(createElement("h2")).setInnerText("Company Details").setAttribute("class", "printOnly");
				toPrint.appendChild(createElement("label")).setInnerText("Company Name");
				toPrint.appendChild(createElement("div")).setInnerText(company.Name);
				toPrint.appendChild(createElement("label")).setInnerText("Company Address");
				toPrint.appendChild(createElement("div")).setInnerText(company.Address);
				toPrint.appendChild(createElement("label")).setInnerText("No. of Clients");
				var numClients = toPrint.appendChild(createElement("div")).setInnerText("-"),
				    numEvents = createElement("div").setInnerText("-");
				toPrint.appendChild(createElement("label")).setInnerText("No. of Events");
				toPrint.appendChild(numEvents);
				toPrint.appendChild(createElement("label")).setInnerText("Notes");
				toPrint.appendChild(makeNote(rpc.getCompanyNote.bind(rpc, company.ID), rpc.setCompanyNote.bind(rpc, company.ID)));
				rpc.getNumClients(company.ID, numClients.setInnerText.bind(numClients));
				rpc.getNumEvents(company.ID, numEvents.setInnerText.bind(numEvents));
			}],
			["Clients", function() {
				var toPrint = layer.appendChild(createElement("div")),
				    printOnly = toPrint.appendChild(createElement("div")),
				    clientsTable = toPrint.appendChild(createElement("table")),
				    headerRow = clientsTable.appendChild(createElement("tr")),
				    i = 0,
				    exportButton = layer.appendChild(createElement("form"));
				exportButton.setAttribute("method", "post");
				exportButton.setAttribute("action", "/export");
				exportButton.setAttribute("target", "_new");
				exportButton.setAttribute("class", "noPrint");
				toPrint.setAttribute("class", "toPrint");
				printOnly.setAttribute("class", "printOnly");
				printOnly.appendChild(createElement("h1")).setInnerText("Clients for " + company.Name);
				headerRow.appendChild(createElement("th")).setInnerText("Name");
				headerRow.appendChild(createElement("th")).setInnerText("Phone Number");
				headerRow.appendChild(createElement("th")).setInnerText("Reference");
				headerRow.appendChild(createElement("th")).setInnerText("# Events");
				rpc.clientsForCompany(company.ID, function(clients) {
					exportButton.removeChildren();
					if (clients.length === 0) {
						clientsTable.appendChild(createElement("tr")).appendChild(createElement("td")).setInnerText("No Clients").setAttribute("colspan", 4);
						return;
					}
					makeExportButton(exportButton, "companyClients", company.ID);
					for (; i < clients.length; i++) {
						var row = clientsTable.appendChild(createElement("tr")),
						    name = row.appendChild(createElement("td")).setInnerText(clients[i].Name),
						    numEvents;
						row.appendChild(createElement("td")).setInnerText(clients[i].PhoneNumber);
						row.appendChild(createElement("td")).setInnerText(clients[i].Reference);
						numEvents = row.appendChild(createElement("td")).setInnerText("-");
						rpc.getNumEventsClient(clients[i].ID, numEvents.setInnerText.bind(numEvents));
					}
				});
			}],
			[ "Events", function() {
				var eventsStartDate = new Date(),
				    eventsEndDate = new Date();
				return function() {
					var startDate = addFormElement("Start Date", "text", "startDate", eventsStartDate.toDateString(), dateCheck),
					    endDate = addFormElement("End Date", "text", "endDate", eventsEndDate.toDateString(), dateCheck),
					    getEvents = addFormSubmit("Show Events", function() {
						eventTable.removeChildren(function(elm) {
							return elm !== tableTitles;
						});
						while (eventTable.nextSibling !== null) {
							eventTable.parentNode.removeChild(eventTable.nextSibling);
						}
						var startParts = startDate[0].value.split("/"),
						    endParts = endDate[0].value.split("/"),
						    pT = "";
						eventsStartDate = new Date(startParts[2], startParts[1]-1, startParts[0]);
						eventsEndDate = new Date(endParts[2], endParts[1]-1, endParts[0]);
						pT = "Events for " + company.Name + " for " + eventsStartDate.toDateString();
						if (eventsStartDate.getTime() !== eventsEndDate.getTime()) {
							pT += " to " + eventsEndDate.toDateString();
						}
						printTitle.setInnerText(pT);
						if (eventsStartDate.getTime() > eventsEndDate.getTime()) {
							endDate[1].setInnerText("Cannot be before start date");
							eventTable.appendChild(createElement("tr")).appendChild(createElement("td")).setInnerText("No Events").setAttribute("colspan", "5");
							return;
						}
						rpc.getEventsWithCompany(company.ID, eventsStartDate.getTime(), eventsEndDate.getTime() + (24 * 3600 * 1000), function(events) {
							exportButton.removeChildren();
							if (events.length === 0) {
								eventTable.appendChild(createElement("tr")).appendChild(createElement("td")).setInnerText("No Events").setAttribute("colspan", "9");
								return;
							}
							makeExportButton(exportButton, "companyEvents", company.ID, eventsStartDate, eventsEndDate);
							var loading = new waitGroup(function() {
								var invoiceButton = createElement("input");
								invoiceButton.setAttribute("class", "noPrint");
								invoiceButton.setAttribute("type", "button");
								invoiceButton.value = "Make Invoice";
								invoiceButton.addEventListener("click", function() {
									makeInvoice(company, eventsStartDate, eventsEndDate, events);
								});
								eventTable.parentNode.appendChild(invoiceButton);
							    }),
							    row, i = 0,
							    totalParking = 0, totalCost = 0,
							    wg = new waitGroup(function() {
								var row = createElement("tr");
								row.appendChild(createElement("td")).setInnerText(events.length + " events").setAttribute("colspan", "7");
								row.appendChild(createElement("td")).setInnerText("£" + (totalParking / 100).formatMoney());
								row.appendChild(createElement("td")).setInnerText("£" + (totalCost / 100).formatMoney());
								eventTable.appendChild(row).setAttribute("class", "overline");
							    });
							for (; i < events.length; i++) {
								row = createElement("tr");
								row.appendChild(createElement("td")).setInnerText(new Date(events[i].Start).toLocaleString());
								row.appendChild(createElement("td")).setInnerText(new Date(events[i].End).toLocaleString());
								var clientCell = row.appendChild(createElement("td")),
								    refCell = row.appendChild(createElement("td")),
								    driverCell = createElement("td").setInnerText("-"),
								    parkingCell = createElement("td").setInnerText("-"),
								    priceCell = createElement("td").setInnerText("-");
								row.appendChild(createElement("td")).setInnerText(events[i].From);
								row.appendChild(createElement("td")).setInnerText(events[i].To);
								row.appendChild(driverCell);
								row.appendChild(parkingCell);
								row.appendChild(priceCell);
								loading.add();
								rpc.getClient(events[i].ClientID, function(clientCell, refCell, i, client) {
									loading.done();
									events[i].ClientReference = client.Reference;
									events[i].ClientName = client.Name;
									clientCell.setInnerText(client.Name);
									refCell.setInnerText(client.Reference);
								}.bind(null, clientCell, refCell, i));
								if (events[i].DriverID === 0) {
									events[i].DriverName = "Unassigned";
									driverCell.setInnerText("Unassigned");
								} else {
									loading.add();
									rpc.getDriver(events[i].DriverID, function(driverCell, i, driver) {
										loading.done();
										events[i].DriverName = driver.Name;
										driverCell.setInnerText(driver.Name);
									}.bind(null, driverCell, i));
								}
								loading.add();
								wg.add();
								rpc.getEventFinals(events[i].ID, function(parkingCell, priceCell, i, eventFinals) {
									if (eventFinals.FinalsSet) {
										loading.done();
										parkingCell.setInnerText("£" + (eventFinals.Parking / 100).formatMoney());
										priceCell.setInnerText("£" + (eventFinals.Price / 100).formatMoney());
										events[i].Parking = eventFinals.Parking;
										events[i].Price = eventFinals.Price;
										events[i].Waiting = eventFinals.Waiting;
										totalParking += eventFinals.Parking;
										totalCost += eventFinals.Price;
									}
									wg.done();
								}.bind(null, parkingCell, priceCell, i));
								eventTable.appendChild(row);
							}
						});
					    }),
					    toPrint = layer.appendChild(createElement("div")),
					    printTitle = toPrint.appendChild(createElement("h2")),
					    eventFormTable = toPrint.appendChild(createElement("table")),
					    eventTable = eventFormTable.appendChild(createElement("table")),
					    tableTitles = eventTable.appendChild(createElement("tr")),
					    exportButton = layer.appendChild(createElement("form"));
					exportButton.setAttribute("method", "post");
					exportButton.setAttribute("action", "/export");
					exportButton.setAttribute("target", "_new");
					toPrint.setAttribute("class", "toPrint");
					printTitle.setAttribute("class", "printOnly");
					printTitle.setInnerText("Events for " + company.Name);
					tableTitles.appendChild(createElement("th")).setInnerText("Start");
					tableTitles.appendChild(createElement("th")).setInnerText("End");
					tableTitles.appendChild(createElement("th")).setInnerText("Client");
					tableTitles.appendChild(createElement("th")).setInnerText("Reference");
					tableTitles.appendChild(createElement("th")).setInnerText("From");
					tableTitles.appendChild(createElement("th")).setInnerText("To");
					tableTitles.appendChild(createElement("th")).setInnerText("Driver");
					tableTitles.appendChild(createElement("th")).setInnerText("Parking Cost");
					tableTitles.appendChild(createElement("th")).setInnerText("Price");
					getEvents.dispatchEvent(new MouseEvent("click", {"view": window, "bubble": false, "cancelable": true}));
				};
			}()],
			[ "Options", function() {
				var edit = layer.appendChild(createElement("div")).setInnerText("Edit Company"),
				    deleter = layer.appendChild(createElement("div")).setInnerText("Delete Company");
				edit.setAttribute("class", "simpleButton");
				edit.addEventListener("click", function() {
					stack.addLayer("editCompany", function(c) {
						if (typeof c !== "undefined") {
							events.reload("company", c.ID);
						}
					});
					setCompany(company);
				});
				deleter.setAttribute("class", "simpleButton");
				deleter.addEventListener("click", function() {
					if(confirm("Are you sure you want to remove this company?\n\nNB: This will also remove all clients and events attached to this company.")) {
						rpc.removeCompany(company.ID);
						events.reload("company", 0);
					}
				});
			}]
		))
		stack.setFragment();
	},
	companyList = function(addList) {
		rpc.companies(function(companies) {
			stack.addFragment();
			layer.appendChild(createElement("h1")).setInnerText("Companies")
			var table = createElement("table"),
			    headerRow = table.appendChild(createElement("tr")),
			    overview = createElement("div"),
			    overviewButton = overview.appendChild(createElement("input")),
			    doOverview = overview.appendChild(createElement("input")),
			    cancelOverview = overview.appendChild(createElement("input")),
			    checks = [],
			    addCompanyToTable = function(company) {
				if (typeof company === "undefined") {
					return;
				}
				var row = createElement("tr"),
				    nameCell = row.appendChild(createElement("td")).appendChild(createElement("div"));
				nameCell.setInnerText(company.Name);
				if (addList === true) {
					addLister(nameCell, stack.removeLayer.bind(null, company));
				} else {
					var check = createElement("input");
					check.setAttribute("type", "checkbox");
					check.setAttribute("checked", "true");
					nameCell.parentNode.insertBefore(check, nameCell);
					checks.push(check);
					nameCell.setAttribute("class", "simpleButton");
					nameCell.style.backgroundColor = company.Colour.formatColour();
					nameCell.addEventListener("click", function() {
						stack.addLayer("showCompany", function(c) {
							if (typeof c !== "undefined") {
								stack.removeLayer();
								stack.addLayer("companyList");
								companyList();
							}
						});
						showCompany(company);
					});
				}
				row.appendChild(createElement("td")).setInnerText(company.Address);
				table.appendChild(row);
			    },
			    exportButton = createElement("form");
			exportButton.setAttribute("method", "post");
			exportButton.setAttribute("action", "/export");
			exportButton.setAttribute("target", "_new");
			exportButton.setAttribute("class", "noPrint");
			overview.setAttribute("id", "overview");
			overviewButton.setAttribute("type", "button");
			overviewButton.setAttribute("value", "Overview");
			overviewButton.addEventListener("click", function() {
				overview.setAttribute("class", "check");
				table.setAttribute("class", "toPrint check")
			});
			doOverview.setAttribute("type", "button");
			doOverview.setAttribute("value", "✔");
			doOverview.addEventListener("click", function() {
				overview.setAttribute("class", "");
				table.setAttribute("class", "toPrint")
				var list = [],
				    i = 0;
				for (; i < checks.length; i++) {
					if (checks[i].checked) {
						list[companies[i].ID] = companies[i];
					}
				}
				if (list.length === 1) {
					stack.addLayer("showCompany", function(c) {
						if (typeof c !== "undefined") {
							stack.removeLayer();
							stack.addLayer("companyList");
							companyList();
						}
					});
					showCompany(list[0]);
				} else if (list.length > 1) {
					stack.addLayer("showOverview");
					showOverview(list);
				}
			});
			cancelOverview.setAttribute("type", "button");
			cancelOverview.setAttribute("value", "✗");
			cancelOverview.addEventListener("click", function () {
				overview.setAttribute("class", "");
				table.setAttribute("class", "toPrint")
			});
			if (companies.length > 0) {
				makeExportButton(exportButton, "companyList");
			}
			addAdder(null, function() {
				stack.addLayer("addCompany", addCompanyToTable);
				addCompany();
			});
			headerRow.appendChild(createElement("th")).setInnerText("Company Name");
			headerRow.appendChild(createElement("th")).setInnerText("Address");
			companies.map(addCompanyToTable);
			table.setAttribute("class", "toPrint");
			layer.appendChild(table);
			layer.appendChild(overview);
			layer.appendChild(exportButton);
			stack.setFragment();
			layer.setAttribute("class", "toPrint");
		});
	},
	showOverview = function(companies) {
		stack.addFragment();
		var ids = [];
		companies.map(function(company) {
			ids.push(company.ID);
		});
		layer.appendChild(createElement("h1")).setInnerText("Overview");
		layer.appendChild(makeTabs(
			[ "Details", function() {
				for (var i = 0; i < ids.length; i++) {
					layer.appendChild(createElement("label")).setInnerText("Company " + (i + 1));
					layer.appendChild(createElement("div")).setInnerText(companies[ids[i]].Name);
				}
				layer.appendChild(createElement("label")).setInnerText("Total Clients")
				var totalClients = layer.appendChild(createElement("div")).setInnerText("-"),
				    totalEvents = createElement("div");
				layer.appendChild(createElement("label")).setInnerText("Total Events")
				totalEvents = layer.appendChild(totalEvents.setInnerText("-"));
				rpc.getNumClientsForCompanies(ids, function (total) {
					totalClients.setInnerText(total);
				});
				rpc.getNumEventsCompanies(ids, function (total) {
					totalEvents.setInnerText(total);
				});
			}],
			[ "Clients", function() {
				var toPrint = layer.appendChild(createElement("div")),
				    printOnly = toPrint.appendChild(createElement("div")),
				    clientsTable = toPrint.appendChild(createElement("table")),
				    headerRow = clientsTable.appendChild(createElement("tr")),
				    i = 0,
				    exportButton = layer.appendChild(createElement("form"));
				exportButton.setAttribute("method", "post");
				exportButton.setAttribute("action", "/export");
				exportButton.setAttribute("target", "_new");
				exportButton.setAttribute("class", "noPrint");
				toPrint.setAttribute("class", "toPrint");
				printOnly.setAttribute("class", "printOnly");
				printOnly.appendChild(createElement("h1")).setInnerText("Client Overview");
				headerRow.appendChild(createElement("th")).setInnerText("Name");
				headerRow.appendChild(createElement("th")).setInnerText("Company");
				headerRow.appendChild(createElement("th")).setInnerText("Phone Number");
				headerRow.appendChild(createElement("th")).setInnerText("Reference");
				headerRow.appendChild(createElement("th")).setInnerText("# Events");
				rpc.clientsForCompanies(ids, function(clients) {
					exportButton.removeChildren();
					if (clients.length === 0) {
						clientsTable.appendChild(createElement("tr")).appendChild(createElement("td")).setInnerText("No Clients").setAttribute("colspan", 4);
						return;
					}
					for (; i < clients.length; i++) {
						var row = clientsTable.appendChild(createElement("tr")),
						    name = row.appendChild(createElement("td")).setInnerText(clients[i].Name),
						    company = row.appendChild(createElement("td")).setInnerText(companies[clients[i].CompanyID].Name),
						    numEvents;
						company.style.backgroundColor = companies[clients[i].CompanyID].Colour.formatColour();
						row.appendChild(createElement("td")).setInnerText(clients[i].PhoneNumber);
						row.appendChild(createElement("td")).setInnerText(clients[i].Reference);
						numEvents = row.appendChild(createElement("td")).setInnerText("-");
						rpc.getNumEventsClient(clients[i].ID, numEvents.setInnerText.bind(numEvents));
					}
				});
			}],
			[ "Events", function() {
				var eventsStartDate = new Date(),
				    eventsEndDate = new Date();
				return function() {
					var startDate = addFormElement("Start Date", "text", "startDate", eventsStartDate.toDateString(), dateCheck),
					    endDate = addFormElement("End Date", "text", "endDate", eventsEndDate.toDateString(), dateCheck),
					    getEvents = addFormSubmit("Show Events", function() {
						eventTable.removeChildren(function(elm) {
							return elm !== tableTitles;
						});
						while (eventTable.nextSibling !== null) {
							eventTable.parentNode.removeChild(eventTable.nextSibling);
						}
						var startParts = startDate[0].value.split("/"),
						    endParts = endDate[0].value.split("/"),
						    pT = "";
						eventsStartDate = new Date(startParts[2], startParts[1]-1, startParts[0]);
						eventsEndDate = new Date(endParts[2], endParts[1]-1, endParts[0]);
						pT = "Events Overview for " + eventsStartDate.toDateString();
						if (eventsStartDate.getTime() !== eventsEndDate.getTime()) {
							pT += " to " + eventsEndDate.toDateString();
						}
						printTitle.setInnerText(pT);
						if (eventsStartDate.getTime() > eventsEndDate.getTime()) {
							endDate[1].setInnerText("Cannot be before start date");
							eventTable.appendChild(createElement("tr")).appendChild(createElement("td")).setInnerText("No Events").setAttribute("colspan", "5");
							return;
						}
						rpc.getEventsWithCompanies(ids, eventsStartDate.getTime(), eventsEndDate.getTime() + (24 * 3600 * 1000), function(events) {
							exportButton.removeChildren();
							if (events.length === 0) {
								eventTable.appendChild(createElement("tr")).appendChild(createElement("td")).setInnerText("No Events").setAttribute("colspan", "9");
								return;
							}
							makeExportButton(exportButton, "overview", ids, eventsStartDate, eventsEndDate);
							var row, i = 0,
							    totalParking = 0, totalCost = 0,
							    wg = new waitGroup(function() {
								var row = createElement("tr");
								row.appendChild(createElement("td")).setInnerText(events.length + " events").setAttribute("colspan", "8");
								row.appendChild(createElement("td")).setInnerText("£" + (totalParking / 100).formatMoney());
								row.appendChild(createElement("td")).setInnerText("£" + (totalCost / 100).formatMoney());
								eventTable.appendChild(row).setAttribute("class", "overline");
							    });
							for (; i < events.length; i++) {
								row = createElement("tr");
								row.appendChild(createElement("td")).setInnerText(new Date(events[i].Start).toLocaleString());
								row.appendChild(createElement("td")).setInnerText(new Date(events[i].End).toLocaleString());
								var clientCell = row.appendChild(createElement("td")),
								    companyCell = row.appendChild(createElement("td")),
								    refCell = row.appendChild(createElement("td")),
								    driverCell = createElement("td").setInnerText("-"),
								    parkingCell = createElement("td").setInnerText("-"),
								    priceCell = createElement("td").setInnerText("-");
								row.appendChild(createElement("td")).setInnerText(events[i].From);
								row.appendChild(createElement("td")).setInnerText(events[i].To);
								row.appendChild(driverCell);
								row.appendChild(parkingCell);
								row.appendChild(priceCell);
								rpc.getClient(events[i].ClientID, function(clientCell, refCell, companyCell, i, client) {
									events[i].ClientReference = client.Reference;
									events[i].ClientName = client.Name;
									clientCell.setInnerText(client.Name);
									refCell.setInnerText(client.Reference);
									companyCell.setInnerText(companies[client.CompanyID].Name).style.backgroundColor = companies[client.CompanyID].Colour.formatColour();
								}.bind(null, clientCell, refCell, companyCell, i));
								if (events[i].DriverID === 0) {
									events[i].DriverName = "Unassigned";
									driverCell.setInnerText("Unassigned");
								} else {
									rpc.getDriver(events[i].DriverID, function(driverCell, i, driver) {
										events[i].DriverName = driver.Name;
										driverCell.setInnerText(driver.Name);
									}.bind(null, driverCell, i));
								}
								wg.add();
								rpc.getEventFinals(events[i].ID, function(parkingCell, priceCell, i, eventFinals) {
									if (eventFinals.FinalsSet) {
										parkingCell.setInnerText("£" + (eventFinals.Parking / 100).formatMoney());
										priceCell.setInnerText("£" + (eventFinals.Price / 100).formatMoney());
										events[i].Parking = eventFinals.Parking;
										events[i].Price = eventFinals.Price;
										totalParking += eventFinals.Parking;
										totalCost += eventFinals.Price;
									}
									wg.done();
								}.bind(null, parkingCell, priceCell, i));
								eventTable.appendChild(row);
							}
						});
					    }),
					    toPrint = layer.appendChild(createElement("div")),
					    printTitle = toPrint.appendChild(createElement("h2")),
					    eventFormTable = toPrint.appendChild(createElement("table")),
					    eventTable = eventFormTable.appendChild(createElement("table")),
					    tableTitles = eventTable.appendChild(createElement("tr")),
					    exportButton = layer.appendChild(createElement("form"));
					exportButton.setAttribute("method", "post");
					exportButton.setAttribute("action", "/export");
					exportButton.setAttribute("target", "_new");
					toPrint.setAttribute("class", "toPrint");
					printTitle.setAttribute("class", "printOnly");
					printTitle.setInnerText("Event Overview");
					tableTitles.appendChild(createElement("th")).setInnerText("Start");
					tableTitles.appendChild(createElement("th")).setInnerText("End");
					tableTitles.appendChild(createElement("th")).setInnerText("Client");
					tableTitles.appendChild(createElement("th")).setInnerText("Company");
					tableTitles.appendChild(createElement("th")).setInnerText("Reference");
					tableTitles.appendChild(createElement("th")).setInnerText("From");
					tableTitles.appendChild(createElement("th")).setInnerText("To");
					tableTitles.appendChild(createElement("th")).setInnerText("Driver");
					tableTitles.appendChild(createElement("th")).setInnerText("Parking Cost");
					tableTitles.appendChild(createElement("th")).setInnerText("Price");
					getEvents.dispatchEvent(new MouseEvent("click", {"view": window, "bubble": false, "cancelable": true}));
				};
			}()]
		));
		stack.setFragment();
	},
	showClient = function(client) {
		stack.addLayer("showClient");
		stack.addFragment();
		layer.appendChild(createElement("h1")).setInnerText(client.Name);
		layer.appendChild(makeTabs(
			[ "Details", function() {
				var toPrint = layer.appendChild(createElement("div"));
				toPrint.setAttribute("class", "toPrint");
				toPrint.appendChild(createElement("h2")).setInnerText("Client Details").setAttribute("class", "printOnly");
				toPrint.appendChild(createElement("label")).setInnerText("Name");
				toPrint.appendChild(createElement("div")).setInnerText(client.Name);
				toPrint.appendChild(createElement("label")).setInnerText("Phone Number");
				if (client.PhoneNumber === "" || client.PhoneNumber === " ") {
					client.PhoneNumber = "-";
				}
				toPrint.appendChild(createElement("div")).setInnerText(client.PhoneNumber);
				toPrint.appendChild(createElement("label")).setInnerText("Reference");
				if (client.Reference === "" || client.Reference === " ") {
					client.Reference = "-";
				}
				toPrint.appendChild(createElement("div")).setInnerText(client.Reference);
				toPrint.appendChild(createElement("label")).setInnerText("Company Name");
				toPrint.appendChild(createElement("div")).setInnerText(client.CompanyName);
				toPrint.appendChild(createElement("label")).setInnerText("No. of Events");
				var bookings = toPrint.appendChild(createElement("div")).setInnerText("-");
				rpc.getNumEventsClient(client.ID, bookings.setInnerText.bind(bookings));
				toPrint.appendChild(createElement("label")).setInnerText("Notes");
				toPrint.appendChild(makeNote(rpc.getClientNote.bind(rpc, client.ID), rpc.setClientNote.bind(rpc, client.ID)));
			}],
			[ "Events", function () {
				var eventsStartDate = new Date(),
				    eventsEndDate = new Date();
				return function() {
					var startDate = addFormElement("Start Date", "text", "startDate", eventsStartDate.toDateString(), dateCheck),
					    endDate = addFormElement("End Date", "text", "endDate", eventsEndDate.toDateString(), dateCheck),
					    getEvents = addFormSubmit("Show Events", function() {
						eventTable.removeChildren(function(elm) {
							return elm !== eventTable;
						});
						var startParts = startDate[0].value.split("/"),
						    endParts = endDate[0].value.split("/"),
						    pT = "";
						eventsStartDate = new Date(startParts[2], startParts[1]-1, startParts[0]);
						eventsEndDate = new Date(endParts[2], endParts[1]-1, endParts[0]);
						pT = "Events for " + client.Name + " for " + eventsStartDate.toDateString();
						if (eventsStartDate.getTime() !== eventsEndDate.getTime()) {
							pT += " to " + eventsEndDate.toDateString();
						}
						printTitle.setInnerText(pT);
						rpc.getEventsWithClient(client.ID, eventsStartDate.getTime(), eventsEndDate.getTime() + (24 * 3600 * 1000), function(events) {
							var row,
							    i = 0,
							    totalWaiting = 0, totalTripTime = 0, totalPrice = 0,
							    wg = new waitGroup(function() {
								var row = createElement("tr");
								row.appendChild(createElement("td")).setInnerText(events.length + " events").setAttribute("colspan", "6");
								row.appendChild(createElement("td")).setInnerText(totalWaiting + " mins");
								row.appendChild(createElement("td")).setInnerText("");
								row.appendChild(createElement("td")).setInnerText((new Date(totalTripTime)).toTimeString());
								row.appendChild(createElement("td")).setInnerText("£" + (totalPrice / 100).formatMoney());
								eventTable.appendChild(row).setAttribute("class", "overline");
							    });
							exportButton.removeChildren();
							if (events.length === 0) {
								eventTable.appendChild(createElement("tr")).appendChild(createElement("td")).setInnerText("No Events").setAttribute("colspan", "10");
								return;
							}
							makeExportButton(exportButton, "clientEvents", client.ID, eventsStartDate, eventsEndDate);
							for (; i < events.length; i++) {
								row = createElement("tr");
								var driverCell = row.appendChild(createElement("td")),
								    inCar = createElement("td").setInnerText("-"),
								    waiting = createElement("td").setInnerText("-"),
								    dropOff = createElement("td").setInnerText("-"),
								    tripTime = createElement("td").setInnerText("-"),
								    price = createElement("td").setInnerText("-");
								row.appendChild(createElement("td")).setInnerText(events[i].From);
								row.appendChild(createElement("td")).setInnerText(events[i].To);
								row.appendChild(createElement("td")).setInnerText(new Date(events[i].Start).toLocaleString());
								row.appendChild(createElement("td")).setInnerText(new Date(events[i].End).toLocaleString());
								row.appendChild(inCar);
								row.appendChild(waiting);
								row.appendChild(dropOff);
								row.appendChild(tripTime);
								row.appendChild(price);
								if (events[i].DriverID === 0) {
									driverCell.setInnerText("Unassigned");
								} else {
									rpc.getDriver(events[i].DriverID, function(driverCell, driver) {
										driverCell.setInnerText(driver.Name);
									}.bind(null, driverCell));
								}
								wg.add();
								rpc.getEventFinals(events[i].ID, function(inCar, waiting, dropOff, tripTime, price, eventFinals) {
									if (eventFinals.FinalsSet) {
										inCar.setInnerText((new Date(eventFinals.InCar)).toTimeString());
										waiting.setInnerText(eventFinals.Waiting + " mins");
										dropOff.setInnerText((new Date(eventFinals.Drop)).toTimeString());
										tripTime.setInnerText((new Date(eventFinals.Trip)).toTimeString());
										price.setInnerText("£" + (eventFinals.Price / 100).formatMoney());
										totalWaiting += eventFinals.Waiting;
										totalTripTime += eventFinals.Trip;
										totalPrice += eventFinals.Price;
									}
									wg.done();
								}.bind(null, inCar, waiting, dropOff, tripTime, price));
								eventTable.appendChild(row);
							}
						});
					    }),
					    toPrint = layer.appendChild(createElement("div")),
					    printTitle = toPrint.appendChild(createElement("h2")),
					    eventTable = toPrint.appendChild(createElement("table")),
					    tableTitles = eventTable.appendChild(createElement("tr")),
					    exportButton = layer.appendChild(createElement("form"));
					exportButton.setAttribute("method", "post");
					exportButton.setAttribute("action", "/export");
					exportButton.setAttribute("target", "_new");
					toPrint.setAttribute("class", "toPrint");
					printTitle.setAttribute("class", "printOnly");
					tableTitles.appendChild(createElement("th")).setInnerText("Driver");
					tableTitles.appendChild(createElement("th")).setInnerText("From");
					tableTitles.appendChild(createElement("th")).setInnerText("To");
					tableTitles.appendChild(createElement("th")).setInnerText("Start");
					tableTitles.appendChild(createElement("th")).setInnerText("End");
					tableTitles.appendChild(createElement("th")).setInnerText("In Car");
					tableTitles.appendChild(createElement("th")).setInnerText("Waiting");
					tableTitles.appendChild(createElement("th")).setInnerText("Drop Off");
					tableTitles.appendChild(createElement("th")).setInnerText("Flight Time");
					tableTitles.appendChild(createElement("th")).setInnerText("Price");
					getEvents.dispatchEvent(new MouseEvent("click", {"view": window, "bubble": false, "cancelable": true}));
				};
			}()],
			[ "Options", function () {
				var edit = layer.appendChild(createElement("div")).setInnerText("Edit Client"),
				    deleter = layer.appendChild(createElement("div")).setInnerText("Delete Client");
				edit.setAttribute("class", "simpleButton");
				edit.addEventListener("click", function() {
					stack.addLayer("editClient", function(c) {
						if (typeof c !== "undefined") {
							events.reload("client", c.ID);
						}
					});
					setClient(client);
				});
				deleter.setAttribute("class", "simpleButton");
				deleter.addEventListener("click", function() {
					if(confirm("Are you sure you want to remove this client?\n\nNB: This will also remove all events attached to this client.")) {
						rpc.removeClient(client.ID);
						events.reload("client", 0);
					}
				});
			}]
		));
		stack.setFragment();
	},
	clientList = function(addList) {
		rpc.clients(function(clients) {
			stack.addFragment()
			layer.appendChild(createElement("h1")).setInnerText("Clients");
			var table = createElement("table"),
			    headerRow = table.appendChild(createElement("tr")),
			    companies = [],
			    addClientToTable = function(client) {
				if (typeof client === "undefined") {
					return;
				}
				var row = createElement("tr"),
				    nameCell = row.appendChild(createElement("td")).appendChild(createElement("div")),
				    companyCell = row.appendChild(createElement("td")),
				    setCompanyCell = function(company) {
					companyCell.setInnerText(companies[client.CompanyID].Name);
					companyCell.style.backgroundColor = company.Colour.formatColour();
					//companyCell.setAttribute("class", "simpleButton");
					//companyCell.addEventListener("click", showCompany.bind(null, companies[client.CompanyID]));
					client.CompanyName = companies[client.CompanyID].Name;
					if (addList === true) {
						addLister(nameCell, stack.removeLayer.bind(null, client));
					} else {
						nameCell.setAttribute("class", "simpleButton");
						nameCell.addEventListener("click", showClient.bind(null, client));
					}
				    };
				nameCell.setInnerText(client.Name);
				if (typeof companies[client.CompanyID] !== "undefined") {
					setCompanyCell();
				} else {
					rpc.getCompany(client.CompanyID, function(company) {
						if (typeof company === "undefined") {
							companyCell.setInnerText("Error!");
							return;
						}
						companies[company.ID] = company;
						setCompanyCell(company);
					});
				}
				row.appendChild(createElement("td")).setInnerText(client.PhoneNumber);
				row.appendChild(createElement("td")).setInnerText(client.Reference);
				table.appendChild(row);
			    },
			    exportButton = createElement("form");
			exportButton.setAttribute("method", "post");
			exportButton.setAttribute("action", "/export");
			exportButton.setAttribute("target", "_new");
			exportButton.setAttribute("class", "noPrint");
			if (clients.length > 0) {
				makeExportButton(exportButton, "clientList");
			}
			addAdder(null, function() {
				stack.addLayer("addClient", addClientToTable);
				addClient();
			});
			headerRow.appendChild(createElement("th")).setInnerText("Client Name");
			headerRow.appendChild(createElement("th")).setInnerText("Company Name");
			headerRow.appendChild(createElement("th")).setInnerText("Phone Number");
			headerRow.appendChild(createElement("th")).setInnerText("Reference");
			clients.map(addClientToTable);
			table.setAttribute("class", "toPrint");
			layer.appendChild(table);
			layer.appendChild(exportButton);
			stack.setFragment();
			layer.setAttribute("class", "toPrint");
		});
	},
	messageList = function() {
		stack.addLayer("messages");
		stack.addFragment();
		layer.appendChild(createElement("h1")).setInnerText("Messages");
		var table = layer.appendChild(createElement("table")),
		    titleRow = table.appendChild(createElement("tr"));
		titleRow.appendChild(createElement("th")).setInnerText("Event Start");
		titleRow.appendChild(createElement("th")).setInnerText("Client Name");
		titleRow.appendChild(createElement("th")).setInnerText("Driver Name");
		rpc.unsentMessages(function(events) {
			if (events.length === 0) {
				table.appendChild(createElement("tr")).appendChild(createElement("td")).setInnerText("No Events").setAttribute("colspan", "3");
				stack.setFragment();
				return;
			}
			for (var i = 0; i < events.length; i++) {
				var row = table.appendChild(createElement("tr")),
				    clientName, driverName;
				row.setAttribute("class", "simpleButton");
				row.appendChild(createElement("td")).setInnerText((new Date(events[i].Start)).toLocaleString());
				clientName = row.appendChild(createElement("td")).setInnerText("-");
				rpc.getClient(events[i].ClientID, function(clientName, eventID, row, client) {
					clientName.setInnerText(client.Name);
					row.addEventListener("click", prepareMessage.bind(null, eventID, client));
				}.bind(null, clientName, events[i].ID, row));
				driverName = row.appendChild(createElement("td")).setInnerText("-");
				rpc.getDriver(events[i].DriverID, function(driverName, driver) {
					driverName.setInnerText(driver.Name);
				}.bind(null, driverName));
			}
			stack.setFragment();
		});
	},
	prepareMessage = function(eventID, client) {
		var textMessageData,
		    emailMessageData,
		    wg = new waitGroup(function() {makeMessage(client, textMessageData, emailMessageData);});
		wg.add(2);
		rpc.prepareMessage(eventID, function(messageData) {
			textMessageData = messageData;
			wg.done();
		});
		rpc.prepareEmail(eventID, function(messageData) {
			emailMessageData = messageData;
			wg.done();
		});
	},
	makeMessage = function(client, messageData, emailData) {
		stack.addLayer("makeMessage", function() {
			stack.removeLayer();
			messageList();
		});
		stack.addFragment();
		layer.appendChild(createElement("h1")).setInnerText("Send Message");
		layer.appendChild(makeTabs(
			[ "Text Message", function() {
				addFormElement("Client Name", "text", "", client.Name);
				addFormElement("Client Number", "text", "", client.PhoneNumber);
				var message = addFormElement("Message", "textarea", "message", messageData.Message, regexpCheck(/.+/, "Please enter a message")),
				    submit = addFormSubmit("Send Message", function() {
					messageData.Message = message[0].value;
					if (messageData.Message === "") {
						return;
					}
					var elements = [message[0], submit];
					elements.map(disableElement);
					rpc.sendMessage(messageData, function(error) {
						if (typeof error === "string" && error.length > 0) {
							elements.map(enableElement);
							message[1].setInnerText(error);
						} else {
							stack.removeLayer();
						}
					});
				});
			}],
			[ "Email", function() {
				addFormElement("Client Name", "text", "", client.Name);
				addFormElement("Client Email", "text", "", client.Email);
				var message = addFormElement("Message", "textarea", "message", emailData.Message, regexpCheck(/.+/, "Please enter a message")),
				    submit = addFormSubmit("Send Message", function() {
					emailData.Message = message[0].value;
					if (emailData.Message === "") {
						return;
					}
					var elements = [message[0], submit];
					elements.map(disableElement);
					rpc.sendEmail(emailData, function(error) {
						if (typeof error === "string" && error.length > 0) {
							elements.map(enableElement);
							message[1].setInnerText(error);
						} else {
							stack.removeLayer();
						}
					});
				});
			}]
		));
		stack.setFragment();
	},
	addFormElement = function(name, type, id, contents, onBlur) {
		var label = createElement("label").setInnerText(name),
		    input;
		if (type === "textarea") {
			input = createElement("textarea");
			input.setAttribute("spellcheck", "false");
			input.value = contents;
		} else {
			input = createElement("input");
			input.setAttribute("type", type);
			if (type === "checkbox") {
				if (contents === true) {
					input.checked = true;
				}
			} else {
				input.setAttribute("value", contents);
			}
		}
		input.setAttribute("id", id);
		if (type === "hidden") {
			return layer.appendChild(input);
		}
		if (id === "") {
			input.setAttribute("readonly", "readonly");
		} else {
			label.setAttribute("for", id);
		}
		if (typeof onBlur === "function") {
			input.addEventListener("blur", onBlur.bind(input));
		}
		var error = createElement("div");
		error.setAttribute("class", "error");
		error.setAttribute("id", "error_"+id);
		layer.appendChild(label);
		layer.appendChild(input);
		layer.appendChild(error);
		layer.appendChild(createElement("br"));
		return [input, error];
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
	makeExportButton = function(exportButton, typeStr, id, startDate, endDate) {
		var type = exportButton.appendChild(createElement("input")),
		    submit = exportButton.appendChild(createElement("input"));
		type.setAttribute("type", "hidden");
		type.setAttribute("name", "type");
		type.setAttribute("value", typeStr);
		submit.setAttribute("type", "submit");
		submit.setAttribute("value", "Export");
		if (typeof id !== "undefined") {
			if (Array.isArray && Array.isArray(id)) {
				for (var i = 0; i < id.length; i++) {
					var idE = exportButton.appendChild(createElement("input"));
					idE.setAttribute("type", "hidden");
					idE.setAttribute("name", "id");
					idE.setAttribute("value", id[i].toString());
				}
			} else {
				var idE = exportButton.appendChild(createElement("input"));
				idE.setAttribute("type", "hidden");
				idE.setAttribute("name", "id");
				idE.setAttribute("value", id.toString());
			}
			if (typeof startDate !== "undefined" && typeof endDate !== "undefined") {
				var start = exportButton.appendChild(createElement("input")),
				    end = exportButton.appendChild(createElement("input"));
				start.setAttribute("type", "hidden");
				start.setAttribute("name", "startTime");
				start.setAttribute("value", startDate.getTime().toString());
				end.setAttribute("type", "hidden");
				end.setAttribute("name", "endTime");
				end.setAttribute("value", endDate.getTime().toString());
			}
		}
	},
	showDriver = function(driver) {
		stack.addLayer("showDriver");
		stack.addFragment();
		layer.appendChild(createElement("h1")).setInnerText(driver.Name);
		layer.appendChild(makeTabs(
			[ "Details", function() {
				var toPrint = layer.appendChild(createElement("div"));
				toPrint.setAttribute("class", "toPrint");
				toPrint.appendChild(createElement("h2")).setInnerText("Driver Details").setAttribute("class", "printOnly");
				toPrint.appendChild(createElement("label")).setInnerText("Name");
				toPrint.appendChild(createElement("div")).setInnerText(driver.Name);
				toPrint.appendChild(createElement("label")).setInnerText("Phone Number");
				toPrint.appendChild(createElement("div")).setInnerText(driver.PhoneNumber);
				toPrint.appendChild(createElement("label")).setInnerText("Registration Number");
				toPrint.appendChild(createElement("div")).setInnerText(driver.RegistrationNumber);
				toPrint.appendChild(createElement("label")).setInnerText("No. of Events");
				var bookings = toPrint.appendChild(createElement("div")).setInnerText("-");
//				    tmpNote = {Note:""};
				toPrint.appendChild(createElement("label")).setInnerText("Notes");
				toPrint.appendChild(makeNote(rpc.getDriverNote.bind(rpc, driver.ID), rpc.setDriverNote.bind(rpc, driver.ID)));
/*				toPrint.appendChild(makeNote(function(callback) {
					rpc.getDriverNote(driver.ID, function(noteText) {
						tmpNote = noteJSON(noteText);
						callback(tmpNote.Note)
					});
				}, function(noteText) {
					tmpNote.Note = noteText;
					rpc.setDriverNote(driver.ID, JSON.stringify(tmpNote));
				}));*/
				rpc.getNumEventsDriver(driver.ID, bookings.setInnerText.bind(bookings));
			}],
			[ "Events", function() {
				var eventsStartDate = new Date(),
				    eventsEndDate = new Date();
				return function () {
					var startDate = addFormElement("Start Date", "text", "startDate", eventsStartDate.toDateString(), dateCheck),
					    endDate = addFormElement("End Date", "text", "endDate", eventsEndDate.toDateString(), dateCheck),
					    getEvents = addFormSubmit("Show Events", function() {
						eventTable.removeChildren(function(elm) {
							return elm !== tableTitles;
						});
						exportButton.removeChildren();
						var startParts = startDate[0].value.split("/"),
						    endParts = endDate[0].value.split("/");
						eventsStartDate = new Date(startParts[2], startParts[1]-1, startParts[0]),
						eventsEndDate = new Date(endParts[2], endParts[1]-1, endParts[0]);
						rpc.getEventsWithDriver(driver.ID, eventsStartDate.getTime(), eventsEndDate.getTime() + (24 * 3600 * 1000), function(events) {
							var row,
							    i = 0,
							    pT = "Driver Sheet for " + driver.Name + " for " + eventsStartDate.toDateString(),
							    totalWaiting = 0, totalMiles = 0, totalTrip = 0, totalDriverHours = 0, totalParking = 0, totalSub = 0, totalPrice = 0,
							    wg = new waitGroup(function() {
								var row = createElement("tr");
								//row.appendChild(createElement("td")).setInnerText(events.length + " events").setAttribute("colspan", "8");
								row.appendChild(createElement("td")).setInnerText(events.length + " events").setAttribute("colspan", "6");
								row.appendChild(createElement("td")).setInnerText(totalWaiting);
								row.appendChild(createElement("td")).setInnerText(totalMiles);
								row.appendChild(createElement("td")).setInnerText((new Date(totalTrip)).toTimeString());
								//row.appendChild(createElement("td")).setInnerText((new Date(totalDriverHours)).toTimeString());
								row.appendChild(createElement("td")).setInnerText(totalDriverHours / 3600000);
								row.appendChild(createElement("td")).setInnerText("£" + (totalParking / 100).formatMoney());
								row.appendChild(createElement("td")).setInnerText("£" + (totalSub / 100).formatMoney());
								row.appendChild(createElement("td")).setInnerText("£" + (totalPrice / 100).formatMoney());
								eventTable.appendChild(row).setAttribute("class", "overline");
							    });
							if (eventsStartDate.getTime() !== eventsEndDate.getTime()) {
								pT += " to " + eventsEndDate.toDateString();
							}
							printTitle.setInnerText(pT);
							exportButton.removeChildren();
							if (events.length === 0) {
								eventTable.appendChild(createElement("tr")).appendChild(createElement("td")).setInnerText("No Events").setAttribute("colspan", "14");
								return;
							}
							makeExportButton(exportButton ,"driverEvents", driver.ID, eventsStartDate, eventsEndDate);
							for (; i < events.length; i++) {
								row = createElement("tr");
								row.appendChild(createElement("td")).setInnerText(new Date(events[i].Start).toLocaleString());
								row.appendChild(createElement("td")).setInnerText(new Date(events[i].End).toLocaleString());
								var clientCell = row.appendChild(createElement("td")),
								    //phoneCell = row.appendChild(createElement("td")),
								    companyCell = createElement("td"),
								    //inCarCell = createElement("td").setInnerText("-"),
								    waitingCell = createElement("td").setInnerText("-"),
								    milesCell = createElement("td").setInnerText("-"),
								    tripCell = createElement("td").setInnerText("-"),
								    driverHoursCell = createElement("td").setInnerText("-"),
								    parkingCell = createElement("td").setInnerText("-"),
								    subCell = createElement("td").setInnerText("-"),
								    priceCell = createElement("td").setInnerText("-");
								row.appendChild(createElement("td")).setInnerText(events[i].From);
								row.appendChild(createElement("td")).setInnerText(events[i].To);
								row.appendChild(companyCell);
								//row.appendChild(inCarCell).setAttribute("class", "noPrint");
								row.appendChild(waitingCell).setAttribute("class", "noPrint");
								row.appendChild(milesCell).setAttribute("class", "noPrint");
								row.appendChild(tripCell).setAttribute("class", "noPrint");
								row.appendChild(driverHoursCell).setAttribute("class", "noPrint");
								row.appendChild(parkingCell).setAttribute("class", "noPrint");
								row.appendChild(subCell).setAttribute("class", "noPrint");
								row.appendChild(priceCell).setAttribute("class", "noPrint");
								//rpc.getClient(events[i].ClientID, function(clientCell, phoneCell, companyCell, client) {
								rpc.getClient(events[i].ClientID, function(clientCell, companyCell, client) {
									clientCell.setInnerText(client.Name);
									//phoneCell.setInnerText(client.PhoneNumber);
									rpc.getCompany(client.CompanyID, function(company) {
										companyCell.setInnerText(company.Name);
									});
								//}.bind(null, clientCell, phoneCell, companyCell));
								}.bind(null, clientCell, companyCell));
								wg.add();
								//rpc.getEventFinals(events[i].ID, function(inCarCell, waitingCell, milesCell, tripCell, driverHoursCell, parkingCell, subCell, i, eventFinals) {
								rpc.getEventFinals(events[i].ID, function(waitingCell, milesCell, tripCell, driverHoursCell, parkingCell, subCell, priceCell, i, eventFinals) {
									if (eventFinals.FinalsSet) {
										//inCarCell.setInnerText((new Date(eventFinals.InCar)).toTimeString()).removeAttribute("class");
										waitingCell.setInnerText(eventFinals.Waiting).removeAttribute("class");
										milesCell.setInnerText(eventFinals.Miles).removeAttribute("class");
										tripCell.setInnerText((new Date(eventFinals.Trip)).toTimeString()).removeAttribute("class");
										//driverHoursCell.setInnerText((new Date(eventFinals.DriverHours)).toTimeString()).removeAttribute("class");
										driverHoursCell.setInnerText(eventFinals.DriverHours / 3600000).removeAttribute("class");
										parkingCell.setInnerText("£" + (eventFinals.Parking / 100).formatMoney()).removeAttribute("class");
										subCell.setInnerText("£" + (eventFinals.Sub / 100).formatMoney()).removeAttribute("class");
										priceCell.setInnerText("£" + (eventFinals.Price / 100).formatMoney()).removeAttribute("class");
										totalMiles += eventFinals.Miles;
										totalWaiting += eventFinals.Waiting;
										totalTrip += eventFinals.Trip;
										totalDriverHours += eventFinals.DriverHours;
										totalParking += eventFinals.Parking;
										totalSub += eventFinals.Sub;
										totalPrice += eventFinals.Price;
									}
									wg.done();
								//}.bind(null, inCarCell, waitingCell, milesCell, tripCell, driverHoursCell, parkingCell, subCell, i));
								}.bind(null, waitingCell, milesCell, tripCell, driverHoursCell, parkingCell, subCell, priceCell, i));
								eventTable.appendChild(row);
							}
						});
					    }),
					    toPrint = layer.appendChild(createElement("div")),
					    printTitle = toPrint.appendChild(createElement("h2")),
					    eventTable = toPrint.appendChild(createElement("table")),
					    tableTitles = eventTable.appendChild(createElement("tr")),
					    exportButton = layer.appendChild(createElement("form"));
					exportButton.setAttribute("method", "post");
					exportButton.setAttribute("action", "/export");
					exportButton.setAttribute("target", "_new");
					toPrint.setAttribute("class", "toPrint");
					printTitle.setAttribute("class", "printOnly");
					tableTitles.appendChild(createElement("th")).setInnerText("Start");
					tableTitles.appendChild(createElement("th")).setInnerText("End");
					tableTitles.appendChild(createElement("th")).setInnerText("Client");
					//tableTitles.appendChild(createElement("th")).setInnerText("Phone Number");
					tableTitles.appendChild(createElement("th")).setInnerText("From");
					tableTitles.appendChild(createElement("th")).setInnerText("To");
					tableTitles.appendChild(createElement("th")).setInnerText("Company");
					//tableTitles.appendChild(createElement("th")).setInnerText("In Car");
					tableTitles.appendChild(createElement("th")).setInnerText("Waiting");
					tableTitles.appendChild(createElement("th")).setInnerText("Miles");
					tableTitles.appendChild(createElement("th")).setInnerText("Flight Time");
					tableTitles.appendChild(createElement("th")).setInnerText("Driver Hours (h)");
					tableTitles.appendChild(createElement("th")).setInnerText("Parking");
					tableTitles.appendChild(createElement("th")).setInnerText("Sub Price");
					tableTitles.appendChild(createElement("th")).setInnerText("Price");
					getEvents.dispatchEvent(new MouseEvent("click", {"view": window, "bubble": false, "cancelable": true}));
				};
			}()],
			[ "Options", function() {
				var edit = layer.appendChild(createElement("div")).setInnerText("Edit Driver"),
				    deleter = layer.appendChild(createElement("div")).setInnerText("Delete Driver");
				edit.setAttribute("class", "simpleButton");
				edit.addEventListener("click", function() {
					stack.addLayer("editDriver", function(d) {
						if (typeof d !== "undefined") {
							events.reload("driver", d.ID);
						}
					});
					setDriver(driver);
				});
				deleter.setAttribute("class", "simpleButton");
				deleter.addEventListener("click", function() {
					if(confirm("Are you sure you want to remove this driver? NB: This will also remove all events attached to the driver.")) {
						rpc.removeDriver(driver.ID);
						events.reload();
					}
				});
			}]
		));
		stack.setFragment();
	},
	setDriver = function(driver) {
		stack.addFragment();
		layer.appendChild(createElement("h1")).setInnerText((driver.ID == 0) ? "Add Driver" : "Edit Driver");
		var driverName = addFormElement("Driver Name", "text", "driver_name", driver.Name, regexpCheck(/.+/, "Please enter a valid name")),
		    regNumber = addFormElement("Registration Number", "text", "driver_reg", driver.RegistrationNumber, regexpCheck(/[a-zA-Z0-9 ]+/, "Please enter a valid Vehicle Registration Number")),
		    phoneNumber = addFormElement("Phone Number", "text", "driver_phone", driver.PhoneNumber);
		addFormSubmit((driver.ID == 0) ? "Add Driver" : "Edit Driver", function() {
			var parts = [this, driverName[0], regNumber[0], phoneNumber[0]];
			parts.map(disableElement);
			driver.Name = driverName[0].value;
			driver.RegistrationNumber = regNumber[0].value;
			driver.PhoneNumber = phoneNumber[0].value;
			rpc.setDriver(driver, function(resp) {
				if (resp.Errors) {
					driverName[1].setInnerText(resp.NameError);
					regNumber[1].setInnerText(resp.RegError);
					phoneNumber[1].setInnerText(resp.PhoneError);
					parts.map(enableElement);
				} else {
					driver.ID = resp.ID;
					stack.removeLayer(driver);
				}
			});
		});
		stack.setFragment();
	},
	toggleDriver = function(id) {

	},
	swapDrivers = function(a, b) {

	},
	driverList = function(drivers) {
		stack.addFragment();
		layer.appendChild(createElement("h1")).setInnerText("Drivers");
		var table = createElement("table"),
		    headerRow = table.appendChild(createElement("tr")),
		    driverIDs = Object.keys(drivers),
		    swapAnim = function(row, dir) {
			if (dir === "up") {
				row = row.previousSibling;
			}
			var toSwap = row.nextSibling,
			    fromID = row.getAttribute("driverID"),
			    toID = toSwap.getAttribute("driverID"),
			    oldPos = drivers[fromID].Pos;
			drivers[fromID].Pos = drivers[toID].Pos;
			drivers[toID].Pos = oldPos;
			Array.slice(row.parentNode.getElementsByTagName("button")).forEach(b=>b.setAttribute("disabled", "disabled"));
			row.setAttribute("class", "rowSwapper swapping");
			toSwap.setAttribute("class", "rowSwapper swapping");
			window.setTimeout(function() {
				row.parentNode.insertBefore(toSwap, row);
				window.setTimeout(function() {
					row.setAttribute("class", "rowSwapper");
					toSwap.setAttribute("class", "rowSwapper");
				}, 20);
				window.setTimeout(function() {
					var buttons = Array.slice(row.parentNode.getElementsByTagName("button"));
					buttons.forEach(function(button) {
						button.removeAttribute("disabled");
					});
					buttons[0].setAttribute("disabled", "disabled");
					buttons[buttons.length-1].setAttribute("disabled", "disabled");
				}, 500);
			}, 500);
		    };
		headerRow.appendChild(createElement("th")).setInnerText("Order").setAttribute("class", "noPrint");
		headerRow.appendChild(createElement("th")).setInnerText("Driver Name");
		headerRow.appendChild(createElement("th")).setInnerText("Registration Number");
		headerRow.appendChild(createElement("th")).setInnerText("Phone Number");
		headerRow.appendChild(createElement("th")).setInnerText("Visible").setAttribute("class", "noPrint");

		driverIDs = driverIDs.sort(function(a, b) {
			return drivers[a].Pos - drivers[b].Pos;
		});

		for (var i = 0; i < driverIDs.length; i++) {
			var driver = drivers[driverIDs[i]],
			    row = table.appendChild(createElement("tr")),
			    pos = row.appendChild(createElement("td")),
			    show = createElement("td"),
			    up = createElement("button"),
			    down = createElement("button"),
			    showHide = show.appendChild(createElement("input")),
			    showHideLabel = show.appendChild(createElement("label"));
			row.setAttribute("class", "rowSwapper");
			row.setAttribute("driverID", driverIDs[i]);
			row.appendChild(createElement("td")).setInnerText(driver.Name);
			row.appendChild(createElement("td")).setInnerText(driver.RegistrationNumber);
			row.appendChild(createElement("td")).setInnerText(driver.PhoneNumber);
			if (i == 0) {
				up.setAttribute("disabled", "disabled");
			}
			up.setInnerText("⇑")
			up.addEventListener("click", swapAnim.bind(null, row, "up"));
			pos.appendChild(up);
			if (i == driverIDs.length - 1) {
				down.setAttribute("disabled", "disabled");
			}
			down.setInnerText("⇓")
			down.addEventListener("click", swapAnim.bind(null, row, "down"));
			pos.appendChild(down);
			pos.setAttribute("class", "noPrint");
			showHide.setAttribute("type", "checkbox");
			if (driver.Show !== false) {
				showHide.setAttribute("checked", "checked");
			}
			showHide.setAttribute("id", "driver_showhide_" + driver.ID);
			showHideLabel.setAttribute("for", "driver_showhide_" + driver.ID);
			show.setAttribute("class", "toggleBox noPrint");
			showHide.addEventListener("change", function(showHide, id) {
				drivers[id].Show = showHide.checked;
			}.bind(null, showHide, driverIDs[i]), false);
			row.appendChild(show);
		}

		layer.appendChild(table);
		stack.setFragment();
		layer.setAttribute("class", "toPrint");
	},
	addDriver = function() {
		setDriver({
			"ID": 0,
			"Name": "",
			"RegistrationNumber": "",
			"PhoneNumber": "",
			"Show": true,
		});
	},
	setClient = function(client) {
		stack.addFragment();
		layer.appendChild(createElement("h1")).setInnerText((client.ID == 0) ? "Add Client" : "Edit Client");
		var clientName = addFormElement("Client Name", "text", "client_name", client.Name, regexpCheck(/.+/, "Please enter a valid name")),
		    companyID = addFormElement("", "hidden", "client_company_id", client.CompanyID),
		    companyName = addFormElement("Company Name", "text", "client_company_name", client.CompanyName, regexpCheck(/.+/, "Please enter a valid name")),
		    clientPhone = addFormElement("Mobile Number", "text", "client_phone", client.PhoneNumber),
		    clientRef = addFormElement("Client Ref", "text", "client_ref", client.Reference, regexpCheck(/.+/, "Please enter a reference code"));
		addLister(companyName[1], function() {
			companyName[1].setInnerText("");
			stack.addLayer("companyList", function(company) {
				if (typeof company === "undefined") {
					return;
				}
				companyID.value = company.ID;
				companyName[0].value = company.Name;
				companyName[1].setInnerText("");
			});
			companyList(true);
		});
		autocomplete(rpc.autocompleteCompanyName, companyName[0], companyID);
		addFormSubmit((client.ID == 0) ? "Add Client" : "Edit Client", function() {
			var parts = [this, clientName[0], companyName[0], clientPhone[0], clientRef[0]];
			parts.map(disableElement);
			client.Name = clientName[0].value;
			client.CompanyID = parseInt(companyID.value);
			client.PhoneNumber = clientPhone[0].value;
			client.Reference = clientRef[0].value;
			rpc.setClient(client, function (resp) {
				if (resp.Errors) {
					clientName[1].setInnerText(resp.NameError);
					companyName[1].setInnerText(resp.CompanyError);
					clientPhone[1].setInnerText(resp.PhoneError);
					clientRef[1].setInnerText(resp.ReferenceError);
					parts.map(enableElement);
				} else {
					client.ID = resp.ID;
					stack.removeLayer(client);
				}
			});
		});
		stack.setFragment();
	},
	addClient = function() {
		setClient({
			"ID": 0,
			"Name": "",
			"CompanyName": "",
			"CompanyID": 0,
			"PhoneNumber": "",
			"Reference": "",
		});
	},
	setCompany = function(company) {
		stack.addFragment();
		layer.appendChild(createElement("h1")).setInnerText((company.ID == 0) ? "Add Company" : "Edit Company");
		var companyName = addFormElement("Company Name", "text", "company_name", company.Name, regexpCheck(/.+/, "Please enter a valid name")),
		    address = addFormElement("Company Address", "textarea", "company_address", company.Address, regexpCheck(/.+/, "Please enter a valid address")),
		    color = addFormElement("Company Colour", "color", "company_color", company.Colour.formatColour());
		addFormSubmit((company.ID == 0) ? "Add Company" : "Edit Company", function() {
			var parts = [this, companyName[0], address[0]];
			parts.map(disableElement);
			company.Name = companyName[0].value;
			company.Address = address[0].value;
			company.Colour = parseInt(color[0].value.substring(1), 16);
			rpc.setCompany(company, function(resp) {
				if (resp.Errors) {
					companyName[1].setInnerText(resp.NameError);
					address[1].setInnerText(resp.AddressError);
					parts.map(enableElement);
				} else {
					company.ID = resp.ID;
					stack.removeLayer(company);
				}
			});
		});
		stack.setFragment();
	},
	addCompany = function() {
		setCompany({
			"ID": 0,
			"Name": "",
			"Address": "",
			"Colour": 16777215,
		});
	},
	showEvent = function(e) {
		stack.addLayer("showEvent");
		stack.addFragment();
		layer.appendChild(createElement("h1")).setInnerText("Event Details");
		var tabData = new Array();
		tabData[0] = [ "Details", function () {
			var toPrint = layer.appendChild(createElement("div"));
			toPrint.setAttribute("class", "toPrint");
			toPrint.appendChild(createElement("h2")).setInnerText("Event Details").setAttribute("class", "printOnly");
			toPrint.appendChild(createElement("label")).setInnerText("Client Name");
			var clientName = toPrint.appendChild(createElement("div")).setInnerText("-"),
			    clientPhone = createElement("div").setInnerText("-"),
			    clientRef = createElement("div").setInnerText("-"),
			    companyName = createElement("div").setInnerText("-"),
			    driverName = createElement("div").setInnerText("-"),
			    driverReg = createElement("div").setInnerText("-");
			toPrint.appendChild(createElement("label")).setInnerText("Client Phone Number");
			toPrint.appendChild(clientPhone);
			toPrint.appendChild(createElement("label")).setInnerText("Client Reference");
			toPrint.appendChild(clientRef);
			toPrint.appendChild(createElement("label")).setInnerText("Other Passengers");
			toPrint.appendChild(createElement("div").setInnerText(e.Other + "\u00A0"));
			toPrint.appendChild(createElement("label")).setInnerText("Company Name");
			toPrint.appendChild(companyName);
			toPrint.appendChild(createElement("label")).setInnerText("Driver Name");
			toPrint.appendChild(driverName);
			toPrint.appendChild(createElement("label")).setInnerText("Driver Registration");
			toPrint.appendChild(driverReg);
			toPrint.appendChild(createElement("label")).setInnerText("Start Time");
			toPrint.appendChild(createElement("div")).setInnerText(new Date(e.Start).toLocaleString());
			toPrint.appendChild(createElement("label")).setInnerText("End Time");
			toPrint.appendChild(createElement("div")).setInnerText(new Date(e.End).toLocaleString());
			toPrint.appendChild(createElement("label")).setInnerText("From");
			toPrint.appendChild(createElement("div")).setInnerText(e.From);
			toPrint.appendChild(createElement("label")).setInnerText("To");
			toPrint.appendChild(createElement("div")).setInnerText(e.To);
			if (e.Start < (new Date()).getTime()) {
				var inCar = createElement("div").setInnerText("-"),
				    parking = createElement("div").setInnerText("-"),
				    waiting = createElement("div").setInnerText("-"),
				    dropOff = createElement("div").setInnerText("-"),
				    miles = createElement("div").setInnerText("-"),
				    tripTime = createElement("div").setInnerText("-"),
				    driverHours = createElement("div").setInnerText("-"),
				    price = createElement("div").setInnerText("-"),
				    sub = createElement("div").setInnerText("-"),
				    invoiceTo = createElement("div").setInnerText("-"),
				    invoiceFrom = createElement("div").setInnerText("-"),
				    invoiceNote = createElement("div").setInnerText("-");
				toPrint.appendChild(createElement("label")).setInnerText("Flight Time");
				toPrint.appendChild(tripTime);
				toPrint.appendChild(createElement("label")).setInnerText("In Car Time");
				toPrint.appendChild(inCar);
				toPrint.appendChild(createElement("label")).setInnerText("Drop Off Time");
				toPrint.appendChild(dropOff);
				toPrint.appendChild(createElement("label")).setInnerText("Waiting Time");
				toPrint.appendChild(waiting);
				toPrint.appendChild(createElement("label")).setInnerText("Parking Costs");
				toPrint.appendChild(parking);
				toPrint.appendChild(createElement("label")).setInnerText("Driver Time (h)");
				toPrint.appendChild(driverHours);
				toPrint.appendChild(createElement("label")).setInnerText("Miles Travelled");
				toPrint.appendChild(miles);
				toPrint.appendChild(createElement("label")).setInnerText("Sub Price");
				toPrint.appendChild(sub);
				toPrint.appendChild(createElement("label")).setInnerText("Total Price");
				toPrint.appendChild(price);
				toPrint.appendChild(createElement("label")).setInnerText("Invoice From");
				toPrint.appendChild(invoiceFrom);
				toPrint.appendChild(createElement("label")).setInnerText("Invoice To");
				toPrint.appendChild(invoiceTo);
				toPrint.appendChild(createElement("label")).setInnerText("Invoice Note");
				toPrint.appendChild(invoiceNote);
				rpc.getEventFinals(e.ID, function(eventFinals) {
					if (!eventFinals.FinalsSet) {
						return;
					}
					inCar.setInnerText((new Date(eventFinals.InCar)).toTimeString());
					parking.setInnerText("£" + (eventFinals.Parking / 100).formatMoney());
					waiting.setInnerText(eventFinals.Waiting + " minutes");
					dropOff.setInnerText((new Date(eventFinals.Drop)).toTimeString());
					miles.setInnerText(eventFinals.Miles);
					tripTime.setInnerText((new Date(eventFinals.Trip)).toTimeString());
					//driverHours.setInnerText((new Date(eventFinals.DriverHours)).toTimeString());
					driverHours.setInnerText(eventFinals.DriverHours / 3600000);
					price.setInnerText("£" + (eventFinals.Price / 100).formatMoney());
					sub.setInnerText("£" + (eventFinals.Sub / 100).formatMoney());

					invoiceNote.setPreText(eventFinals.InvoiceNote + "\u00A0");
					invoiceFrom.setInnerText(eventFinals.InvoiceFrom + "\u00A0");
					invoiceTo.setInnerText(eventFinals.InvoiceTo + "\u00A0");
				});
			}
			toPrint.appendChild(createElement("label")).setInnerText("Notes");

			toPrint.appendChild(makeNote(rpc.getEventNote.bind(rpc, e.ID), rpc.setEventNote.bind(rpc, e.ID)));
			/*var tmpNote = {"Note":""};
			toPrint.appendChild(makeNote(function(callback) {
				rpc.getEventNote(e.ID, function(noteText) {
					tmpNote = noteJSON(noteText);
					if (typeof tmpNote.InvoiceNote !== "undefined") {
						invoiceNote.setPreText(tmpNote.InvoiceNote);
					}
					if (typeof tmpNote.InvoiceFrom !== "undefined") {
						invoiceFrom.setInnerText(tmpNote.InvoiceFrom);
					}
					if (typeof tmpNote.InvoiceTo !== "undefined") {
						invoiceTo.setInnerText(tmpNote.InvoiceTo);
					}
					if (typeof tmpNote.ClientRef !== "undefined") {
						clientRef.setInnerText(tmpNote.ClientRef + "\u00A0");
					}
					callback(tmpNote.Note);
				});
			}, function(noteText) {
				tmpNote.Note = noteText;
				rpc.setEventNote(e.ID, JSON.stringify(tmpNote));
			}));*/
			rpc.getClient(e.ClientID, function(client) {
				clientName.setInnerText(client.Name);
				clientPhone.setInnerText(client.PhoneNumber);
				clientRef.setInnerText(client.Reference + "\u00A0");
				rpc.getCompany(client.CompanyID, function(company) {
					companyName.setInnerText(company.Name);
				});
			});
			if (e.DriverID === 0) {
				driverName.setInnerText("Unassigned");
			} else {
				rpc.getDriver(e.DriverID, function(driver) {
					driverName.setInnerText(driver.Name);
					if (driver.RegistrationNumber !== "" && driver.RegistrationNumber != " ") {
						driverReg.setInnerText(driver.RegistrationNumber);
					}
				});
			}
		}];
		if (e.Start < (new Date()).getTime() && e.DriverID > 0) {
			tabData[tabData.length] = [ "Final Details", function() {
				var tripTime = addFormElement("Flight Time", "text", "trip", "", regexpCheck(/^([0-1]?[0-9]|2[0-3]):[0-5]?[0-9]$/, "Time format unrecognised (HH:MM)")),
				    inCar = addFormElement("In Car Time", "text", "inCar", "", regexpCheck(/^([0-1]?[0-9]|2[0-3]):[0-5]?[0-9]$/, "Time format unrecognised (HH:MM)")),
				    dropOff = addFormElement("Drop Off Time", "text", "dropOff", "", regexpCheck(/^([0-1]?[0-9]|2[0-3]):[0-5]?[0-9]$/, "Time format unrecognised (HH:MM)")),
				    waiting = addFormElement("Waiting Time (minutes)", "text", "waiting", "", regexpCheck(/^[0-9]+$/, "Please insert a number (or 0)")),
				    parking = addFormElement("Parking Costs (£)", "text", "parking", "", regexpCheck(/^[0-9]+(\.[0-9][0-9])?$/, "Please enter a valid amount")),
				    driverHours = addFormElement("Driver Time (h)", "text", "driverHours", "", regexpCheck(/^[0-9]+(\.[0-9]+)?$/, "Number format unrecognised (0.00)")),
				    miles = addFormElement("Miles Travelled", "text", "miles", "", regexpCheck(/^[0-9]+$/, "Please insert a number (or 0)")),
				    //driverHours = addFormElement("Driver Time", "text", "driverHours", "", regexpCheck(/^([0-1]?[0-9]|2[0-3]):[0-5]?[0-9]$/, "Time format unrecognised (HH:MM)")),
				    sub = addFormElement("Sub Price (£)", "text", "sub", "", regexpCheck(/^[0-9]+(\.[0-9][0-9])?$/, "Please enter a valid amount")),
				    price = addFormElement("Total Price To Client (£)", "text", "price", "", regexpCheck(/^[0-9]+(\.[0-9][0-9])?$/, "Please enter a valid amount")),
				    invoiceFrom = addFormElement("Invoice From", "text", "invoiceFrom", e.From.replace(/\n/g, " ")),
				    invoiceTo = addFormElement("Invoice To", "text", "invoiceTo", e.To.replace(/\n/g, " ")),
				    invoiceNotes = addFormElement("Invoice Notes", "textarea", "invoiceNotes", "");
				    //invoiceNotesJSON = {"Note":""};
				addFormSubmit("Set Details", function() {
					var errors = false,
					    eventFinals = {},
					    parts;
					[inCar, waiting, dropOff, miles, tripTime, parking, sub, price].map(function(error) {
						if (error[1].hasChildNodes()) {
							errors = true;
						}
					});
					if (errors) {
						return;
					}
					parts = inCar[0].value.split(":");
					eventFinals.InCar = (new Date(1970, 0, 1, parseInt(parts[0]), parseInt(parts[1]))).getTime();
					eventFinals.Waiting = parseInt(waiting[0].value);
					parts = dropOff[0].value.split(":");
					eventFinals.Drop = (new Date(1970, 0, 1, parseInt(parts[0]), parseInt(parts[1]))).getTime();
					eventFinals.Miles = parseInt(miles[0].value);
					parts = tripTime[0].value.split(":");
					eventFinals.Trip = (new Date(1970, 0, 1, parseInt(parts[0]), parseInt(parts[1]))).getTime();
					//parts = driverHours[0].value.split(":");
					//eventFinals.DriverHours = (new Date(1970, 0, 1, parseInt(parts[0]), parseInt(parts[1]))).getTime();
					eventFinals.DriverHours = Math.floor(parseFloat(driverHours[0].value) * 3600000);
					eventFinals.Parking = Math.floor(parseFloat(parking[0].value) * 100);
					eventFinals.Sub = Math.floor(parseFloat(sub[0].value) * 100);
					eventFinals.Price = Math.floor(parseFloat(price[0].value) * 100);
					eventFinals.ID = e.ID;
					eventFinals.InvoiceFrom = invoiceFrom[0].value;
					eventFinals.InvoiceTo = invoiceTo[0].value;
					eventFinals.InvoiceNote = invoiceNotes[0].value;
				/*	invoiceNotesJSON.InvoiceFrom = invoiceFrom[0].value;
					invoiceNotesJSON.InvoiceTo = invoiceTo[0].value;
					invoiceNotesJSON.InvoiceNote = invoiceNotes[0].value;
					invoiceNotesJSON.ClientRef = clientRef[0].value;
					rpc.setEventNote(e.ID, JSON.stringify(invoiceNotesJSON));*/
					rpc.setEventFinals(eventFinals, function() {
						stack.removeLayer();
						showEvent(e);
					});
				});
				/*rpc.getEventNote(e.ID, function(noteText) {
					invoiceNotesJSON = noteJSON(noteText);
					if (typeof invoiceNotesJSON.InvoiceNote !== "undefined") {
						invoiceNotes[0].value = invoiceNotesJSON.InvoiceNote;
					}
					if (typeof invoiceNotesJSON.InvoiceFrom !== "undefined") {
						invoiceFrom[0].value = invoiceNotesJSON.InvoiceFrom;
					}
					if (typeof invoiceNotesJSON.InvoiceTo !== "undefined") {
						invoiceTo[0].value = invoiceNotesJSON.InvoiceTo;
					}
					if (typeof invoiceNotesJSON.ClientRef !== "undefined") {
						clientRef[0].value = invoiceNotesJSON.ClientRef;
					} else {
						rpc.getClient(e.ClientID, function(client) {
							clientRef[0].value = client.Reference;
						});
					}*/
				rpc.getEventFinals(e.ID, function(eventFinals) {
					inCar[0].value = (new Date(eventFinals.InCar)).toTimeString();
					waiting[0].value = eventFinals.Waiting;
					dropOff[0].value = (new Date(eventFinals.Drop)).toTimeString();
					miles[0].value = eventFinals.Miles;
					tripTime[0].value = (new Date(eventFinals.Trip)).toTimeString();
					//driverHours[0].value = (new Date(eventFinals.DriverHours)).toTimeString();
					driverHours[0].value = eventFinals.DriverHours / 3600000
					parking[0].value = (eventFinals.Parking / 100).formatMoney();
					sub[0].value = (eventFinals.Sub / 100).formatMoney();
					price[0].value = (eventFinals.Price / 100).formatMoney();
					if (eventFinals.InvoiceTo != "") {
						invoiceTo[0].value = eventFinals.InvoiceTo;
					}
					if (eventFinals.InvoiceFrom != "") {
						invoiceFrom[0].value = eventFinals.InvoiceFrom;
					}
					invoiceNotes[0].value = eventFinals.InvoiceNote;
				});
				//});
			}];
		}
		tabData[tabData.length] = [ "Options", function () {
			var edit = layer.appendChild(createElement("div")).setInnerText("Edit Event"),
			    deleter = layer.appendChild(createElement("div")).setInnerText("Delete Event");
			edit.setAttribute("class", "simpleButton");
			edit.addEventListener("click", function() {
				stack.addLayer("editEvent", function(e) {
					if (typeof e !== "undefined") {
						events.reload("event", e.ID);
					}
				});
				rpc.getClient(e.ClientID, function(c) {
					e.ClientName = c.Name;
					if (e.DriverID === 0) {
						e.DriverName = "Unassigned";
						setEvent(e);
					} else {
						rpc.getDriver(e.DriverID, function(d) {
							e.DriverName = d.Name;
							setEvent(e);
						});
					}
				});
			});
			deleter.setAttribute("class", "simpleButton");
			deleter.addEventListener("click", function() {
				if(confirm("Are you sure you want to remove this event?")) {
					rpc.removeEvent(e.ID);
					events.reload();
				}
			});
		}];
		layer.appendChild(makeTabs.apply(null, tabData));
		stack.setFragment();
	},
	setEvent = function(event) {
		stack.addFragment();
		layer.appendChild(createElement("h1")).setInnerText((event.ID == 0) ? "Add Event" : "Edit Event");
		addFormElement("Driver", "text", "", event.DriverName);
		addFormElement("Start", "text", "", dateTimeFormat(event.Start));
		var driverTime = addFormElement("End", "text", "", dateTimeFormat(event.End)),
		    clientID = addFormElement("", "hidden", "", event.ClientID),
		    clientName = addFormElement("Client Name", "text", "client_name", event.ClientName, regexpCheck(/.+/, "Client Name Required")),
		    other = addFormElement("Other Passengers", "text", "other", event.Other),
		    from = addFormElement("From", "textarea", "from", event.From, regexpCheck(/.+/, "From Address Required")),
		    to = addFormElement("To", "textarea", "to", event.To, regexpCheck(/.+/, "To Address Required"));
		addLister(clientName[1], function() {
			clientName[1].setInnerText("");
			stack.addLayer("clientList", function(client) {
				if (typeof client === "undefined") {
					return;
				}
				clientID.value = client.ID;
				clientName[0].value = client.Name;
				clientName[1].setInnerText("");
			});
			clientList(true);
		});
		autocomplete(rpc.autocompleteClientName, clientName[0], clientID);
		autocomplete(function(partial, callback) {
			rpc.autocompleteAddress(0, parseInt(clientID.value), partial, callback);
		}, from[0]);
		autocomplete(function(partial, callback) {
			rpc.autocompleteAddress(1, parseInt(clientID.value), partial, callback);
		}, to[0]);
		addFormSubmit((event.ID == 0) ? "Add Event" : "Edit Event", function() {
			var parts = [this, clientName[0], to[0], from[0], other[0]];
			parts.map(disableElement);
			event.ClientID = parseInt(clientID.value);
			event.Other = other[0].value;
			event.From = from[0].value;
			event.To = to[0].value;
			rpc.setEvent(event, function(resp) {
				if (resp.Errors) {
					clientName[1].setInnerText(resp.ClientError);
					from[1].setInnerText(resp.FromError);
					to[1].setInnerText(resp.ToError);
					driverTime[1].setInnerText(resp.TimeError);
					parts.map(enableElement);
				} else {
					event.ID = resp.ID;
					stack.removeLayer(event);
				}
			});
		});
		stack.setFragment();
	},
	addEvent = function(driver, startTime, endTime) {
		setEvent({
			"ID": 0,
			"Start": startTime.getTime(),
			"End": endTime.getTime(),
			"From": "",
			"To": "",
			"ClientID": 0,
			"ClientName": "",
			"Other": "",
			"DriverID": driver.ID,
			"DriverName": driver.Name,
			//"ClientRef": "",
			"InvoiceNote": "",
			"InvoiceFrom": "",
			"InvoiceTo": ""
		});
	},
	regexpCheck = function(regexp, error) {
		return function() {
			var errorDiv = document.getElementById("error_" + this.getAttribute("id"));
			if (this.value.match(regexp)) {
				errorDiv.removeChildren();
			} else {
				errorDiv.setInnerText(error);
			}
		}
	},
	dateCheck = regexpCheck(/^(0?[1-9]|1[0-9]|2[0-9]|3[01])\/(0?[1-9]|1[0-2])\/[0-9]{1,4}$/, "Please enter a valid date (DD/MM/YYYY)"),
	autocomplete = function(rpcCall, nameDiv, idDiv) {
		var autocompleteDiv = createElement("ul"),
		    cache = {},
		    clicker,
		    activator,
		    func = function(valUp, values){
			autocompleteDiv.removeChildren();
			var bounds = nameDiv.getBoundingClientRect();
			autocompleteDiv.style.left = Math.round(bounds.left + (window.pageXOffset || document.documentElement.scrollLeft || document.body.scrollLeft) - (document.documentElement.clientLeft || document.body.clientLeft || 0)) + "px";
			autocompleteDiv.style.top = Math.round(bounds.bottom + (window.pageYOffset || document.documentElement.scrollTop || document.body.scrollTop) - (document.documentElement.clientTop || document.body.clientTop || 0)) + "px";
			autocompleteDiv.style.width = (bounds.right - bounds.left) + "px";
			if (typeof idDiv !== "undefined") {
				idDiv.value = 0;
			}
			for (var i = 0; i < values.length; i++) {
				var li = autocompleteDiv.appendChild(createElement("li")),
				    value = values[i].Value,
				    startPos = value.toUpperCase().indexOf(valUp),
				    matchHighlight = createElement("b");
				if (typeof idDiv !== "undefined") {
					if (value.toUpperCase() === valUp) {
						idDiv.value = values[i].ID;
					}
				}
				li.appendChild(document.createTextNode(value.slice(0, startPos)));
				matchHighlight.appendChild(document.createTextNode(value.slice(startPos, startPos+valUp.length)));
				li.appendChild(matchHighlight);
				li.appendChild(document.createTextNode(value.slice(startPos+valUp.length)));
				li.addEventListener("mousedown", function(value, e) {
					e = e || event;
					if (e.button === 0) {
						clicker(value);
					}
				}.bind(null, values[i]));
				if (values[i].Disambiguation !== "") {
					var disambiguator = li.appendChild(createElement("div"));
					disambiguator.setInnerText(values[i].Disambiguation);
					disambiguator.setAttribute("class", "disambiguator");
					disambiguator.style.left = autocompleteDiv.style.width;
				}
			}
			layer.appendChild(autocompleteDiv);
		    };
		if (typeof idDiv !== "undefined") {
			clicker = function(val) {
				nameDiv.value = val.Value;
				idDiv.value = val.ID;
				if (autocompleteDiv.parentNode !== null) {
					autocompleteDiv.parentNode.removeChild(autocompleteDiv);
				}
			};
		} else {
			clicker = function(val) {
				nameDiv.value = val.Value;
				if (autocompleteDiv.parentNode !== null) {
					autocompleteDiv.parentNode.removeChild(autocompleteDiv);
				}
			};
		}
		autocompleteDiv.setAttribute("class", "autocompleter");
		nameDiv.addEventListener("blur", window.setTimeout.bind(window, function(e) {
			cache = {};
			if (autocompleteDiv.parentNode !== null) {
				autocompleteDiv.parentNode.removeChild(autocompleteDiv);
			}
		}, 100), false);
		activator = function() {
			var valUp = nameDiv.value.toUpperCase();
			if (autocompleteDiv.parentNode !== null) {
				autocompleteDiv.parentNode.removeChild(autocompleteDiv);
			}
			if (valUp.length === 0 && typeof idDiv !== "undefined") {
				return;
			}
			if (typeof cache[valUp] === "undefined") {
				rpcCall(valUp, function(values) {
					func(valUp, values);
					if (values.length > 0) {
						cache[valUp] = values;
					}
				});
			} else {
				func(valUp, cache[valUp]);
			}
		};
		nameDiv.addEventListener("keyup", activator, true);
		nameDiv.addEventListener("focus", activator, true);
	},
	/*noteJSON = function(text) {
		if (text.length === 0) {
			return {"Note":""};
		}
		if (text.charAt(0) !== "{") {
			return {"Note":text};
		}
		var data = JSON.parse(text);
		if (typeof data.Note === "undefined") {
			data.Note = "";
		}
		return data;
	},*/
	Date;
	(function() {
		var monthNames = ["January", "February", "March", "April", "May", "June", "July", "August", "September", "October", "November", "December"],
		    dayNames = ["Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"],
		    daysInMonth = [31, 0, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31],
		    suf = ["th", "st", "nd", "rd"],
		    argsToDateTime = function() {
			if (arguments.length === 0) {
				return oldDate.now();
			} else if (arguments.length === 1) {
				return arguments[0];
			}
			var ms = 0,
			    year = 1970,
			    month = 0,
			    day = 0,
			    hour = 0,
			    minute = 0,
			    seconds = 0,
			    milliseconds = 0,
			    daysInFebruary = 28;
			while (arguments[1] < 0) {
				year--;
				arguments[1] += 12;
			}
			if (arguments[0] < 1970) {
				for (year--; year >= arguments[0]; year--) {
					if (Date.prototype.isLeapYear(year)) {
						ms -= 31622400000;
					} else {
						ms -= 31536000000;
					}
				}
			} else {
				for (; year < arguments[0]; year++) {
					if (Date.prototype.isLeapYear(year)) {
						ms += 31622400000;
					} else {
						ms += 31536000000;
					}
				}
			}
			if (Date.prototype.isLeapYear(arguments[0])) {
				daysInFebruary = 29;
			}
			for (;month < arguments[1]; month++) {
				if (month % 12 === 1) {
					ms += daysInFebruary * 86400000;
				} else {
					ms += daysInMonth[month % 12] * 86400000;
				}
			}
			if (arguments.length > 2) {
				ms += (arguments[2] - 1) * 86400000;
			}
			if (arguments.length > 3) {
				ms += arguments[3] * 3600000;
			}
			if (arguments.length > 4) {
				ms += arguments[4] * 60000;
			}
			if (arguments.length > 5) {
				ms += arguments[5] * 1000;
			}
			if (arguments.length > 6) {
				ms += arguments[6];
			}
			return ms;
		    },
		    getYear = function(ms) {
			    var year = 1970;
			    if (ms < 0) {
				while (true) {
					year--;
					var msInYear = 31536000000;
					if (Date.prototype.isLeapYear(year)) {
						msInYear = 31622400000;
					}
					ms += msInYear;
					if (ms > 0) {
						return [year, ms];
					}
				}
			    } else {
				    while (true) {
					var msInYear = 31536000000;
					if (Date.prototype.isLeapYear(year)) {
						msInYear = 31622400000;
					}
					if (ms < msInYear) {
						return [year, ms];
					}
					year++;
					ms -= msInYear;
				    }
			    }
		    },
		    getMonth = function(ms) {
			    var ym = getYear(ms),
			        month = 0;
			    while (true) {
				var msInMonth;
				if (month === 1) {
					if (Date.prototype.isLeapYear(ym[0])) {
						msInMonth = 2505600000;
					} else {
						msInMonth = 2419200000;
					}
				} else {
					msInMonth = daysInMonth[month] * 86400000;
				}
				if (ym[1] < msInMonth) {
					return [month, ym[1]];
				}
				month++;
				ym[1] -= msInMonth;
			    }
		    };

		Date = function() {
			this._unixms = argsToDateTime.apply(null, arguments);
			Object.freeze(this);
		}

		Date.prototype = {
			getTime: function () {
				return this._unixms;
			},
			getFullYear: function() {
				return getYear(this._unixms)[0];
			},
			getMonth: function() {
				return getMonth(this._unixms)[0];
			},
			getDate: function() {
				return ((getMonth(this._unixms)[1] / 86400000)|0) + 1;
			},
			getDay: function() {
				return (((this._unixms / 86400000)|0) + 4) % 7;
			},
			getHours: function() {
				return ((this._unixms / 3600000)|0) % 24;
			},
			getMinutes: function() {
				return ((this._unixms / 60000)|0) % 60;
			},
			getSeconds: function() {
				return ((this._unixms / 1000)|0) % 60;
			},
			getMilliseconds: function() {
				return this._unixms % 1000;
			},
			getTimezoneOffset: function() {
				return 0;
			},
			isLeapYear: function(y) {
				if (typeof y === "undefined") {
					y = this.getFullYear();
				}
				return y % 4 === 0 && (y % 100 !== 0 || y % 400 === 0);
			},
			daysInMonth: function(y, m) {
				if (typeof y === "undefined") {
					y = this.getFullYear();
				}
				if (typeof m === "undefined") {
					m = this.getMonth();
				}
				while (m >= 12) {
					y++;
					m -= 12;
				}
				while (m < 0) {
					y--;
					m += 12;
				}
				if (m === 1) {
					if (this.isLeapYear(y)) {
						return 29;
					}
					return 28;
				}
				return daysInMonth[m]
			},
			getOrdinalSuffix: function(d) {
				if (typeof d === "undefined") {
					d = this.getDate();
				}
				var v = d % 100;
				return suf[(v - 20) % 10] || suf[v] || suf[0];
			},
			getMonthName: function(m) {
				if (typeof m === "undefined") {
					m = this.getMonth();
				} else if (m < 0 || m >= 12) {
					return "";
				}
				return monthNames[m];
			},
			getDayName: function(w) {
				if (typeof w === "undefined") {
					w = this.getDay();
				} else if (w < 0 || w >= 7) {
					return "";
				}
				return dayNames[w];
			},
			toDateString: function() {
				var year = this.getFullYear(),
				    month = this.getMonth() + 1,
				    date = this.getDate();
				if (month < 10) {
					month = "0" + month;
				}
				if (date < 10) {
					date = "0" + date;
				}
				return date + "/" + month + "/" + year;
			},
			toTimeString: function() {
				var hour = this.getHours(),
				    minutes = this.getMinutes();
				if (hour < 10) {
					hour = "0" + hour;
				}
				if (minutes < 10) {
					minutes = "0" + minutes;
				}
				return hour + ":" + minutes;
			},
			toOrdinalDate: function() {
				var year = this.getFullYear(),
				    month = this.getMonth() + 1,
				    date = this.getDate(),
				    suffix = this.getOrdinalSuffix(date);
				return date + suffix + " " + monthNames[month] + " " + year;
			},
			toLocaleString: function() {
				var year = this.getFullYear(),
				    month = this.getMonth() + 1,
				    date = this.getDate(),
				    hour = this.getHours(),
				    minutes = this.getMinutes();
				if (month < 10) {
					month = "0" + month;
				}
				if (date < 10) {
					date = "0" + date;
				}
				if (hour < 10) {
					hour = "0" + hour;
				}
				if (minutes < 10) {
					minutes = "0" + minutes;
				}
				return hour + ":" + minutes + " " + date + "/" + month + "/" + year;
			},
			toString: function() {
				return this.getDayName() + ", " + this.getDate() + this.getOrdinalSuffix() + " of " + this.getMonthName() + ", " + this.getFullYear() +" @ " + this.getHours() + ":" + this.getMinutes() + ":" + this.getSeconds();
			},
		};
	}());
	Element.prototype.removeChildren = (function() {
		var docFrag = document.createDocumentFragment();
		return function(filter) {
			if (typeof filter === "function") {
				while (this.hasChildNodes()) {
					if (filter(this.firstChild)) {
						this.removeChild(this.firstChild);
					} else {
						docFrag.appendChild(this.firstChild);
					}
				}
				this.appendChild(docFrag);
			} else {
				while (this.hasChildNodes()) {
					this.removeChild(this.lastChild);
				}
			}
		};
	}());
	Element.prototype.getElementById = function(id) {
		return this.querySelector("#" + id);
	};
	Element.prototype.getChildElementById = function(id) {
		for (var i = 0; i < this.childNodes.length; i++) {
			if (this.childNodes[i].getAttribute("id") === id) {
				return this.childNodes[i];
			}
		}
		return null;
	};
	Element.prototype.setInnerText = function(text) {
		this.removeChildren();
		this.appendChild(document.createTextNode(text));
		return this;
	};
	Element.prototype.setPreText = function(text) {
		this.removeChildren();
		var parts = text.split("\n"),
		    i = 0;
		for (; i < parts.length; i++) {
			if (i > 0) {
				this.appendChild(createElement("br"));
			}
			this.appendChild(document.createTextNode(parts[i]));
		}
		return this;
	};
	String.prototype.getWidth = (function (){
		var canvas = document.createElement("canvas");
		return function(font) {
			var ctx = canvas.getContext("2d");
			ctx.font = font;
			return ctx.measureText(this).width;
		};
	}());
	Number.prototype.formatMoney = function(amount) {
		amount = amount || this;
		var toRet = "",
		    integer = +amount | 0 + "",
		    fract = 0;
		if (amount < 0) {
			toRet = "-";
			amount = -amount;
		}
		fract = amount - (amount | 0);
		while (integer.length > 3) {
			toRet += "," + integer.substr(0, 3);
			integer = integer.substr(3);
		}
		toRet += integer + "." + fract.toFixed(2).substr(2);
		return toRet;
	};
	Number.prototype.formatColour = function(num) {
		num = num || this;
		var colourStr = num.toString(16);
		while (colourStr.length < 6) {
			colourStr = "0" + colourStr;
		}
		return "#" + colourStr;
	}
}.bind(null, Date));

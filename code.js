"use strict";
window.addEventListener("load", function(oldDate) {
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
		this.drivers       = request.bind(this, "Drivers", null);          // callback
		this.companies     = request.bind(this, "Companies", null);        // callback
		this.clients       = request.bind(this, "Clients", null);          // callback
		this.unsentMessages = request.bind(this, "UnsentMessages", null);  // callback
		this.getEventsWithDriver = function(driverID, start, end, callback) {
			request("DriverEvents", {"DriverID": driverID, "Start": start, "End": end}, callback);
		}
		this.autocompleteAddress = function(priority, partial, callback) {
			request("AutocompleteAddress", {"Priority": priority, "Partial": partial}, callback);
		}
		this.autocompleteCompanyName = request.bind(this, "AutocompleteCompanyName") // partial, callback
		this.autocompleteClientName = request.bind(this, "AutocompleteClientName")   // partial, callback
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
			var outerLayer = createElement("div"),
			    cancelButton = createElement("div");
			outerLayer.style.zIndex = stack.length + 1;
			outerLayer.className = "layer";
			layer = createElement("div");
			layer.setAttribute("id", layerID);
			if (stack.length > 1) {
				cancelButton.setAttribute("class", "canceller");
				cancelButton.innerHTML = "X";
				cancelButton.addEventListener("click", this.removeLayer.bind(this, undefined));
			}
			layer.appendChild(cancelButton);
			outerLayer.appendChild(layer);
			body.appendChild(outerLayer);
		};
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
				layer = document.createDocumentFragment();
			}
		};
		this.setFragment = function () {
			if (typeof layer == "object" && layer.nodeType === 11) {
				var firstChild = body.lastChild.getElementsByTagName("div")[0];
				firstChild.appendChild(layer);
				layer = firstChild;
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
		adder.innerHTML = "+";
		adder.addEventListener("click", callback);
		adder.setAttribute("class", "adder");
		layer.insertBefore(adder, elementBefore);
	},
	dateTimeFormat = function(date) {
		return (new Date(date)).toLocaleString('en-GB');
	},
	events = new (function() {
		var dateTime,
		    dateShift,
		    eventList = createElement("div"),
		    drivers = [],
		    days = {},
		    startEnd = [dateShift, dateShift],
		    plusDriver = createElement("div"),
		    nextDriverPos = 100,
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
			    mins = (unix - dateShift) / 60000,
			    minOnScreen = unix - ((screenWidth / 2) * 60000),
			    maxOnScreen = unix + ((screenWidth / 2) * 60000),
			    minOnScreenDayStart = minOnScreen - (minOnScreen % 86400000) - 86400000,
			    maxOnScreenDayEnd = maxOnScreen - (maxOnScreen % 86400000) + 2 * 86400000,
			    tDate, year, month, day, t,
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
				var node = days[keys[t]];
				if (node.parentNode !== null) {
					var parts = keys[t].split("_");
					unix = (new Date(parts[0], parts[1], parts[2])).getTime();
					if (unix < minOnScreenDayStart || unix > maxOnScreenDayEnd) {
						eventList.removeChild(days[keys[t]]);
					}
				}
			}
			for (t = minOnScreenDayStart; t < maxOnScreenDayEnd; t += 86400000) {
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
			var monthDate = new Date(year, month),
			    monthDiv = createElement("div"),
			    textDiv = monthDiv.appendChild(createElement("div")),
			    monthEnclosure = createElement("div");
			textDiv.innerHTML = monthDate.getMonthName();
			textDiv.setAttribute("class", "slider");
			monthDiv.setAttribute("class", "month");
			monthDiv.setAttribute("id", "month_" + year + "_" + month);
			monthDiv.style.left = timeToPos(monthDate);
			monthDiv.style.width = (monthDate.daysInMonth() * 24 * 60) + "px";
			eventList.appendChild(monthDiv);
		    },
		    addDay = function(year, month, day) {
			if (typeof days[year + "_" + month + "_" + day] !== "undefined") {
				eventList.appendChild(days[year + "_" + month + "_" + day]);
				return;
			} else if (document.getElementById("month_" + year + "_" + month) === null) {
				addMonth(year, month);
			}
			var dayDate = new Date(year, month, day),
			    dayDiv = createElement("div"),
			    dayEnclosure = createElement("div"),
			    textDiv = dayDiv.appendChild(createElement("div")),
			    i = 0;
			textDiv.innerHTML = dayDate.getDayName() + ", " + day + dayDate.getOrdinalSuffix();
			textDiv.setAttribute("class", "slider");
			dayDiv.setAttribute("class", "day");
			dayDiv.setAttribute("id", "day_" + year + "_" + month + "_" + day);
			dayDiv.style.left = timeToPos(dayDate);
			dayEnclosure.appendChild(dayDiv);
			dayEnclosure.setAttribute("class", "dayEnclosure");
			days[year + "_" + month + "_" + day] = dayEnclosure;
			for (; i < 24; i++) {
				addHour(year, month, day, i);
			}
			eventList.appendChild(dayEnclosure);
			return true;
		    },
		    addHour = function(year, month, day, hour) {
			var hourDate = new Date(year, month, day, hour),
			    hourDiv = createElement("div");
			hourDiv.setAttribute("class", "hour");
			hourDiv.innerHTML = formatNum(hour);
			hourDiv.style.left = timeToPos(hourDate);
			days[year + "_" + month + "_" + day].appendChild(hourDiv);
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
			    leftPos = timeToPos(fifteenDate);
			fifteenDiv.setAttribute("class", "minute");
			fifteenDiv.innerHTML = formatNum(block * 15);
			fifteenDiv.style.left = leftPos;
			dayDiv.appendChild(fifteenDiv);
			for (var i = 0; i < driverIDs.length; i++) {
				cellDiv = createElement("div");
				cellDiv.setAttribute("class", "eventCell " + (block % 2 == i % 2 ? "cellOdd" : "cellEven"));
				cellDiv.setAttribute("id", "cell_" + driverIDs[i] + "_" + year + "_" + month + "_" + day + "_" + hour + "_" + block);
				cellDiv.style.left = leftPos;
				cellDiv.style.top = drivers[driverIDs[i]].yPos + "px";
				cellDiv.addEventListener("mouseover", eventOnMouseOver);
				cellDiv.addEventListener("mouseout", eventOnMouseOut);
				cellDiv.addEventListener("click", eventOnClick);
				dayDiv.appendChild(cellDiv);
			}
		    },
		    isOnScreen = function(div) {
			var left = parseInt(eventList.style.left, 10) + parseInt(div.style.left, 10),
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
			var now = new Date(),
			    addToBar = function() {
				var topBar = layer.appendChild(createElement("div"));
				topBar.setAttribute("id", "topBar");
				return function(text, callback) {
					var item = topBar.appendChild(createElement("div"));
					item.innerHTML = text;
					item.setAttribute("class", "simpleButton");
					item.addEventListener("click", callback);
				};
			    }(),
			    params = window.location.search.substring(1).split("&"), i = 0, paramParts;
			for (; i < params.length; i++) {
				paramParts = params[i].split("=");
				if (paramParts.length === 2 && paramParts[0] === "date") {
					now = new Date(parseInt(paramParts[1]));
				}
			}
			addToBar("Companies", companyList);
			addToBar("Clients", clientList);
			addToBar("Messages", messageList);
			dateShift = now.getTime();
			rpc.drivers(function(ds) {
				plusDriver.appendChild(createElement("div")).innerHTML = "+";
				plusDriver.setAttribute("id", "plusDriver");
				plusDriver.setAttribute("class", "simpleButton");
				plusDriver.addEventListener("click", function() {
					stack.addLayer("addDriver", this.addDriver.bind(this));
					addDriver();
				}.bind(this));
				for (var i = 0; i < ds.length; i++) {
					this.addDriver(ds[i]);
					//drivers[ds[i].ID] = ;
				}
				layer.appendChild(plusDriver);
				layer.appendChild(eventList).setAttribute("class", "events slider");
				for (i = 0; i < 10; i++) {
					var div = layer.appendChild(createElement("div"));
					if (i % 2 === 0) {
						div.appendChild(createElement("div")).innerHTML = "&lt;";
						div.setAttribute("class", "moveLeft simpleButton");
					} else {
						div.appendChild(createElement("div")).innerHTML = "&gt;";
						div.setAttribute("class", "moveRight simpleButton");
					}
					div.style.top = 20 + Math.floor(i / 2) * 20 + "px";
					div.addEventListener("click", moveHandler(i));
				}
				stack.setFragment();
				update(now);
				window.addEventListener("resize", update.bind(this, undefined));
			}.bind(this));
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
			return new Date(parts[2], parts[3], parts[4], parts[5], parts[6] * 15);
		    },
		    getEventsBetween = function(id) {
			if (eventSelected === null) {
				return null;
			}
			var thatID = eventSelected.getAttribute("id"),
			    thisDriverID = cellIdToDriver(id),
			    thatDriverID = cellIdToDriver(thatID),
			    thisTime = cellIdToDate(id).getTime(),
			    thatTime = cellIdToDate(thatID).getTime(),
			    events = [];
			if (thisDriverID !== thatDriverID || thisTime <= thatTime || thisTime - thatTime > 86400000) {
				return null;
			}
			for (var t = thatTime + 900000; t <= thisTime; t += 900000) {
				var tDate = new Date(t),
				    year = tDate.getFullYear(),
				    month = tDate.getMonth(),
				    day = tDate.getDate(),
				    hour = tDate.getHours(),
				    block = tDate.getMinutes() / 15,
				    cell = days[year + "_" + month + "_" + day].querySelector("#cell_" + thisDriverID + "_" + year + "_" + month + "_" + day + "_" + hour + "_" + block);
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
			if (e.target === eventSelected) {
				return;
			}
			if (eventSelected !== null) {
				if (cellIdToDriver(e.target.getAttribute("id")) === cellIdToDriver(eventSelected.getAttribute("id"))) {
					var cells = getEventsBetween(e.target.getAttribute("id"));
					if (cells === null) {
						return;
					}
					for (var i = 0; i < cells.length; i++) {
						changeThirdCellClass(cells[i], "eventsInBetween");
					}
					eventsHighlighted = cells;
				}
				return;
			}
			changeThirdCellClass(e.target, "eventHover");
			eventsHighlighted = [e.target];
		    },
		    eventOnMouseOut = function() {
			for (var i = 0; i < eventsHighlighted.length; i++) {
				changeThirdCellClass(eventsHighlighted[i], "");
			}
			eventsHighlighted = [];
		    },
		    eventSelected = null,
		    eventsHighlighted = [],
		    eventOnClick = function(e) {
			e = e || eventsent;
			if (e.target === eventSelected) {
				eventSelected = null;
				changeThirdCellClass(e.target, "eventHover");
				eventsHighlighted.push(e.target);
			} else if (eventSelected === null) {
				eventSelected = e.target;
				changeThirdCellClass(e.target, "eventSelected");
				eventsHighlighted = [];
			} else if (getEventsBetween(e.target.getAttribute("id")) !== null){
				eventsHighlighted.push(eventSelected);
				var id = e.target.getAttribute("id");
				stack.addLayer("addEvent", this.addEvent.bind(this));
				addEvent(drivers[cellIdToDriver(id)], cellIdToDate(eventSelected.getAttribute("id")), cellIdToDate(id));
				eventSelected = null;
			}
		    }.bind(this);
		this.addDriver = function(d) {
			if (typeof d === "undefined") {
				return;
			}
			drivers[d.ID] = d;
			drivers[d.ID].yPos = nextDriverPos;
			drivers[d.ID].events = [];
			var dDiv = createElement("div"),
			    t;
			dDiv.appendChild(createElement("div")).innerHTML = d.Name;
			dDiv.setAttribute("class", "driverName simpleButton");
			dDiv.setAttribute("id", "driver_" + d.ID);
			dDiv.addEventListener("click", function() {
				stack.addLayer("viewDriver");
				viewDriver(drivers[d.ID]);
			});
			dDiv.style.top = (nextDriverPos + 20) + "px";
			nextDriverPos += 100;
			plusDriver.style.top = (nextDriverPos + 20) + "px";
			layer.appendChild(dDiv);
			var keys = Object.keys(days),
			    oddEven = Object.keys(drivers).length % 2;
			for (var i = 0; i < keys.length; i++) {
				var parts = keys[i].split("_"),
				    year = parts[0],
				    month = parts[1],
				    day = parts[2],
				    dayDiv = days[keys[i]];
				for (var hour = 0; hour < 24; hour++) {
					for (var block = 0; block < 4; block++) {
						var cellDiv = createElement("div");
						cellDiv.setAttribute("class", "eventCell " + (block % 2 !== oddEven ? "cellOdd" : "cellEven"));
						cellDiv.setAttribute("id", "cell_" + d.ID + "_" + year + "_" + month + "_" + day + "_" + hour + "_" + block);
						cellDiv.style.left = timeToPos(new Date(year, month, day, hour, block * 15));
						cellDiv.style.top = drivers[d.ID].yPos + "px";
						cellDiv.addEventListener("mouseover", eventOnMouseOver);
						cellDiv.addEventListener("mouseout", eventOnMouseOut);
						cellDiv.addEventListener("click", eventOnClick);
						dayDiv.appendChild(cellDiv);
					}
				}
			}
		};
		this.updateDriver = function(d) {
			document.getElementById("driver_" + d.ID).getElementsByTagName("div")[0].innerHTML = d.Name;
			d.events = drivers[d.ID].events;
			d.yPos = drivers[d.ID].yPos;
			//for (var i = 0; i < d.events.length; i++) {
			//	d.events[i].DriverName = d.Name;
			//}
			drivers[d.ID] = d;
		};
		this.removeDriver = function(d) {
			window.location.search = "?date="+dateTime.getTime();
		};
		this.addEvent = function(e) {
			if (typeof e === "undefined") {
				return;
			}
			drivers[e.DriverID].events[e.Start] = e;

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
	showCompany = function(company) {
		stack.addLayer("showCompany");
		alert(company.Name);
	},
	companyList = function() {
		stack.addLayer("companies");
		rpc.companies(function(companies) {
			stack.addFragment();
			var title = layer.appendChild(createElement("h1")),
			    table = createElement("table"),
			    headerRow = table.appendChild(createElement("tr")),
			    addCompanyToTable = function(company) {
				if (typeof company === "undefined") {
					return;
				}
				var row = createElement("tr"),
				    nameCell = row.appendChild(createElement("td"));
				nameCell.innerHTML = company.Name;
				nameCell.setAttribute("class", "simpleButton");
				nameCell.addEventListener("click", showCompany.bind(null, company));
				row.appendChild(createElement("td")).innerHTML = company.Address;
				table.appendChild(row);
			    };
			title.innerHTML = "Companies";
			addAdder(null, function() {
				stack.addLayer("addCompany", addCompanyToTable);
				addCompany();
			});
			headerRow.appendChild(createElement("th")).innerHTML = "Company Name";
			headerRow.appendChild(createElement("th")).innerHTML = "Address";
			companies.map(addCompanyToTable);
			layer.appendChild(table);
			stack.setFragment();
		});
	},
	showClient = function(client, company) {
		stack.addLayer("showClient");
		alert(client.Name);
	},
	clientList = function() {
		stack.addLayer("clients");
		rpc.clients(function(clients) {
			stack.addFragment()
			var title = layer.appendChild(createElement("h1")),
			    table = createElement("table"),
			    headerRow = table.appendChild(createElement("tr")),
			    companies = [],
			    addClientToTable = function(client) {
				if (typeof client === "undefined") {
					return;
				}
				var row = createElement("tr"),
				    nameCell = row.appendChild(createElement("td")),
				    companyCell = row.appendChild(createElement("td")),
				    setCompanyCell = function() {
					companyCell.innerHTML = companies[client.CompanyID].Name;
					companyCell.setAttribute("class", "simpleButton");
					companyCell.addEventListener("click", showCompany.bind(null, companies[client.CompanyID]));
				    };
				nameCell.innerHTML = client.Name;
				nameCell.setAttribute("class", "simpleButton");
				nameCell.addEventListener("click", showClient.bind(null, client));
				if (typeof companies[client.CompanyID] !== "undefined") {
					setCompanyCell();
				} else {
					rpc.getCompany(client.CompanyID, function(company) {
						if (typeof company === "undefined") {
							companyCell.innerHTML = "Error!";
							return;
						}
						companies[company.ID] = company;
						setCompanyCell();
					});
				}
				row.appendChild(createElement("td")).innerHTML = client.PhoneNumber;
				row.appendChild(createElement("td")).innerHTML = client.Reference;
				table.appendChild(row);
			    };
			title.innerHTML = "Clients";
			addAdder(null, function() {
				stack.addLayer("addClient", addClientToTable);
				addClient();
			});
			headerRow.appendChild(createElement("th")).innerHTML = "Client Name";
			headerRow.appendChild(createElement("th")).innerHTML = "Company Name";
			headerRow.appendChild(createElement("th")).innerHTML = "Phone Number";
			headerRow.appendChild(createElement("th")).innerHTML = "Reference";
			clients.map(addClientToTable);
			layer.appendChild(table);
			stack.setFragment();
		});
	},
	messageList = function() {
		stack.addLayer("messages");
		layer.appendChild(createElement("h1")).innerHTML = "Messages";
	},
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
		} else {
			input = createElement("input");
			input.setAttribute("type", type);
		}
		input.setAttribute("value", contents);
		input.setAttribute("id", id);
		if (type === "hidden") {
			return layer.appendChild(input);
		}
		label.innerHTML = name;
		if (id === "") {
			input.setAttribute("readonly", "readonly");
		}
		label.setAttribute("for", id);
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
	setDriver = function(driver) {
		stack.addFragment();
		addTitle(driver.ID, "Add Driver", "Edit Driver");
		var driverName = addFormElement("Driver Name", "text", "driver_name", driver.Name, regexpCheck(/.+/, "Please enter a valid name")),
		    regNumber = addFormElement("Registration Number", "text", "driver_reg", driver.RegistrationNumber, regexpCheck(/[a-zA-Z0-9 ]+/, "Please enter a valid Vehicle Registration Number")),
		    phoneNumber = addFormElement("Phone Number", "text", "driver_phone", driver.PhoneNumber, regexpCheck(/^(0|\+?44)[0-9 ]{10}$/, "Please enter a valid mobile telephone number")),
		    submit = function() {
			var parts = [this, driverName[0], regNumber[0], phoneNumber[0]];
			parts.map(disableElement);
			driver.Name = driverName[0].value;
			driver.RegistrationNumber = regNumber[0].value;
			driver.PhoneNumber = phoneNumber[0].value;
			rpc.setDriver(driver, function(resp) {
				if (resp.Errors) {
					driverName[1].innerHTML = resp.NameError;
					regNumber[1].innerHTML = resp.RegError;
					phoneNumber[1].innerHTML = resp.PhoneError;
					parts.map(enableElement);
				} else {
					driver.ID = resp.ID;
					stack.removeLayer(driver);
				}
			});
		};
		addFormSubmit("Add Driver", submit);

		stack.setFragment();
	},
	addDriver = function() {
		setDriver({
			"ID": 0,
			"Name": "",
			"RegistrationNumber": "",
			"PhoneNumber": "",
		});
	},
	setClient = function(client) {
		stack.addFragment();
		addTitle(client.ID, "Add Client", "Edit Client");
		var clientName = addFormElement("Client Name", "text", "client_name", client.Name, regexpCheck(/.+/, "Please enter a valid name")),
		    companyID = addFormElement("", "hidden", "client_company_id", client.CompanyID),
		    companyName = addFormElement("Company Name", "text", "client_company_name", client.CompanyName, regexpCheck(/.+/, "Please enter a valid name")),
		    clientPhone = addFormElement("Mobile Number", "text", "client_phone", client.PhoneNumber, regexpCheck(/^(0|\+?44)[0-9 ]{10}$/, "Please enter a valid mobile telephone number")),
		    clientRef = addFormElement("Client Ref", "text", "client_ref", client.Reference, regexpCheck(/.+/, "Please enter a reference code"));
		addAdder(companyName[1], function() {
			stack.addLayer("addCompany", function(company) {
				if (typeof company === "undefined") {
					return;
				}
				companyID.value = company.ID;
				companyName[0].value = company.Name;
				companyName[1].innerHTML = "";
			});
			addCompany();
		});
		autocomplete(rpc.autocompleteCompanyName, companyName[0], companyID);
		addFormSubmit("Add Client", function() {
			var parts = [this, clientName[0], companyName[0], clientPhone[0], clientRef[0]];
			parts.map(disableElement);
			client.Name = clientName[0].value;
			client.CompanyID = parseInt(companyID.value);
			client.PhoneNumber = clientPhone[0].value;
			client.Reference = clientRef[0].value;
			rpc.setClient(client, function (resp) {
				if (resp.Errors) {
					clientName[1].innerHTML = resp.NameError;
					companyName[1].innerHTML = resp.CompanyError;
					clientPhone[1].innerHTML = resp.PhoneError;
					clientRef[1].innerHTML = resp.ReferenceError;
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
		addTitle(company.ID, "Add Company", "Edit Company");
		var companyName = addFormElement("Company Name", "text", "company_name", company.Name, regexpCheck(/.+/, "Please enter a valid name")),
		    address = addFormElement("Company Address", "textarea", "company_address", company.Address, regexpCheck(/.+/, "Please enter a valid address"));
		addFormSubmit("Add Company", function() {
			var parts = [this, companyName[0], address[0]];
			parts.map(disableElement);
			company.Name = companyName[0].value;
			company.Address = address[0].value;
			rpc.setCompany(company, function(resp) {
				if (resp.Errors) {
					companyName[1].innerHTML = resp.NameError;
					address[1].innerHTML = resp.AddressError;
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
		});
	},
	setEvent = (function() {
		var fromAddressRPC = rpc.autocompleteAddress.bind(rpc, 0),
		    toAddressRPC = rpc.autocompleteAddress.bind(rpc, 1);
		return function(event) {
			stack.addFragment();
			addTitle(event.ID, "Add Event", "Edit Event");
			addFormElement("Driver", "text", "", event.DriverName);
			addFormElement("Start", "text", "", dateTimeFormat(event.Start));
			addFormElement("End", "text", "", dateTimeFormat(event.End));
			var from = addFormElement("From", "textarea", "from", event.From),
			    to = addFormElement("To", "textarea", "to", event.To),
			    clientID = addFormElement("", "hidden", "", event.ClientID),
			    clientName = addFormElement("Client Name", "text", "client_name", event.ClientName);
			addAdder(clientName[1], function() {
				stack.addLayer("addClient", function(client) {
					if (typeof client === "undefined") {
						return;
					}
					clientID.value = client.ID;
					clientName[0].value = client.Name;
					clientName[1].innerHTML = "";
				});
				addClient();
			});
			autocomplete(fromAddressRPC, from[0]);
			autocomplete(toAddressRPC, to[0]);
			autocomplete(rpc.autocompleteClientName, clientName[0], clientID);
			addFormSubmit("Add Event", function() {
				var parts = [this, clientName[0], to[0], from[0]];
				parts.map(disableElement);
				event.ClientID = parseInt(clientID.value);
				event.From = from[0].innerHTML;
				event.To = to[0].innerHTML;
				rpc.setEvent(event, function(resp) {
					if (resp.Errors) {
						clientName[1].innerHTML = resp.ClientError;
						from[1].innerHTML = resp.FromError;
						to[1].innerHTML = resp.ToError;
						parts.map(enableElement);
					} else {
						event.ID = resp.ID;
						stack.removeLayer(event);
					}
				});
			});
			stack.setFragment();
		}
	}()),
	addEvent = function(driver, startTime, endTime) {
		setEvent({
			"ID": 0,
			"Start": startTime.getTime(),
			"End": endTime.getTime(),
			"From": "",
			"To": "",
			"ClientID": 0,
			"ClientName": "",
			"DriverID": driver.ID,
			"DriverName": driver.Name,
		});
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
	autocomplete = function(rpcCall, nameDiv, idDiv) {
		var autocompleteDiv = createElement("ul"),
		    cache = {},
		    clicker,
		    func = function(valUp, values){
			while (autocompleteDiv.hasChildNodes()) {
				autocompleteDiv.removeChild(autocompleteDiv.lastChild);
			}
			for (var i = 0; i < values.length; i++) {
				var li = autocompleteDiv.appendChild(createElement("li")),
				    value = values[i].Value,
				    startPos = value.toUpperCase().indexOf(valUp),
				    matchHighlight = createElement("b");
				if (typeof idDiv !== "undefined") {
					if (value.toUpperCase() === valUp) {
						idDiv.value = values[i].ID;
					} else {
						idDiv.value = 0;
					}
				}
				li.appendChild(document.createTextNode(value.slice(0, startPos)));
				matchHighlight.appendChild(document.createTextNode(value.slice(startPos, startPos+valUp.length)));
				li.appendChild(matchHighlight);
				li.appendChild(document.createTextNode(value.slice(startPos+valUp.length)));
				li.addEventListener("click", clicker.bind(null, values[i]));
			}

			var bounds = nameDiv.getBoundingClientRect();
			autocompleteDiv.style.left = Math.round(bounds.left + (window.pageXOffset || document.documentElement.scrollLeft || document.body.scrollLeft) - (document.documentElement.clientLeft || document.body.clientLeft || 0)) + "px";
			autocompleteDiv.style.top = Math.round(bounds.bottom + (window.pageYOffset || document.documentElement.scrollTop || document.body.scrollTop) - (document.documentElement.clientTop || document.body.clientTop || 0)) + "px";
			autocompleteDiv.style.width = (bounds.right - bounds.left) + "px";
			layer.appendChild(autocompleteDiv);
		    };
		if (typeof idDiv !== "undefined") {
			clicker = function(val) {
				nameDiv.value = val.Value;
				idDiv.value = val.ID;
				layer.removeChild(autocompleteDiv);
			};
		} else {
			clicker = function(val) {
				nameDiv.value = val.Value;
				layer.removeChild(autocompleteDiv);
			};
		}
		autocompleteDiv.setAttribute("class", "autocompleter");
		nameDiv.addEventListener("blur", window.setTimeout.bind(window, function(e) {
			cache = {};
			if (autocompleteDiv.parentNode !== null) {
				autocompleteDiv.parentNode.removeChild(autocompleteDiv);
			}
		}, 100), false);
		nameDiv.addEventListener("keyup", function() {
			var valUp = nameDiv.value.toUpperCase();
			if (autocompleteDiv.parentNode !== null) {
				autocompleteDiv.parentNode.removeChild(autocompleteDiv);
			}
			if (valUp.length === 0) {
				return;
			}
			if (typeof cache[valUp] === "undefined") {
				rpcCall(valUp, function(values) {
					func(valUp, values);
					cache[valUp] = values;
				});
			} else {
				func(valUp, cache[valUp]);
			}
		}, true);
	},
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
				return year + "/" + month + "/" + date + " " + hour + ":" + minutes;
			},
			toString: function() {
				return this.getDayName() + ", " + this.getDate() + this.getOrdinalSuffix() + " of " + this.getMonthName() + ", " + this.getFullYear() +" @ " + this.getHours() + ":" + this.getMinutes() + ":" + this.getSeconds();
			},
		};
	}());
}.bind(null, Date));

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
			document.body.setInnerText("An error occurred");
		}
		ws.onclose = function(event) {
			if (event.code !== 1000) {
				document.body.setInnerText("Lost Connection To Server! Code: " + event.code);
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
		this.clientsForCompany = request.bind(this, "ClientsForCompany"); // id, callback
		this.getEventsWithDriver = function(driverID, start, end, callback) {
			request("DriverEvents", {"ID": driverID, "Start": start, "End": end}, callback);
		};
		this.getEventsWithClient = function(clientID, start, end, callback) {
			request("ClientEvents", {"ID": clientID, "Start": start, "End": end}, callback);
		};
		this.getEventsWithCompany = function(companyID, start, end, callback) {
			request("CompanyEvents", {"ID": companyID, "Start": start, "End": end}, callback);
		};
		this.autocompleteAddress = function(priority, partial, callback) {
			request("AutocompleteAddress", {"Priority": priority, "Partial": partial}, callback);
		};
		this.autocompleteCompanyName = request.bind(this, "AutocompleteCompanyName"); // partial, callback
		this.autocompleteClientName = request.bind(this, "AutocompleteClientName");   // partial, callback
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
	events = new (function() {
		var dateTime,
		    dateShift,
		    driverEvents = createElement("div"),
		    eventCells = driverEvents.appendChild(createElement("div")),
		    dates = createElement("div"),
		    drivers = [],
		    days = {},
		    startEnd = [dateShift, dateShift],
		    plusDriver = driverEvents.appendChild(createElement("div")),
		    nextDriverPos = 0,
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
					for (i = 0; i < driverIDs.length; i++) {
						rpc.getEventsWithDriver(parseInt(driverIDs[i]), tDate.getTime(), tDate.getTime() + 86400000, function(events) {
							for(var i = 0; i < events.length; i++) {
								addEventToTable(events[i]);
							}
						});
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
				return;
			} else if (document.getElementById("month_" + year + "_" + month) === null) {
				addMonth(year, month);
			}
			var dayDate = new Date(year, month, day),
			    dayDiv = createElement("div"),
			    dayEnclosure = createElement("div"),
			    i = 0;
			dayDiv.appendChild(createElement("div")).setInnerText(dayDate.getDayName() + ", " + day + dayDate.getOrdinalSuffix()).setAttribute("class", "slider");
			dayDiv.setAttribute("class", "day");
			dayDiv.setAttribute("id", "day_" + year + "_" + month + "_" + day);
			dayDiv.style.left = timeToPos(dayDate);
			dayEnclosure.appendChild(dayDiv);
			dayEnclosure.setAttribute("class", "dayEnclosure");

			days[year + "_" + month + "_" + day] = [dayEnclosure, eventCells.appendChild(createElement("div"))];
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
			    leftPos = timeToPos(fifteenDate);
			fifteenDiv.setAttribute("class", "minute");
			fifteenDiv.setAttribute("id", "minute_" + year + "_" + month + "_" + day + "_" + hour + "_" + block);
			fifteenDiv.setInnerText(formatNum(block * 15));
			fifteenDiv.style.left = leftPos;
			dayDiv[0].appendChild(fifteenDiv);
			for (var i = 0; i < driverIDs.length; i++) {
				cellDiv = createElement("div");
				cellDiv.setAttribute("class", "eventCell " + (block % 2 == i % 2 ? "cellOdd" : "cellEven"));
				cellDiv.setAttribute("id", "cell_" + driverIDs[i] + "_" + year + "_" + month + "_" + day + "_" + hour + "_" + block);
				cellDiv.style.left = leftPos;
				cellDiv.style.top = drivers[driverIDs[i]].yPos + "px";
				cellDiv.addEventListener("mouseover", eventOnMouseOver);
				cellDiv.addEventListener("mouseover", fifteenDiv.setAttribute.bind(fifteenDiv, "class", "minute select"));
				cellDiv.addEventListener("mouseout", eventOnMouseOut);
				cellDiv.addEventListener("mouseout", fifteenDiv.setAttribute.bind(fifteenDiv, "class", "minute"));
				cellDiv.addEventListener("click", eventOnClick);
				dayDiv[1].appendChild(cellDiv);
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
			var now = new Date(),
			    addToBar = function() {
				var topBar = layer.appendChild(createElement("div"));
				topBar.setAttribute("id", "topBar");
				return function(text, callback) {
					var item = topBar.appendChild(createElement("div"));
					item.setInnerText(text);
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
			addToBar("Companies", function() {
				stack.addLayer("companyList");
				companyList();
			});
			addToBar("Clients", function() {
				stack.addLayer("clientList");
				clientList();
			});
			addToBar("Messages", messageList);
			dateShift = now.getTime();
			rpc.drivers(function(ds) {
				plusDriver.appendChild(createElement("div")).setInnerText("+");
				plusDriver.setAttribute("id", "plusDriver");
				plusDriver.setAttribute("class", "simpleButton");
				plusDriver.addEventListener("click", function() {
					stack.addLayer("addDriver", this.addDriver.bind(this));
					addDriver();
				}.bind(this));
				for (var i = 0; i < ds.length; i++) {
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
			return new Date(parts[2], parts[3], parts[4], parts[5], parts[6] * 15).getTime();
		    },
		    getEventsBetween = function(id) {
			if (eventSelected === null) {
				return null;
			}
			var thatID = eventSelected.getAttribute("id"),
			    thisDriverID = cellIdToDriver(id),
			    thatDriverID = cellIdToDriver(thatID),
			    thisTime = cellIdToDate(id),
			    thatTime = cellIdToDate(thatID),
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
				    cell = days[year + "_" + month + "_" + day][1].getElementById("cell_" + thisDriverID + "_" + year + "_" + month + "_" + day + "_" + hour + "_" + block);
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
				stack.addLayer("addEvent", addEventToTable);
				addEvent(drivers[cellIdToDriver(id)], new Date(cellIdToDate(eventSelected.getAttribute("id"))), new Date(cellIdToDate(id) + 900000));
				eventSelected = null;
			}
		    }.bind(this),
		    addEventToTable = function(e) {
			if (typeof e === "undefined") {
				return;
			}
			drivers[e.DriverID].events[e.Start] = e;
			var eventDate = new Date(e.Start),
			    year = eventDate.getFullYear(),
			    month = eventDate.getMonth(),
			    day = eventDate.getDate(),
			    hour = eventDate.getHours(),
			    block = eventDate.getMinutes() / 15,
			    dayStr = year + "_" + month + "_" + day,
			    blockStr = e.DriverID + "_" + dayStr + "_" + hour + "_" + block,
			    eventDiv = createElement("div"),
			    eventCell, left, width;
			if (typeof days[dayStr] === "undefined") {
				return;
			}
			eventCell = days[dayStr][1].removeChild(days[dayStr][1].getElementById("cell_" + blockStr));
			left = eventCell.style.left;
			width = (e.End - e.Start) / 60000 + "px";
			eventDiv.setAttribute("class", "event");
			eventDiv.addEventListener("click", showEvent.bind(null, e));
			eventDiv.style.left = left;
			eventDiv.style.top = eventCell.style.top;
			eventDiv.style.width = width;
			eventDiv.setAttribute("id", "event_" + blockStr);
			rpc.getClient(e.ClientID, function(c) {
				var name = eventDiv.appendChild(createElement("div")).setInnerText(c.Name),
				    from = eventDiv.appendChild(createElement("div")).setInnerText(e.From),
				    to = eventDiv.appendChild(createElement("div")).setInnerText(e.To),
				    nameWidth = c.Name.getWidth("14px Serif"),
				    fromWidth = e.From.getWidth("14px Serif"),
				    toWidth = e.To.getWidth("14px Serif"),
				    maxWidth = nameWidth;
				name.style.width = nameWidth + "px";
				from.style.width = fromWidth + "px";
				to.style.width = toWidth + "px";
				if (fromWidth > maxWidth) {
					maxWidth = fromWidth;
				}
				if (toWidth > maxWidth) {
					maxWidth = toWidth;
				}
				
				if (maxWidth + 12 > parseInt(width)) { // 1px left border + 5px left padding + 5px right padding + 1px right border
					var newLeft = parseInt(left) - ((maxWidth - parseInt(width)) / 2);
					eventDiv.addEventListener("mouseover", function() {
						name.style.marginLeft = (maxWidth - nameWidth) / 2 + "px";
						from.style.marginLeft = (maxWidth - fromWidth) / 2 + "px";
						to.style.marginLeft = (maxWidth - toWidth) / 2 + "px";
						eventDiv.style.width = maxWidth + 12 + "px";
						eventDiv.style.left = newLeft + "px";
					});
					eventDiv.addEventListener("mouseout", function() {
						name.style.marginLeft = "0";
						from.style.marginLeft = "0";
						to.style.marginLeft = "0";
						eventDiv.style.left = left;
						eventDiv.style.width = width;
					});
				}
			});
			days[dayStr][1].appendChild(eventDiv);
		};
		this.addDriver = function(d) {
			if (typeof d === "undefined") {
				return;
			}
			drivers[d.ID] = d;
			drivers[d.ID].yPos = nextDriverPos;
			drivers[d.ID].events = [];
			var dDiv = createElement("div"),
			    t;
			drivers[d.ID].driverDiv = dDiv;
			dDiv.appendChild(createElement("div")).setInnerText(d.Name);
			dDiv.setAttribute("class", "driverName simpleButton");
			dDiv.setAttribute("id", "driver_" + d.ID);
			dDiv.addEventListener("click", function() {
				showDriver(drivers[d.ID]);
			});
			dDiv.style.top = nextDriverPos + "px";
			nextDriverPos += 100;
			plusDriver.style.top = nextDriverPos + "px";
			driverEvents.appendChild(dDiv);
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
			d.events = drivers[d.ID].events;
			d.yPos = drivers[d.ID].yPos;
			d.driverDiv = drivers[d.ID].driverDiv;
			d.driverDiv.getElementsByTagName("div")[0].setInnerText(d.Name);
			drivers[d.ID] = d;
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
	showCompany = function(company) {
		stack.addFragment();
		layer.appendChild(createElement("h1")).setInnerText(company.Name);
		var editDelete = layer.appendChild(createElement("div")),
		    edit = editDelete.appendChild(createElement("div")).setInnerText("Edit"),
		    deleter = editDelete.appendChild(createElement("div")).setInnerText("Delete"),
		    eventsClients = layer.appendChild(createElement("div")),
		    clientsButton = eventsClients.appendChild(createElement("div")).setInnerText("Clients"),
		    eventsButton = eventsClients.appendChild(createElement("div")).setInnerText("Events"),
		    eventsClientsDiv = layer.appendChild(createElement("div")),
		    eventsStartDate = new Date(),
		    eventsEndDate = new Date();
		eventsClients.setAttribute("class", "eventsClients");
		eventsClientsDiv.setAttribute("class", "eventsClientsDiv");
		clientsButton.addEventListener("click", function() {
			if (clientsButton.getAttribute("class") === "selected") {
				return;
			}
			rpc.clientsForCompany(company.ID, function(clients) {
				while (eventsClientsDiv.hasChildNodes()) {
					eventsClientsDiv.removeChild(eventsClientsDiv.lastChild);
				}
				eventsButton.removeAttribute("class");
				clientsButton.setAttribute("class", "selected");
				var clientsTable = createElement("table"),
				    headerRow = clientsTable.appendChild(createElement("tr")),
				    i = 0;
				headerRow.appendChild(createElement("th")).setInnerText("Name");
				headerRow.appendChild(createElement("th")).setInnerText("Phone Number");
				headerRow.appendChild(createElement("th")).setInnerText("Reference");
				for (; i < clients.length; i++) {
					var row = clientsTable.appendChild(createElement("tr")),
					    name = row.appendChild(createElement("td")).setInnerText(clients[i].Name);
					row.appendChild(createElement("td")).setInnerText(clients[i].PhoneNumber);
					row.appendChild(createElement("td")).setInnerText(clients[i].Reference);
				}
				eventsClientsDiv.appendChild(clientsTable);
			});
		});
		eventsButton.addEventListener("click", function() {
			if (eventsButton.getAttribute("class") === "selected") {
				return;
			}
			while (eventsClientsDiv.hasChildNodes()) {
				eventsClientsDiv.removeChild(eventsClientsDiv.lastChild);
			}
			clientsButton.removeAttribute("class");
			eventsButton.setAttribute("class", "selected");
			var oLayer = layer;
			layer = eventsClientsDiv;
			stack.addFragment();
			var dateCheck = regexpCheck(/^[0-9]{1,4}\/(0?[1-9]|1[0-2])\/(0?[1-9]|1[0-9]|2[0-9]|3[01])$/, "Please enter a valid date (YYYY/MM/DD)"),
			    startDate = addFormElement("Start Date", "text", "startDate", eventsStartDate.toDateString(), dateCheck),
			    endDate = addFormElement("End Date", "text", "endDate", eventsEndDate.toDateString(), dateCheck),
			    getEvents = addFormSubmit("Show Events", function() {
				while (eventTable.hasChildNodes()) {
					if (eventTable.lastChild === tableTitles) {
						break;
					}
					eventTable.removeChild(eventTable.lastChild);
				}
				var startParts = startDate[0].value.split("/"),
				    endParts = endDate[0].value.split("/");
				    eventsStartDate = new Date(startParts[0], startParts[1]-1, startParts[2]);
				    eventsEndDate = new Date(endParts[0], endParts[1]-1, endParts[2]);
				if (eventsStartDate.getTime() > eventsEndDate.getTime()) {
					endDate[1].setInnerText("Cannot be before start date");
					eventTable.appendChild(createElement("tr")).appendChild(createElement("td")).setInnerText("No Events").setAttribute("colspan", "5");
					return;
				}
				rpc.getEventsWithCompany(company.ID, eventsStartDate.getTime(), eventsEndDate.getTime() + (24 * 3600 * 1000), function(events) {
					var row,
					    i = 0;
					if (events.length === 0) {
						eventTable.appendChild(createElement("tr")).appendChild(createElement("td")).setInnerText("No Events").setAttribute("colspan", "5");
						return;
					}
					for (; i < events.length; i++) {
						row = createElement("tr");
						var clientCell = row.appendChild(createElement("td")),
						    driverCell = row.appendChild(createElement("td"));
						row.appendChild(createElement("td")).setInnerText(events[i].From);
						row.appendChild(createElement("td")).setInnerText(events[i].To);
						row.appendChild(createElement("td")).setInnerText(new Date(events[i].Start).toLocaleString());
						row.appendChild(createElement("td")).setInnerText(new Date(events[i].End).toLocaleString());
						rpc.getClient(events[i].ClientID, function(clientCell, client) {
							clientCell.setInnerText(client.Name);
						}.bind(null, clientCell));
						rpc.getDriver(events[i].DriverID, function(driverCell, driver) {
							driverCell.setInnerText(driver.Name);
						}.bind(null, driverCell));
						eventTable.appendChild(row);
					}
				});
			    }),
			    eventFormTable = layer.appendChild(createElement("table")),
			    eventTable = eventFormTable.appendChild(createElement("table")),
			    tableTitles = eventTable.appendChild(createElement("tr"));
			tableTitles.appendChild(createElement("th")).setInnerText("Client");
			tableTitles.appendChild(createElement("th")).setInnerText("Driver");
			tableTitles.appendChild(createElement("th")).setInnerText("From");
			tableTitles.appendChild(createElement("th")).setInnerText("To");
			tableTitles.appendChild(createElement("th")).setInnerText("Start");
			tableTitles.appendChild(createElement("th")).setInnerText("End");
			getEvents.dispatchEvent(new MouseEvent("click", {"view": window, "bubble": false, "cancelable": true}));
			stack.setFragment();
			layer = oLayer;
		});

		editDelete.setAttribute("class", "editDelete");
		edit.setAttribute("class", "simpleButton");
		edit.addEventListener("click", function() {
			stack.addLayer("editCompany", function(c) {
				if (typeof c !== "undefined") {
					stack.removeLayer(c);
					showCompany(c.ID);
				}
			});
			setCompany(company);
		});
		deleter.setAttribute("class", "simpleButton");
		deleter.addEventListener("click", function() {
			if(confirm("Are you sure you want to remove this company?")) {
				rpc.removeCompany(company.ID);
				stack.removeLayer(company.ID);
			}
		});
		stack.setFragment();
		clientsButton.dispatchEvent(new MouseEvent("click", {"view": window, "bubble": false, "cancelable": true}));
	},
	companyList = function(addList) {
		rpc.companies(function(companies) {
			stack.addFragment();
			layer.appendChild(createElement("h1")).setInnerText("Companies")
			var table = createElement("table"),
			    headerRow = table.appendChild(createElement("tr")),
			    addCompanyToTable = function(company) {
				if (typeof company === "undefined") {
					return;
				}
				var row = createElement("tr"),
				    nameCell = row.appendChild(createElement("td")).appendChild(createElement("div"));
				nameCell.setInnerText(company.Name);
				nameCell.setAttribute("class", "simpleButton");
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
				if (addList === true) {
					addLister(nameCell, stack.removeLayer.bind(null, company));
				}
				row.appendChild(createElement("td")).setInnerText(company.Address);
				table.appendChild(row);
			    };
			addAdder(null, function() {
				stack.addLayer("addCompany", addCompanyToTable);
				addCompany();
			});
			headerRow.appendChild(createElement("th")).setInnerText("Company Name");
			headerRow.appendChild(createElement("th")).setInnerText("Address");
			companies.map(addCompanyToTable);
			layer.appendChild(table);
			stack.setFragment();
		});
	},
	showClient = function(client) {
		stack.addLayer("showClient");
		stack.addFragment();
		layer.appendChild(createElement("h1")).setInnerText(client.Name);
		var editDelete = layer.appendChild(createElement("div")),
		    edit = editDelete.appendChild(createElement("div")).setInnerText("Edit"),
		    deleter = editDelete.appendChild(createElement("div")).setInnerText("Delete"),
		    dateCheck = regexpCheck(/^[0-9]{1,4}\/(0?[1-9]|1[0-2])\/(0?[1-9]|1[0-9]|2[0-9]|3[01])$/, "Please enter a valid date (YYYY/MM/DD)"),
		    startDate = addFormElement("Start Date", "text", "startDate", (new Date()).toDateString(), dateCheck),
		    endDate = addFormElement("End Date", "text", "endDate", (new Date()).toDateString(), dateCheck),
		    getEvents = addFormSubmit("Show Events", function() {
			while (eventTable.hasChildNodes()) {
				if (eventTable.lastChild === tableTitles) {
					break;
				}
				eventTable.removeChild(eventTable.lastChild);
			}
			var startParts = startDate[0].value.split("/"),
			    endParts = endDate[0].value.split("/"),
			    start = new Date(startParts[0], startParts[1]-1, startParts[2]),
			    end = new Date(endParts[0], endParts[1]-1, endParts[2]);
			rpc.getEventsWithClient(client.ID, start.getTime(), end.getTime() + (24 * 3600 * 1000), function(events) {
				var row,
				    i = 0;
				if (events.length === 0) {
					eventTable.appendChild(createElement("tr")).appendChild(createElement("td")).setInnerText("No Events").setAttribute("colspan", "5");
					return;
				}
				for (; i < events.length; i++) {
					row = createElement("tr");
					var driverCell = row.appendChild(createElement("td"));
					row.appendChild(createElement("td")).setInnerText(events[i].From);
					row.appendChild(createElement("td")).setInnerText(events[i].To);
					row.appendChild(createElement("td")).setInnerText(new Date(events[i].Start).toLocaleString());
					row.appendChild(createElement("td")).setInnerText(new Date(events[i].End).toLocaleString());
					rpc.getDriver(events[i].DriverID, function(driverCell, driver) {
						driverCell.setInnerText(driver.Name);
					}.bind(null, driverCell));
					eventTable.appendChild(row);
				}
			});
		    }),
		    eventTable = layer.appendChild(createElement("table")),
		    tableTitles = eventTable.appendChild(createElement("tr"));

		editDelete.setAttribute("class", "editDelete");
		edit.setAttribute("class", "simpleButton");
		edit.addEventListener("click", function() {
			stack.addLayer("editClient", function(c) {
				if (typeof c !== "undefined") {
					stack.removeLayer(c);
					showClient(c.ID);
				}
			});
			setClient(client);
		});
		deleter.setAttribute("class", "simpleButton");
		deleter.addEventListener("click", function() {
			if(confirm("Are you sure you want to remove this client?")) {
				rpc.removeClient(client.ID);
				stack.removeLayer(client.ID);
			}
		});
		tableTitles.appendChild(createElement("th")).setInnerText("Driver");
		tableTitles.appendChild(createElement("th")).setInnerText("From");
		tableTitles.appendChild(createElement("th")).setInnerText("To");
		tableTitles.appendChild(createElement("th")).setInnerText("Start");
		tableTitles.appendChild(createElement("th")).setInnerText("End");
		getEvents.dispatchEvent(new MouseEvent("click", {"view": window, "bubble": false, "cancelable": true}));
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
				    setCompanyCell = function() {
					companyCell.setInnerText(companies[client.CompanyID].Name);
					//companyCell.setAttribute("class", "simpleButton");
					//companyCell.addEventListener("click", showCompany.bind(null, companies[client.CompanyID]));
					client.CompanyName = companies[client.CompanyID].Name;
					nameCell.addEventListener("click", showClient.bind(null, client));
				    };
				nameCell.setInnerText(client.Name);
				nameCell.setAttribute("class", "simpleButton");
				if (addList === true) {
					addLister(nameCell, stack.removeLayer.bind(null, client));
				}
				if (typeof companies[client.CompanyID] !== "undefined") {
					setCompanyCell();
				} else {
					rpc.getCompany(client.CompanyID, function(company) {
						if (typeof company === "undefined") {
							companyCell.setInnerText("Error!");
							return;
						}
						companies[company.ID] = company;
						setCompanyCell();
					});
				}
				row.appendChild(createElement("td")).setInnerText(client.PhoneNumber);
				row.appendChild(createElement("td")).setInnerText(client.Reference);
				table.appendChild(row);
			    };
			addAdder(null, function() {
				stack.addLayer("addClient", addClientToTable);
				addClient();
			});
			headerRow.appendChild(createElement("th")).setInnerText("Client Name");
			headerRow.appendChild(createElement("th")).setInnerText("Company Name");
			headerRow.appendChild(createElement("th")).setInnerText("Phone Number");
			headerRow.appendChild(createElement("th")).setInnerText("Reference");
			clients.map(addClientToTable);
			layer.appendChild(table);
			stack.setFragment();
		});
	},
	messageList = function() {
		stack.addLayer("messages");
		layer.appendChild(createElement("h1")).setInnerText("Messages");
	},
	addTitle = function(id, add, edit) {
		layer.appendChild(createElement("h1")).setInnerText((id == 0) ? add : edit);
	},
	addFormElement = function(name, type, id, contents, onBlur) {
		var label = createElement("label").setInnerText(name),
		    input;
		if (type === "textarea") {
			input = createElement("textarea");
			input.setAttribute("spellcheck", "false");
		} else {
			input = createElement("input");
			input.setAttribute("type", type);
		}
		input.setAttribute("value", contents);
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
	showDriver = function(driver) {
		stack.addLayer("showDriver");
		stack.addFragment();
		layer.appendChild(createElement("h1")).setInnerText(driver.Name);
		var editDelete = layer.appendChild(createElement("div")),
		    edit = editDelete.appendChild(createElement("div")).setInnerText("Edit"),
		    deleter = editDelete.appendChild(createElement("div")).setInnerText("Delete"),
		    dateCheck = regexpCheck(/^[0-9]{1,4}\/(0?[1-9]|1[0-2])\/(0?[1-9]|1[0-9]|2[0-9]|3[01])$/, "Please enter a valid date (YYYY/MM/DD)"),
		    startDate = addFormElement("Start Date", "text", "startDate", (new Date()).toDateString(), dateCheck),
		    endDate = addFormElement("End Date", "text", "endDate", (new Date()).toDateString(), dateCheck),
		    getEvents = addFormSubmit("Show Events", function() {
			while (eventTable.hasChildNodes()) {
				if (eventTable.lastChild === tableTitles) {
					break;
				}
				eventTable.removeChild(eventTable.lastChild);
			}
			var startParts = startDate[0].value.split("/"),
			    endParts = endDate[0].value.split("/"),
			    start = new Date(startParts[0], startParts[1]-1, startParts[2]),
			    end = new Date(endParts[0], endParts[1]-1, endParts[2]);
			rpc.getEventsWithDriver(driver.ID, start.getTime(), end.getTime() + (24 * 3600 * 1000), function(events) {
				var row,
				    i = 0;
				if (events.length === 0) {
					eventTable.appendChild(createElement("tr")).appendChild(createElement("td")).setInnerText("No Events").setAttribute("colspan", "6");
					return;
				}
				for (; i < events.length; i++) {
					row = createElement("tr");
					var clientCell = row.appendChild(createElement("td")),
					    companyCell = createElement("td");
					row.appendChild(createElement("td")).setInnerText(events[i].From);
					row.appendChild(createElement("td")).setInnerText(events[i].To);
					row.appendChild(createElement("td")).setInnerText(new Date(events[i].Start).toLocaleString());
					row.appendChild(createElement("td")).setInnerText(new Date(events[i].End).toLocaleString());
					row.appendChild(companyCell);
					rpc.getClient(events[i].ClientID, function(clientCell, companyCell, client) {
						clientCell.setInnerText(client.Name);
						rpc.getCompany(client.CompanyID, function(company) {
							companyCell.setInnerText(company.Name);
						});
					}.bind(null, clientCell, companyCell));
					eventTable.appendChild(row);
				}
			});
		    }),
		    eventTable = layer.appendChild(createElement("table")),
		    tableTitles = eventTable.appendChild(createElement("tr"));

		editDelete.setAttribute("class", "editDelete");
		edit.setAttribute("class", "simpleButton");
		edit.addEventListener("click", function() {
			stack.addLayer("editDriver", function(d) {
				if (typeof d !== "undefined") {
					stack.removeLayer();
					events.updateDriver(d);
					showDriver(d);
				}
			});
			setDriver(driver);
		});
		deleter.setAttribute("class", "simpleButton");
		deleter.addEventListener("click", function() {
			if(confirm("Are you sure you want to remove this driver?")) {
				rpc.removeDriver(driver.ID);
				stack.removeLayer();
				events.removeDriver(driver);
			}
		});


		tableTitles.appendChild(createElement("th")).setInnerText("Client");
		tableTitles.appendChild(createElement("th")).setInnerText("From");
		tableTitles.appendChild(createElement("th")).setInnerText("To");
		tableTitles.appendChild(createElement("th")).setInnerText("Start");
		tableTitles.appendChild(createElement("th")).setInnerText("End");
		tableTitles.appendChild(createElement("th")).setInnerText("Company");
		getEvents.dispatchEvent(new MouseEvent("click", {"view": window, "bubble": false, "cancelable": true}));
		stack.setFragment();
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
					driverName[1].setInnerText(resp.NameError);
					regNumber[1].setInnerText(resp.RegError);
					phoneNumber[1].setInnerText(resp.PhoneError);
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
		addFormSubmit("Add Client", function() {
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
		});
	},
	showEvent = function(e) {
		stack.addLayer("showEvent");
		alert(e.Start);
	},
	setEvent = (function() {
		var fromAddressRPC = rpc.autocompleteAddress.bind(rpc, 0),
		    toAddressRPC = rpc.autocompleteAddress.bind(rpc, 1);
		return function(event) {
			stack.addFragment();
			addTitle(event.ID, "Add Event", "Edit Event");
			addFormElement("Driver", "text", "", event.DriverName);
			addFormElement("Start", "text", "", dateTimeFormat(event.Start));
			var driverTime = addFormElement("End", "text", "", dateTimeFormat(event.End)),
			    from = addFormElement("From", "textarea", "from", event.From, regexpCheck(/.+/, "From Address Required")),
			    to = addFormElement("To", "textarea", "to", event.To, regexpCheck(/.+/, "To Address Required")),
			    clientID = addFormElement("", "hidden", "", event.ClientID),
			    clientName = addFormElement("Client Name", "text", "client_name", event.ClientName, regexpCheck(/.+/, "Client Name Required"));
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
			autocomplete(fromAddressRPC, from[0]);
			autocomplete(toAddressRPC, to[0]);
			autocomplete(rpc.autocompleteClientName, clientName[0], clientID);
			addFormSubmit("Add Event", function() {
				var parts = [this, clientName[0], to[0], from[0]];
				parts.map(disableElement);
				event.ClientID = parseInt(clientID.value);
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
				errorDiv.setInnerText("");
			} else {
				errorDiv.setInnerText(error);
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
			var bounds = nameDiv.getBoundingClientRect();
			autocompleteDiv.style.left = Math.round(bounds.left + (window.pageXOffset || document.documentElement.scrollLeft || document.body.scrollLeft) - (document.documentElement.clientLeft || document.body.clientLeft || 0)) + "px";
			autocompleteDiv.style.top = Math.round(bounds.bottom + (window.pageYOffset || document.documentElement.scrollTop || document.body.scrollTop) - (document.documentElement.clientTop || document.body.clientTop || 0)) + "px";
			autocompleteDiv.style.width = (bounds.right - bounds.left) + "px";
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
				return year + "/" + month + "/" + date;
			},
			toTimeString: function() {
				var hour = this.getHours(),
				    minutes = this.getMinutes(),
				    seconds = this.getSeconds();
				if (hour < 10) {
					hour = "0" + hour;
				}
				if (minutes < 10) {
					minutes = "0" + minutes;
				}
				if (seconds < 10) {
					seconds = "0" + seconds;
				}
				return hour + ":" + minutes + ":" + seconds;
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
	Element.prototype.getElementById = function(id) {
		return this.querySelector("#" + id);
	};
	Element.prototype.setInnerText = function(text) {
		while (this.hasChildNodes()) {
			this.removeChild(this.lastChild);
		}
		this.appendChild(document.createTextNode(text));
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
}.bind(null, Date));

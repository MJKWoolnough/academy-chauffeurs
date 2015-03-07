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
		this.setDriver = function(driver, callback) {
			request("SetDriver", driver, callback);
		}
		this.drivers = function(callback) {
			request("Drivers", 0, callback);
		}
		this.events = function(driverID, From, To, callback) {
			request("Events", {"DriverID": driverID, "From": from, "To": to}, callback);
		}
		ws.onopen = onload;
	})(function() {
		eventList();
	}),
	drivers = [],
	companies = [],
	clients = [],
	events = [],
	layer,
	stack = new (function(){
		var stack = [];
		this.addLayer = function(layerID, callback) {
			stack.push(callback);
			var outerLayer = document.createElement("div");
			outerLayer.className = "layer";
			layer = document.createElement("div");
			layer.setAttribute("id", layerID);
			outerLayer.appendChild(layer);
			document.body.appendChild(outerLayer);
		};
		this.removeLayer = function() {
			if (stack.length === 0) {
				return;
			}
			var callback = stack.pop();
			if (typeof callback === "function") {
				callback.apply(arguments);
			}
			document.body.removeChild(document.body.lastChild);
			layer = document.body.lastChild.firstChild;
		};
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
				addDriver();
			} else {
			}
		});
	},
	addFormElement = function(name, type, id, contents, onBlur, onChange) {
		var label = document.createElement("label"),
		error = document.createElement("div"),
		input;
		if (type === "textarea") {
			input = document.createElement("textarea");
			input.innerHTML = contents;
		} else {
			input = document.createElement("input");
			input.setAttribute("type", type);
			input.setAttribute("value", contents);
		}
		label.innerHTML = name;
		label.setAttribute("for", id);
		input.setAttribute("id", id);
		if (typeof onChange === "function") {
			input.addEventListener("change", onChange.bind(input));
		}
		if (typeof onBlur === "function") {
			input.addEventListener("blur", onBlur.bind(input));
		}
		error.setAttribute("class", "error");
		error.setAttribute("id", "error_"+id);
		layer.appendChild(label);
		layer.appendChild(input);
		layer.appendChild(error);
		layer.appendChild(document.createElement("br"));
		return input;
	},
	addFormSubmit = function(value, onClick) {
		var button = document.createElement("input");
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
	addDriver = function() {
		layer.appendChild(document.createElement("h1")).innerHTML = "Add Driver";
		var driverName = addFormElement("Driver Name", "text", "driver_name", "", regexpCheck(/.+/, "Please enter a valid name")),
		regNumber = addFormElement("Registration Number", "text", "driver_reg", "", regexpCheck(/[a-zA-Z0-9 ]+/, "Please enter a valid Vehicle Registration Number")),
		phoneNumber = addFormElement("Phone Number", "text", "driver_phone", "", regexpCheck(/^(0|\+?44)[0-9 ]{10}$/, "Please enter a valid mobile telephone number"));
		addFormSubmit("Add Driver", function() {
			var parts = [this, driverName, regNumber, phoneNumber];
			parts.map(disableElement);
			rpc.setDriver({
				"Name": driverName.value,
				"RegistrationNumber": regNumber.value,
				"PhoneNumber": phoneNumber.value,
			}, function(resp) {
				if (resp.Errors) {
					document.getElementById("error_driver_name").innerHTML = resp.NameError;
					document.getElementById("error_driver_reg").innerHTML = resp.RegError;
					document.getElementById("error_driver_phone").innerHTML = resp.PhoneError;
					parts.map(enableElement);
				} else {
					// add driver to a list?
					stack.removeLayer();
				}
			});
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
	};
};

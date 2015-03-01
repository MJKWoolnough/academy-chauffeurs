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
				alert("!undefined!");
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
		this.append = function(callback) {
			stack.push(callback);
		};
		this.doStack = function() {
			if (stack.length === 0) {
				return;
			}
			var callback = stack.pop();
			callback.apply(arguments);
		};
		this.addLayer = function() {
			layer = document.createElement("div");
			layer.className = "layer";
			document.body.appendChild(layer);
		};
		this.removeLayer = function() {
			document.body.removeChild(document.body.lastChild);
			layer = document.body.lastChild;
		};
		this.addLayer();
	})(),
	eventList = function(date) {
		if (arguments.length == 0) {
			date = Date.now()
		}
		rpc.drivers(function(d) {
			drivers = d;
			if (drivers.length === 0) {
				//addDriver();
			} else {
				
			}
		});
	}
};

window.onload = function() {
	var rpc = new (function(onload){
		var ws = new WebSocket("ws://127.0.0.1:8080/rpc", "rpc"),
		requests = [],
		nextID = 0;
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
		},
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
		eventList(Date.now());
	}),
	drivers = [],
	companies = [],
	clients = [],
	events = [],
	stack = new (function(){
		var stack = [],
		layer = document.createElement("div"),
		layers = 0;
		layer.id = "fadeLayer";
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
		this.addLayer() {
			layers += 2
			layer.style.zindex = layers - 1;
			document.body.addChild(layer);
			return layer;
		};
		this.removeLayer() {
			if (layers === 0) {
				return 0;
			}
			layers -= 2;
			if (layers === 0) {
				document.body.removeChild(layer);
			}
			layer.style.zindex = layer - 1;
			return layer;
		};
	})(),
	eventList = function(date) {
		rpc.drivers(function(d) {
			drivers = d;
			if (drivers.length === 0) {
				addDriver();
			} else {
				
			}
		});
	}
};

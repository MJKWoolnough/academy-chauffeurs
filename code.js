(function() {
	var rpc = new (function(){
		var ws = new WebSocket("ws://127.0.0.1:8080/rpc", "rpc"),
		requests = [],
		nextID = 0;
		ws.onmessage = function (event) {
			var data = JSON.parse(event.data),
			req = requests[data.id];
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
		this.test = function(num, callback) {
			request("Test", num, callback);
		}
	})();
	window.setTimeout(function() {
		alert("Requesting");
		rpc.test(1, function(data) {
			alert(data);
		});
	}, 1000);
}())

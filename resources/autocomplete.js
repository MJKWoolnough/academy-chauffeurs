(function () {
	"use strict";
	var ns = document.getElementsByTagName("HTML")[0].namespaceURI,
	removeChildren = function (node) {
		while (node.hasChildNodes()) {
			node.removeChild(node.firstChild);
		}
	},
	setCompleted = function (data, input, list, errDiv) {
		return function() {
			input.value = data;
			errDiv.textContent = "";
			removeChildren(list);
		}
	},
	autocompleteIt = function (url, input, list, errDiv) {
		var latest = 0;
		return function() {
			errDiv.textContent = "";
			if (input.value === "") {
				removeChildren(list);
				return;
			}
			var xh = new XMLHttpRequest(),
			text = input.value,
			data = "partial="+encodeURIComponent(text).replace(/%20/g, '+');
			xh.open("POST", url, true);
			xh.setRequestHeader("Content-type", "application/x-www-form-urlencoded");
			xh.setRequestHeader("Content-length", data.length);
			xh.setRequestHeader("Connection", "close");
			xh.onreadystatechange = function() {
				if (xh.readyState === 4 && xh.status === 200) {
					var jsonData = JSON.parse(xh.responseText), i;
					if (jsonData.Time > latest) {
						latest = jsonData.Time;
						removeChildren(list);
						if (jsonData.Data.length === 0) {
							errDiv.textContent = "No matching company";
						} else if (jsonData.Data.length === 1 && jsonData.Data[0] === text) {
							return;
						}
						for (i = 0; i < jsonData.Data.length; i++) {
							var li = document.createElementNS(ns, "li"),
							jData = jsonData.Data[i],
							startPos = jData.toUpperCase().indexOf(text.toUpperCase()),
							matchHighlight = document.createElementNS(ns, "b");
							li.appendChild(document.createTextNode(jData.slice(0, startPos)));
							matchHighlight.appendChild(document.createTextNode(jData.slice(startPos, startPos+text.length)));
							li.appendChild(matchHighlight);
							li.appendChild(document.createTextNode(jData.slice(startPos+text.length)));
							li.addEventListener("click", setCompleted(jsonData.Data[i], input, list, errDiv));
							list.appendChild(li);
						}
					}
				}
			};
			xh.send(data)
		}
	},
	allInputs = Array.prototype.slice.apply(document.getElementsByTagName("input")),
	inputLength = allInputs.length,
	i;
	for (i = 0; i < inputLength; i++) {
		if (allInputs[i].className === "autocomplete") {
			var errDiv = allInputs[i].nextSibling, 
			autocompleteDiv = document.createElementNS(ns, "div"),
			list = document.createElementNS(ns, "ul"),
			url = allInputs[i].getAttribute("autocomplete-url");
			while (errDiv.nodeType !== 1) {
				errDiv = errDiv.nextSibling;
			}
			allInputs[i].removeAttribute("autocomplete-url");
			allInputs[i].setAttribute("autocomplete", "off");
			allInputs[i].parentNode.replaceChild(autocompleteDiv, allInputs[i]);
			autocompleteDiv.appendChild(allInputs[i]);
			autocompleteDiv.appendChild(list);
			allInputs[i].addEventListener("keyup", autocompleteIt(url, allInputs[i], list, errDiv));
		}
	}
}());
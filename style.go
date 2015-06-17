package main

var styleCSS = []byte(`@media screen {
	html, body {
		color : #000;
		background-color : #fff;
		width : 100%;
		overflow-x : hidden;
	}

	h1 {
		text-align : center;
	}

	.layer {
		height : 100%;
		left : 0;
		position : absolute;
		top : 0;
		width : 100%;
		overflow-x : hidden;
	}

	.layer:first-child {
		background-color : transparent !important;
	}

	.layer:first-child > div {
		width : 100%;
	}

	.layer:last-child {
		background-color : rgba(128, 128, 128, 0.5);
	}

	.layer > div {
		background-color: #fff;
		margin-left : auto;
		margin-right : auto;
		width : 90%;
	}

	/* form */

	label + input, label + div, label + textarea {
		float : left;
		width : 200px;
	}

	input[readonly=readonly] {
		background-color : #eee;
		border-color : #fff;
		border-style : solid;
	}

	div.adder {
		float : left;
		font-weight : bold;
		color : #fff;
		text-shadow : #000 -1px -1px 0, #000 1px -1px 0, #000 1px 1px 0, #000 -1px 1px 0, #000 -1px 0 0, #000 1px 0 0, #000 0 -1px 0, #000 0 1px 0;
		transition : color 0.25s;
		transition-timing-function : ease;
	}

	div.adder:hover {
		color : #00f;
		cursor : pointer;
	}

	div.error {
		float : left;
		color : #f00;
		padding-left : 10px;
	}

	div.error + br {
		clear : left;
	}

	.driverName, #plusDriver {
		-moz-user-select : none;
		-webkit-user-select : none;
		left : 0;
		height : 100px;
		position : absolute;
		width : 100px;
		z-index : 4;
	}

	.driverName > div, #plusDriver > div {
		position : relative;
		top : 50%;
		text-align : center;
		transform : translateY(-50%);
	}

	.slider {
		transition : left 1s;
		transition-timing-function : ease-out;
	}

	.dates {
		-moz-user-select : none;
		-webkit-user-select : none;
		position : absolute;
		top : 20px;
	}

	.driverEvents {
		position : absolute;
		top : 120px;
		left : 0px;
		right : 0;
		bottom : 0;
		overflow-x : hidden;
	}

	.events {
		position : absolute;
	}

	.year {
		position : absolute;
		height : 20px;
		top : 0;
		background-color : #fff;
		z-index : 4;
	}

	.month {
		position : absolute;
		top : 20px;
		height : 20px;
		background-color : #fff;
		z-index : 4;
	}

	.day {
		position : absolute;
		height : 20px;
		width : 1440px;
		top : 40px;
		margin : 0;
		padding : 0;
		background-color : #fff;
		z-index : 4;
	}

	.year > div, .month > div, .day > div {
		display : inline-block;
		position : absolute;
		white-space : nowrap;
		margin : 0;
		padding : 0;
	}

	.hour {
		position : absolute;
		top : 60px;
		width : 60px;
		text-align : center;
		border-style : solid;
		border-color : #000;
		border-width : 0 0 0 1px;
		background-color : #fff !important;
		z-index : 4;
	}

	.minute {
		position : absolute;
		height : 20px;
		top : 80px;
		width : 15px;
		text-align : center;
		transform: rotate(-90deg);
		background-color : #fff;
		z-index : 5;
		transition :color 0.2s;
		transition-timing-function : ease;
	}

	.minute.select {
		color : #f00;
	}

	.moveLeft, .moveRight {
		-moz-user-select : none;
		-webkit-user-select : none;
		line-height : 100%;
		position : absolute;
		width : 20px;
		z-index : 6;
		height : 20px;
		text-align : center;
	}

	.moveLeft span, .moveRight span {
		vertical-align : middle;
	}

	.moveLeft {
		left : 0;
	}

	.moveRight {
		right : 0;
	}

	.dayEnclosure {
		position : absolute;
		top : 0;
		left : 0;
	}

	.eventCell {
		position : absolute;
		width : 15px;
		height : 100px;
		z-index : 1;
		transition : background-color 0.2s;
		transition-timing-function : ease;
	}

	.driverUnassignedEven, .driverUnassignedOdd {
		position : absolute;
	}

	.cellOdd, .driverUnassignedOdd .cellEven {
		background-color : #ddd;
	}

	.cellEven, .driverUnassignedOdd .cellOdd {
		background-color : #aaa;
	}

	.eventCell.eventHover {
		background-color : #f00;
	}

	.eventCell.eventSelected {
		background-color : #0f0;
	}

	.eventCell.eventsInBetween {
		background-color : #00f;
	}

	.eventMover {
		border : 1px solid #000;
		position : absolute;
		width : 6px;
		height : 6px;
		background-color : #fff;
		font-size : 5px;
		text-align : center;
		top : 0px;
		left : 0px;
		padding : 0;
		margin : 0;
		transform : translate(-1px, -1px);
		transition : left 1s, background-color 0.25s !important;
		transition-timing-function : ease;
	}

	.eventMover.selected {
		background-color : #000;
	}

	.canceller {
		float : right;
		margin-right : 2px;
		color : #fff;
		text-shadow : #000 -1px -1px 0, #000 1px -1px 0, #000 1px 1px 0, #000 -1px 1px 0, #000 -1px 0 0, #000 1px 0 0;
		font-family : serif;
		transition : color 0.25s;
		transition-timing-function : ease;
	}

	.canceller:hover {
		color : #f00;
		cursor : pointer;
	}

	#topBar > div {
		height : 20px;
		float : left;
		text-align : center;
		width : 30%;
	}

	#topBar > div:last-child {
		width : 10%;
	}

	.simpleButton {
		background-color : #eee;
		transition : background-color 0.25s, color 0.25s;
		transition-timing-function : ease;
	}

	.pulse {
		animation: pulser 2s infinite;
	}

	@keyframes pulser {
		0% { background-color : #eee; }
		50% { background-color : #aaa; }
		100% { background-color : #eee; }
	}

	.nearPulse {
		animation: nearPulser 2s infinite;
	}

	@keyframes nearPulser {
		0% { background-color : #eee; }
		50% { background-color : #f00; }
		100% { background-color : #eee; }
	}

	.simpleButton:hover {
		cursor : pointer;
		background-color : #000 !important;
		color : #eee;
	}

	.autocompleter {
		position : absolute;
		background-color : #fff;
		cursor : pointer;
		list-style : none;
		margin : 1px 0 0 0;
		padding : 0;
	}

	.autocompleter li {
		border : 1px solid #000 !important;
		border-collapse: collapse;
		text-align : center;
		margin-top : -1px;
		overflow : hidden;
	}

	.autocompleter li:nth-child(even), .autocompleter li:nth-child(even) .disambiguator {
		background-color : #ddd;
		border-left-color : #ddd;
	}

	.autocompleter li:nth-child(odd), .autocompleter li:nth-child(odd) .disambiguator {
		background-color : #eee;
		border-left-color : #eee;
	}

	.event {
		position : absolute;
		height : 100px;
		z-index : 2;
		background-color : #eee;
		box-sizing : border-box;
		border : 1px solid #000;
		padding : 5px;
		overflow : hidden;
		font : 14px Serif;
		-moz-user-select : none;
		-webkit-user-select : none;
		cursor : pointer;
	}

	.event .eventCell {
		display : none;
	}

	.event.expandable {
		transition : z-index 1s step-end, left 1s, width 1s;
	}

	.event.expandable:hover {
		z-index : 3;
		transition : z-index 0s step-start, left 1s, width 1s;
	}

	.event > div {
		margin-left : 0;
		transition : margin-left 1s;
		white-space : nowrap;
	}

	.event > div.time {
		width : 100%;
		text-align : center;
		transition : color 1s;
		color : transparent;
	}

	.event:hover > div.time {
		visibility : visible;
		color : #000;
	}

	.disambiguator {
		display : inline;
		transform : translate(-1px, -1px) scaleX(0);
		position : absolute;
		overflow : hidden;
		background-color : transparent;
		transition : transform 0.5s;
		transition-timing-function : ease-in-out;
		transform-origin : 0;
		-webkit-transform-origin : 0;
		white-space : nowrap;
		border-width : 1px 1px 1px 0;
		border-style : solid;
		border-color : #000;
	}

	.autocompleter li:hover .disambiguator {
		transform : translate(-1px, -1px) scaleX(1);
	}

	.editDelete {
		width : 50%;
		margin : 0 auto 20px auto;
		overflow : hidden;
	}

	.editDelete > div {
		width : 50%;
		float : left;
		text-align : center;
	}

	.tabs {
		line-height : 24px;
		position : relative;
		width : 100%;
		overflow : hidden;
		margin : 0;
		padding : 0 0 0 20px;
	}

	.tabs:after {
		position : absolute;
		content : "";
		width : 100%;
		bottom : 0;
		left : 0;
		border-bottom: 1px solid #000;
		z-index : 1;
		overflow : hidden;
		text-align : center;
		transform : translateX(-20px);
	}

	.tabs > div {
		border : 1px solid #000;
		display : inline-block;
		position : relative;
		z-index : 1;
		margin : 0 -5px;
		padding : 0 20px;
		border-top-right-radius: 6px;
		border-top-left-radius: 6px;
		background : linear-gradient(to bottom, #ececec 50%, #d1d1d1 100%);
		box-shadow : 0 3px 3px rgba(0, 0, 0, 0.4), inset 0 1px 0 #fff;
		text-shadow : 0 1px #fff;
		-moz-user-select : none;
		-webkit-user-select : none;
	}

	.tabs > div:hover {
		background : linear-gradient(to bottom, #faa 1%, #ffecec 50%, #d1d1d1 100%);
		cursor : pointer;
	}

	.tabs > div:before, .tabs > div:after {
		position : absolute;
		bottom : -1px;
		width : 6px;
		height : 6px;
		content : " ";
		border : 1px solid #000;
	}

	.tabs > div:before {
		left : -7px;
		border-bottom-right-radius : 6px;
		border-width : 0 1px 1px 0;
		box-shadow: 2px 2px 0 #d1d1d1;
	}

	.tabs > div:after {
		right : -7px;
		border-bottom-left-radius : 6px;
		border-width : 0 0 1px 1px;
		box-shadow: -2px 2px 0 #d1d1d1;
	}

	.tabs > div.selected {
		border-bottom-color : #fff;
		z-index : 2;
		background : #fff;
	}

	.tabs > div.selected:hover {
		background : #fff;
		cursor : default;
	}

	.tabs > div.selected:before {
		box-shadow: 2px 2px 0 #fff;
	}

	.tabs > div.selected:after {
		box-shadow: -2px 2px 0 #fff;
	}

	.tabs + .content {
		margin-top : 10px;
		overflow : hidden;
		padding : 10px;
	}

	.printOnly {
		display : none;
	}

	.invoiceTop td[contenteditable=true], .invoice td[contenteditable=true], .invoiceBottom td[contenteditable=true] {
		background-color : #f7f8f8;
	}
}

@media print {

	* {
		visibility : hidden;
		height : 0;
		margin : 0;
		padding : 0;
		position : absolute;
		width : 100%;
	}

	.layer {
		display : none;
	}

	.layer:last-child {
		display : initial;
	}

	h1, h2, h3, h4 {
		text-align : center;
	}

	div.adder, .canceller {
		display : none;
	}

	.toPrint, .toPrint * {
		position : initial;
		height : initial !important;
		margin : initial;
		padding : initial;
		width : initial;
		visibility : visible !important;
	}

	.toPrint table {
		width : 100%;
	}

	.noPrint {
		display : none;
	}

	textarea, input {
		background-color : #fff;
		border : 0;
	}

	input[type="text"] {
		width : 100%;
	}
}

table {
	width : 100%;
	border-collapse : collapse;
}

table td, table tr {
	padding : 0 5px;
}

table tr:last-child td:first-child:last-child {
	text-align : center;
}

.tabs + .content label + textarea {
	width : 400px;
	height : 200px !important;
	margin-left : 10px;
}

label {
	display : inline-block;
	float : left;
	width : 250px;
	text-align : right;
}

label:after {
	content: ":"
}

.tabs + .content label + div {
	float : none;
	width : auto;
	margin-left : 260px;
}


.invoiceTop tr:first-child + tr td:first-child {
	width : 99%;
}

.invoiceTop td:not(:first-child), .invoiceTop td:nth-last-child(2) {
	white-space : nowrap;
}

.invoiceTop td:nth-last-child(2) {
	text-align : right;
	padding-right : 20px;
}

.invoice td:last-child, .invoice td:nth-last-child(2) {
	text-align : right;
}

.invoice td:nth-last-child(2) {
	padding-right : 10px;
}

.invoice td:nth-child(2) {
	text-align : center;
}

.invoice td:nth-child(4) {
	text-align : left;
}

.invoice {
	margin : 20px 0;
}

.invoiceBottom td:first-child {
	width : 99%;
}

.invoiceBottom td:first-child + td {
	padding-right : 20px;
	text-align : right;
	white-space : nowrap;
}

.invoiceBottom td:last-child {
	white-space : nowrap;
	text-align : right !important;
}

.invoiceBottom td:last-child:before {
	margin-left : -10px;
}

.invoiceBottom tr.line td:last-child {
	border-bottom : 1px solid #000;
}

.invoiceBottom tr.line + tr td {
	border-bottom : 10px solid transparent;
}

.invoiceBottom tr.doubleLine td:last-child {
	border-top : 10px solid transparent;
	border-bottom : 3px double #000;
}

tr.overline td {
	border-top : 2px solid #000;
	cell
}`)

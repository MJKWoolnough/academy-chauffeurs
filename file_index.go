package main

import (
	"time"

	"vimagination.zapto.org/httpdir"
)

func init() {
	date := time.Unix(1589836937, 0)
	httpdir.Default.Create("index.html", httpdir.FileString(`<!DOCTYPE html>
<html xmlns="http://www.w3.org/1999/xhtml">
	<head>
		<title>Calendar</title>
		<link rel="stylesheet" type="text/css" href="style.css" />
		<script type="text/javascript" src="code.js"></script>
	</head>
	<body></body>
</html>
`, date))
}

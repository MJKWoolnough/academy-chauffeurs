package main

import (
	"time"

	"github.com/MJKWoolnough/httpdir"
)

func init() {
	httpdir.Default.Create("index.html", httpdir.FileString(`<!DOCTYPE html>
<html xmlns="http://www.w3.org/1999/xhtml">
	<head>
		<title>Calendar</title>
		<link rel="stylesheet" type="text/css" href="style.css" />
		<script type="text/javascript" src="code.js"></script>
	</head>
	<body></body>
</html>
`, time.Unix(1471785913, 0)))
}

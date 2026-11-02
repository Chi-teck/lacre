// Command lacre turns plain text or an HTML snippet into a full HTML page.
//
//	echo 'Hello world!' | lacre build > hello-world.html
//
// The build command reads all of standard input and writes the page to
// standard output. Messages go to standard error.
//
// The page is self-contained: nothing lacre adds to it refers to anything
// outside the page. It is a single file that works offline, unless what you
// put in it refers to outside files.
//
// The input goes into the page as HTML, as it is, and is not cleaned: a script
// in the input runs when the page is opened, so use lacre only on input you
// trust. Plain text needs escaping, and a <pre> to keep its line breaks,
// before it goes in.
package main

import "github.com/Chi-teck/lacre/internal/app"

func main() { app.Main() }

package main

//import "fmt"
//import "github.com/sofiazaidenberg/newmath"

//func main() {
//	fmt.Printf("Hello, world.  Sqrt(2) = %v\n", newmath.Sqrt(2))
//}

import (
	"fmt"
	"html"
	"net/http"
	"time"
)

type Hello struct{}

type String string

type Struct struct {
	Greeting string
	Punct    string
	Who      string
}

func (s Struct) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var strNow string
	strNow = fmt.Sprintf("%v\n", time.Now())
	fmt.Fprintf(w, "<html><font color=\"red\">"+s.Greeting+
		"<font color=\"green\">"+s.Punct+"</font>"+
		"<font color=\"blue\">"+s.Who+"</font><br>"+
		"<font color=\"orange\">"+strNow+
		"</font></html>")
}

func (s String) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, string(s))
}

func (h Hello) ServeHTTP(
	w http.ResponseWriter,
	r *http.Request) {
	//    fmt.Fprint(w, "<html><font color=\"red\">Hello!</font></html>")
	fmt.Fprintf(w, "Hello, %q", html.EscapeString(r.Host))
}

func main() {
	//    var h Hello
	//    http.ListenAndServe("localhost:4000", h)
	str := String("I'm a frayed knot.")
	http.Handle("/string", str)
	//	http.ListenAndServe("localhost:4000", s)
	s := Struct{"Hello", ":", "Gophers!"}
	http.Handle("/struct", s)
	http.ListenAndServe("localhost:4000", nil)
}

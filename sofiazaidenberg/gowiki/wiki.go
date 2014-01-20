package main

import (
	"errors"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strings"
	"fmt"
)

var tmplDir = "tmpl/"
var dataDir = "data/"
var templates = template.Must(template.ParseFiles(tmplDir+"edit.html", tmplDir+"view.html", tmplDir+"index.html", tmplDir+"create.html"))
var validPath = regexp.MustCompile("^/(edit|save|view)/([a-zA-Z0-9]+)$")

type Site struct {
	Name  string
	Pages []string
}

type Page struct {
	Title string
	Body  []byte
}

func (p *Page) save() error {
	filename := p.Title + ".txt"
	return ioutil.WriteFile(dataDir+filename, p.Body, 0600)
}

func loadPage(title string) (*Page, error) {
	filename := title + ".txt"
	body, err := ioutil.ReadFile(dataDir + filename)
	if err != nil {
		return nil, err
	}
	return &Page{Title: title, Body: body}, nil
}

//func handler(w http.ResponseWriter, r *http.Request) {
//	fmt.Fprintf(w, "Hi there, I love %s!", r.URL.Path[1:])
//}

func getTitle(w http.ResponseWriter, r *http.Request) (string, error) {
	m := validPath.FindStringSubmatch(r.URL.Path)
	if m == nil {
		http.NotFound(w, r)
		return "", errors.New("Invalid Page Title")
	}
	return m[2], nil // The title is the second subexpression.
}

func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
	err := templates.ExecuteTemplate(w, tmpl+".html", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func viewHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(title)
	if err != nil {
		http.Redirect(w, r, "/edit/"+dataDir+title, http.StatusFound)
		return
	}
	renderTemplate(w, "view", p)
}

func editHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(title)
	if err != nil {
		p = &Page{Title: title}
	}
	renderTemplate(w, "edit", p)
}

func createHandler(w http.ResponseWriter, r *http.Request) {
	p := new(Page)
	renderTemplate(w, "create", p)
	title := r.FormValue("title")
	body := r.FormValue("body")
	p.Title = title
	p.Body = []byte(body)
	err := p.save()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/view/"+title, http.StatusFound)
}

func saveHandler(w http.ResponseWriter, r *http.Request, title string) {
	body := r.FormValue("body")
	p := &Page{Title: title, Body: []byte(body)}
	err := p.save()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/view/"+title, http.StatusFound)
}

func renderHomeTemplate(w http.ResponseWriter, s *Site) {
	err := templates.ExecuteTemplate(w, "index.html", s)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func homeHandler(w http.ResponseWriter, r *http.Request, name string) {
	pwd, err := os.Getwd()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	//[]os.FileInfo
	files, err := ioutil.ReadDir(pwd + "/" + dataDir)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	site := &Site{Name: name, Pages: make([]string, 0)}

	for _, f := range files {
		if !f.IsDir() {
			name := f.Name()
			if strings.LastIndex(name, ".txt") == len(name)-len(".txt") {
				name = name[:len(name)-len(".txt")]
				site.Pages = append(site.Pages, name)
			}
		}
	}
	renderHomeTemplate(w, site)
}

func makeHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m := validPath.FindStringSubmatch(r.URL.Path)
		if m == nil {
			http.NotFound(w, r)
			return
		}
		fn(w, r, m[2])
	}
}

func makeHomeHandler(fn func(http.ResponseWriter, *http.Request, string), name string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fn(w, r, name)
	}
}

func main() {
	//	p1 := &Page{Title: "TestPage", Body: []byte("This is a sample Page.")}
	//	p1.save()
	//	p2, _ := loadPage("TestPage")
	//	fmt.Println(string(p2.Body))
	
	fmt.Println("Initialized")

	http.HandleFunc("/view/", makeHandler(viewHandler))
	http.HandleFunc("/edit/", makeHandler(editHandler))
	http.HandleFunc("/save/", makeHandler(saveHandler))
	http.HandleFunc("/create/", createHandler)
	http.HandleFunc("/", makeHomeHandler(homeHandler, "WiKi"))
	//	pwd, err := os.Getwd()
	//	if err == nil {
	//		http.Handle("/", http.FileServer(http.Dir(pwd+"/"+dataDir)))
	//	}
	http.ListenAndServe(":8080", nil)
}

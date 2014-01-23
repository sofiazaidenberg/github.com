package main

import (
	"errors"
	"fmt"
	"github.com/gorilla/sessions"
	"html"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strings"
)

var tmplDir = "tmpl/"
var dataDir = "data/"
var templates = template.Must(template.ParseFiles(tmplDir+"edit.html", tmplDir+"view.html", tmplDir+"index.html", tmplDir+"create.html"))
var validPath = regexp.MustCompile("^/(edit|save|view|delete)/([a-zA-Z0-9]+)$")
var store = sessions.NewCookieStore([]byte("something-very-secret"))

const PAGE_EXISTS string = "A page with this title already exists. Please choose a different title."
const EMPTY_TITLE string = "Empty title not admitted"
const INVALID_TITLE string = "Invalid Page Title"

type Site struct {
	Name  string
	Pages []string
}

type Page struct {
	Title string
	Body  template.HTML
	Error string
}

func (p *Page) HasError() bool {
	return p.Error != ""
}

func (p *Page) exists() bool {
	file, err := os.Open(dataDir + p.Title + ".txt")
	defer file.Close()
	return !os.IsNotExist(err)
}

func (p *Page) save() error {
	if p.Title == "" {
		return errors.New(EMPTY_TITLE)
	}
	filename := dataDir + p.Title + ".txt"

	return ioutil.WriteFile(filename, []byte(html.EscapeString(string(p.Body))), 0600)
}

func (p *Page) del() error {
	if p.Title == "" {
		return errors.New(EMPTY_TITLE)
	}

	return os.Rename(dataDir+p.Title+".txt", dataDir+p.Title+".deleted")
}

func loadPage(title string) (*Page, error) {
	filename := title + ".txt"
	body, err := ioutil.ReadFile(dataDir + filename)
	if err != nil {
		return nil, err
	}
	return &Page{Title: title, Body: template.HTML(body)}, nil
}

//func handler(w http.ResponseWriter, r *http.Request) {
//	fmt.Fprintf(w, "Hi there, I love %s!", r.URL.Path[1:])
//}

func getTitle(w http.ResponseWriter, r *http.Request) (string, error) {
	m := validPath.FindStringSubmatch(r.URL.Path)
	if m == nil {
		http.NotFound(w, r)
		return "", errors.New(INVALID_TITLE)
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

	p.Body = template.HTML(strings.Replace(string(p.Body), "\n", "<br>", -1))
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
	session, _ := store.Get(r, "session-name")
	p := new(Page)
	if title, ok := session.Values["title"].(string); ok {
		p.Title = title
		session.Values["title"] = ""
	}
	if body, ok := session.Values["body"].(string); ok {
		p.Body = template.HTML(body)
		session.Values["body"] = ""
	}
	if err, ok := session.Values["error"].(string); ok {
		p.Error = err
		session.Values["error"] = ""
	}
	session.Save(r, w)
	renderTemplate(w, "create", p)
}

func saveNewHandler(w http.ResponseWriter, r *http.Request) {
	title := r.FormValue("title")
	if title == "" {
		saveNewErrorHandler(w, r, errors.New(EMPTY_TITLE))
		return
	}
	m := validPath.FindStringSubmatch("/save/" + title)
	if m == nil {
		saveNewErrorHandler(w, r, errors.New(INVALID_TITLE))
		return
	}
	p := &Page{Title: m[2]}
	if p.exists() {
		saveNewErrorHandler(w, r, errors.New(PAGE_EXISTS))
		return
	}
	saveHandler(w, r, m[2])
}

func saveNewErrorHandler(w http.ResponseWriter, r *http.Request, err error) {
	session, _ := store.Get(r, "session-name")
	session.Values["title"] = r.FormValue("title")
	session.Values["body"] = r.FormValue("body")
	session.Values["error"] = fmt.Sprintf("%s", err)
	session.Save(r, w)

	http.Redirect(w, r, "/create/", http.StatusFound)
}

func saveHandler(w http.ResponseWriter, r *http.Request, title string) {
	body := r.FormValue("body")
	p := &Page{Title: title, Body: template.HTML(body)}
	err := p.save()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/view/"+title, http.StatusFound)
}

func deleteHandler(w http.ResponseWriter, r *http.Request, title string) {
	p := &Page{Title: title}
	err := p.del()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/", http.StatusFound)
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

func Test(w http.ResponseWriter, r *http.Request) {
	// Get a session. We're ignoring the error resulted from decoding an
	// existing session: Get() always returns a session, even if empty.
	session, _ := store.Get(r, "session-name")
	// Set some session values.
	session.Values["foo"] = "bar"
	session.Values[42] = 43
	// Save it.
	session.Save(r, w)
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
	http.HandleFunc("/save/new", saveNewHandler)
	http.HandleFunc("/create/", createHandler)
	http.HandleFunc("/delete/", makeHandler(deleteHandler))
	http.HandleFunc("/", makeHomeHandler(homeHandler, "WiKi"))
	//	pwd, err := os.Getwd()
	//	if err == nil {
	//		http.Handle("/", http.FileServer(http.Dir(pwd+"/"+dataDir)))
	//	}
	http.ListenAndServe(":8080", nil)

	//	http.ListenAndServe(":8181", http.HandlerFunc(Test))
}

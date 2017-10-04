// wiki example from https://golang.org/doc/articles/wiki/
package main

import (
	"errors"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"regexp"
)

type Page struct {
	Title string
	Body  []byte
}

// should read from a config
var templatePath string = os.Getenv("GOWIKI_TEMPLATE_PATH") //"/tmp/gowiki/templates/"
const htmlPath string = "/tmp/gowiki/pages/"

// holds pre-cached templates to avoid lookups on every view
var templates = template.Must(
	template.ParseFiles(path.Join(templatePath, "edit.html"), path.Join(templatePath, "view.html")))

// Still super naive since this is a simple example, but this is used with getTitle to ensure valid paths so that
// random paths cannot be read/written path must be /edit/ or /save/ or /view/ followed by only letters and numbers
// this way things like .. cannot be used to do bad things
var validPath = regexp.MustCompile("^/(edit|save|view)/([a-zA-Z0-9]+)$")

func makeHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	// takes the actual function we want to call as an arg, validates the path very similarly to
	// getTitle() and then calls the function we passed in after verifying a valid path.
	return func(w http.ResponseWriter, r *http.Request) {
		// Here we will extract the page title from the Request,
		// and call the provided handler 'fn'

		m := validPath.FindStringSubmatch(r.URL.Path)
		if m == nil {
			http.NotFound(w, r)
			return
		}
		fn(w, r, m[2]) // The title is the second subexpression.
	}
}

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

func (p *Page) save() error {
	// Takes pointer to a page as a receive - like a class method in tradtional OO stuff
	filename := p.Title + ".txt"
	fullPath := path.Join(htmlPath, filename)
	return ioutil.WriteFile(fullPath, p.Body, 0600)
}

func loadPage(title string) (*Page, error) {
	// Takes title string, returns pointer to a page
	filename := title + ".txt"
	body, err := ioutil.ReadFile(path.Join(htmlPath, filename))
	if err != nil {
		return nil, err
	}
	return &Page{Title: title, Body: body}, nil
}

func viewHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(title)
	if err != nil {
		http.Redirect(w, r, "/edit/"+title, http.StatusFound)
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

func main() {
	// p1 := &Page{Title: "TestPage", Body: []byte("This is a sample Page.")}
	// p1.save()
	// p2, _ := loadPage("TestPage")
	// fmt.Println(string(p2.Body))

	os.MkdirAll(htmlPath, os.ModePerm)
	http.HandleFunc("/view/", makeHandler(viewHandler))
	http.HandleFunc("/edit/", makeHandler(editHandler))
	http.HandleFunc("/save/", makeHandler(saveHandler))
	http.ListenAndServe(":8080", nil)
}

package main

import (
	"net/http"
	"io/ioutil"
	"regexp"
	"html/template"
)

type Page struct {
	Title string
	Body  template.HTML
}

func (p *Page) save() error {
	filename := "data/" + p.Title + ".txt"
	return ioutil.WriteFile(filename, []byte(p.Body), 0600)
}

func loadPage(title string) (*Page, error) {
	filename := "data/" + title + ".txt"
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return &Page{Title: title, Body: template.HTML(body)}, nil
}

func viewHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(title)
	if err != nil {
		http.Redirect(w, r, "/edit/"+title, http.StatusFound)
		return
	}
	renderMarkdown(p)
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
	p := &Page{Title: title, Body: template.HTML(body)}
	err := p.save()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/view/"+title, http.StatusFound)
}

func redirHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/view/FrontPage", http.StatusFound)
}


var templates = make(map[string]*template.Template)

func init() {
	for _, tmpl := range []string{"edit", "view"} {
		t := template.Must(template.ParseFiles("tmpl/"+tmpl+".html"))
		templates[tmpl] = t
	}
}

func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
	err := templates[tmpl].Execute(w, p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

const lenPath = len("/view/")

var titleValidator = regexp.MustCompile("^[a-zA-Z0-9]+$")

func makeHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		title := r.URL.Path[lenPath:]
		if !titleValidator.MatchString(title) {
			http.NotFound(w, r)
			return
		}
		fn(w, r, title)
	}
}

func renderMarkdown(p *Page) {
	var h1regexp = regexp.MustCompile("(?m)^# (.+) #")
	p.Body = template.HTML(h1regexp.ReplaceAll([]byte(p.Body), []byte(`<h1>$1</h1>`)))
	var h2regexp = regexp.MustCompile("(?m)^## (.+) ##")
        p.Body = template.HTML(h2regexp.ReplaceAll([]byte(p.Body), []byte(`<h2>$1</h2>`)))
	var h3regexp = regexp.MustCompile("(?m)^### (.+) ###")
        p.Body = template.HTML(h3regexp.ReplaceAll([]byte(p.Body), []byte(`<h3>$1</h3>`)))
	var h4regexp = regexp.MustCompile("(?m)^#### (.+) ####")
        p.Body = template.HTML(h4regexp.ReplaceAll([]byte(p.Body), []byte(`<h4>$1</h4>`)))
	var h5regexp = regexp.MustCompile("(?m)^##### (.+) #####")
        p.Body = template.HTML(h5regexp.ReplaceAll([]byte(p.Body), []byte(`<h5>$1</h5>`)))
	var h6regexp = regexp.MustCompile("(?m)^###### (.+) ######")
        p.Body = template.HTML(h6regexp.ReplaceAll([]byte(p.Body), []byte(`<h5>$1</h5>`)))

	var italicregexp = regexp.MustCompile(`(?m) \*(.+)\*`)
        p.Body = template.HTML(italicregexp.ReplaceAll([]byte(p.Body), []byte(`<i>$1</i>`)))
	var boldregexp = regexp.MustCompile(`(?m)\*\*(.+)\*\*`)
        p.Body = template.HTML(boldregexp.ReplaceAll([]byte(p.Body), []byte(`<b>$1</b>`)))
	var striketroughregexp = regexp.MustCompile("(?m)~(.+)~")
        p.Body = template.HTML(striketroughregexp.ReplaceAll([]byte(p.Body), []byte(`<s>$1</s>`)))

	var hrregexp = regexp.MustCompile(`(?m)^[*-_]{3,}$`)
        p.Body = template.HTML(hrregexp.ReplaceAll([]byte(p.Body), []byte(`<hr>`)))
}

func main() {
	http.HandleFunc("/", redirHandler)
	http.HandleFunc("/view/", makeHandler(viewHandler))
	http.HandleFunc("/edit/", makeHandler(editHandler))
	http.HandleFunc("/save/", makeHandler(saveHandler))
	http.ListenAndServe(":8080", nil)
}

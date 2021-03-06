// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"text/template"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
	"fmt"
	"expvar"
)

type Page struct {
	Title string
	Body  []byte
}

var tmplDir = "tmpl"
var dataDir = "data"
var actions = []string{}
var validPath *regexp.Regexp
var templates = map[string]*template.Template{
	"view": nil,
	"edit": nil,
}
var baseTmplName = "base";
var (
  counts = expvar.NewMap("counters")
)

func main() {
	http.HandleFunc("/view/", makeHandler(viewHandler))
	actions = append(actions, "view")
	http.HandleFunc("/edit/", makeHandler(editHandler))
	actions = append(actions, "edit")
	http.HandleFunc("/save/", makeHandler(saveHandler))
	actions = append(actions, "save")
	http.HandleFunc("/", rootHandler)

	validPath = regexp.MustCompile("^/(" + strings.Join(actions, "|") + ")/([a-zA-Z0-9]+)$")
	
	for k, _ := range templates { 
	    templates[k] =
		    template.Must(template.ParseFiles(tmplDir + "/" + k + ".html", tmplDir + "/" + baseTmplName + ".html"))
	}

	http.ListenAndServe(":8080", nil)
}

func (p *Page) save() error {
	filename := dataDir + "/" + p.Title + ".txt"
	return ioutil.WriteFile(filename, p.Body, 0600)
}

func loadPage(title string) (*Page, error) {
	filename := dataDir + "/" + title + ".txt"
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return &Page{Title: title, Body: body}, nil
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/view/FrontPage", http.StatusFound)
	return
}

func viewHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(title)
	if err != nil {
		http.Redirect(w, r, "/edit/"+title, http.StatusFound)
		return
	}
	pageLink := regexp.MustCompile("\\[(.+)\\]")
	p.Body = pageLink.ReplaceAllFunc(p.Body, func(match []byte) []byte {
		title := match[1 : len(match)-1]
		ret := fmt.Sprintf("<a href=\"%s\">%s</a>", title, title)
		return []byte(ret)
	})
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

func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
	err := templates[tmpl].ExecuteTemplate(w, baseTmplName, p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
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



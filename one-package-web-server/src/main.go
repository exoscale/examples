package main

import (
	"html/template"
	"net/http"
	"os"
)

type MultipleDomains map[string]http.Handler

func (md MultipleDomains) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	handler := md[r.Host]

	if handler != nil {
		handler.ServeHTTP(w, r)
	} else {
		http.Redirect(w, r, "/", http.StatusFound)
	}
}

// templateHandler assigns each URL to the corresponding template structure.
func templateHandler(w http.ResponseWriter, r *http.Request) {
	layout := "../structure.html"
	content := "../pages/" + r.URL.Path + ".html"
	if r.URL.Path == "/" {
		content = "../pages/index.html"
	}

	info, err := os.Stat(content)
	if err != nil {
		if os.IsNotExist(err) {
			http.NotFound(w, r)
			return
		}
	}

	if info.IsDir() {
		http.NotFound(w, r)
		return
	}

	tmpl, err := template.ParseFiles(layout, content)
	if err != nil {
		http.Error(w, http.StatusText(500), 500)
		return
	}
	tmpl.ExecuteTemplate(w, "structure", nil)
}

// main starts the web server and routes URLS.
func main() {
	muxex := http.NewServeMux()

	md := make(MultipleDomains)
	md["localhost:8080"] = muxex

	muxex.Handle("/public/images/", http.StripPrefix("/public/images/", http.FileServer(http.Dir("../public/images"))))
	muxex.Handle("/public/", http.StripPrefix("/public/", http.FileServer(http.Dir("../public"))))
	muxex.HandleFunc("/", templateHandler)

	http.ListenAndServe(":8080", md)
}

package main

import (
	"flag"
	"html/template"
	"net/http"
	"os"
)

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
	listenAddress := flag.String("host", "localhost:8080", "address and port to listen on, type -host=yourdomain.com:80 for accessing online")

	flag.Parse()

	http.Handle("/public/images/", http.StripPrefix("/public/images/", http.FileServer(http.Dir("../public/images"))))
	http.Handle("/public/", http.StripPrefix("/public/", http.FileServer(http.Dir("../public"))))
	http.HandleFunc("/", templateHandler)

	http.ListenAndServe(*listenAddress, nil)
}

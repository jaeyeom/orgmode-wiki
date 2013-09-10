// Binary run runs a simple wiki server.
package main

import (
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/jaeyeom/gofiletable/table"
)

var (
	addr = flag.String("addr", ":8000", "address of server")
	dbPath = flag.String("db_path", "/tmp/wiki-db/", "Path to wiki db")
)

var wikiTable *table.Table

type document struct {
	Title       string
	ContentHtml string
}

// viewHandler is view/edit page handler.
func viewHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" {
		// Index page
		http.Redirect(w, r, "/view/Main", http.StatusSeeOther)
		return
	}
	action := r.URL.Path[1:5]  // view or edit
	title := r.URL.Path[6:]
	if r.Method != "GET" || title == "" {
		return
	}
	content, _ := wikiTable.Get([]byte(title))
	w.Header().Set("Content-Type", "text/html")
	t, _ := template.ParseFiles(fmt.Sprintf("template/%s.html", action))
	t.Execute(w, document{
		Title: title,
		ContentHtml: string(content),
	})
}

// modifyHandler is modify action handler.
func modifyHandler(w http.ResponseWriter, r *http.Request) {
	title := r.URL.Path[8:]
	if r.Method != "POST" || title == "" {
		return
	}
	content := r.PostFormValue("content")
	if err := wikiTable.Put([]byte(title), []byte(content)); err != nil {
		log.Println(err)
		return
	}
	http.Redirect(w, r, "/view/"+title, http.StatusSeeOther)
}

func main() {
	flag.Parse()
	var err error
	wikiTable, err = table.Create(table.TableOption{*dbPath, nil})
	if err != nil {
		log.Println(err)
		return
	}
	http.HandleFunc("/", viewHandler)
	http.HandleFunc("/view/", viewHandler)
	http.HandleFunc("/edit/", viewHandler)
	http.HandleFunc("/modify/", modifyHandler)
	http.ListenAndServe(*addr, nil)
}

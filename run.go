// Binary run runs a simple wiki server.
package main

import (
	"bytes"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/jaeyeom/gofiletable/table"
	"github.com/jaeyeom/orgmode-wiki/parser"
)

var (
	addr   = flag.String("addr", ":8000", "address of server")
	dbPath = flag.String("db_path", "/tmp/wiki-db/", "Path to wiki db")
)

var wikiTable *table.Table

type document struct {
	Title       string
	Content     string
	ContentHTML template.HTML
}

// viewHandler is view/edit page handler.
func viewHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" {
		// Index page
		http.Redirect(w, r, "/view/Main", http.StatusSeeOther)
		return
	}
	action := r.URL.Path[1:5] // view or edit
	title := r.URL.Path[6:]
	if r.Method != "GET" || title == "" {
		return
	}
	content, _ := wikiTable.Get([]byte(title))
	if len(content) == 0 && action == "view" {
		// Index page
		http.Redirect(w, r, "/edit/"+title, http.StatusSeeOther)
		return
	}
	w.Header().Set("Content-Type", "text/html")
	t, _ := template.ParseFiles(fmt.Sprintf("template/%s.html", action))

	p := parser.Parser{}
	p.ParseDocument(bytes.NewBuffer(content))
	contentHTML := &bytes.Buffer{}
	p.Write(parser.NewHTMLWriter(contentHTML), false)

	t.Execute(w, document{
		Title:       title,
		Content:     string(content),
		ContentHTML: template.HTML(contentHTML.String()),
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
	log.Printf("Server being ready: %s\n", *addr)
	if err := http.ListenAndServe(*addr, nil); err != nil {
		log.Println(err)
	}
}

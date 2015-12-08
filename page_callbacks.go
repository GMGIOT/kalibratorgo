package main

import (
	"html/template"
	"net/http"
	//"net/url"
	"fmt"
	"path"
	"runtime"
	//"encoding/json"
	"log"
	//"strings"
	//"strconv"
	//"errors"
)

const (
	OkAnsver = "OK"

	SERVER_VERSION = "0.0.1"
)

type HaderPageData struct {
	Version string
}

var currentPath string
var fileserverHandler http.Handler

func FillPagesMap(m *map[string]func(http.ResponseWriter, *http.Request)) {
	(*m)["/"] = indexHandlr
	(*m)["/asserts"] = assertsServer

	_, filename, _, _ := runtime.Caller(1)
	currentPath = path.Dir(filename)

	fileserverHandler = http.FileServer(http.Dir(currentPath))
}

func patchPath(name1 string, names ...string) []string {

	res := make([]string, len(names)+1)

	res[0] = currentPath + "/" + name1
	for i, na := range names {
		res[i+1] = currentPath + "/" + na
	}

	return res
}

func indexHandlr(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles(patchPath(
		"templates/header.html",
		"templates/index.html",
		"templates/footer.html")...)

	if err != nil {
		fmt.Fprintf(w, "Error %s", err.Error())
		return
	}

	if err := t.ExecuteTemplate(w, "index", nil); err != nil {
		log.Println(err.Error())
	}
}

func assertsServer(w http.ResponseWriter, r *http.Request) {
	fileserverHandler.ServeHTTP(w, r)
}

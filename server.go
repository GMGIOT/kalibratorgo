package main

import (
	"net/http"
	"io"
	"fmt"
	"log"
	"strings"
	"sort"
)

var Log_requests = true

type KalibratorGoHttpServer struct {
	http.Server
	mux           map[string]func(http.ResponseWriter, *http.Request) // карта страница - обработчик
	surtedURLList []string
}

// Это пустой класс, соответствующий интерфейсу myHandler
type serverHandler struct { 
	server *KalibratorGoHttpServer
} 

func (this *serverHandler) findBestMatchHandler(url string) func(http.ResponseWriter, *http.Request) {
	    
    var result func(http.ResponseWriter, *http.Request)
    
    for _, template := range this.server.surtedURLList {
    	if strings.HasPrefix(url, template) {
    		result = this.server.mux[template]
    	}
    }
    
    return result
}

// метод ServeHTTP
func (this *serverHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	url := r.URL.String()
	
	h := this.findBestMatchHandler(url)

	if h != nil {
		if Log_requests {
			log.Printf("Processing url: %s -> %#x", url, h)
		}
		h(w, r)
		return
	}
	
	// default ansver
	io.WriteString(w, "KalibratorGo server: "+url)
}

func GetPathsList(m *map[string]func(http.ResponseWriter, *http.Request)) []string {
	keys := make([]string, len(*m))
	i := 0
    for k,_ := range ( *m ) {
    	keys[i] = k
    	i++
    }
    sort.Strings(keys)
    
    return keys
}

func KalibratorGoServer(port int) *KalibratorGoHttpServer {
	res := new(KalibratorGoHttpServer) 
	
	res.Addr = fmt.Sprintf(":%d", port)
	res.Handler = &serverHandler{res}
	
	res.mux = make(map[string]func(http.ResponseWriter, *http.Request))
	FillPagesMap(&res.mux)
	res.surtedURLList = GetPathsList(&res.mux)
	
	return res
}
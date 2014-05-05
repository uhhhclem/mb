package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"mb"
	"strings"
)

var g *mb.Game

func appHandler(w http.ResponseWriter, r *http.Request) {
	filename := r.URL.Path
	if len(filename) > 0 && strings.HasPrefix(filename, "/") {
		filename = filename[1:]
	}
	if filename == "" {
		filename = "index.html"
	}
	filename = strings.ToLower("../client/app/" + filename)
	if strings.HasSuffix(filename, ".js") {
		w.Header()["Content-Type"] = []string{"application/javascript"}
	}
	if strings.HasSuffix(filename, ".css") {
		w.Header()["Content-Type"] = []string{"text/css"}
	}
	if strings.HasSuffix(filename, ".jpg") {
		w.Header()["Content-Type"] = []string{"image/JPEG"}
	}
	log.Printf("%s\n", filename)
	if body, err := ioutil.ReadFile(filename); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	} else {
		fmt.Fprint(w, string(body))
	}
}

func mbBoardHandler(w http.ResponseWriter, q *http.Request) {
	if q.Method == "POST" {
		reader := io.LimitReader(q.Body, 1000)
		if b, err := ioutil.ReadAll(reader); err != nil {
			log.Println(err)
			return
		} else {
			r := &mb.Request{}
			if err := json.Unmarshal(b, &r); err != nil {
				log.Println(err)			
			} else {
				g.HandleRequest(*r)
			}			
		}
	}

	type response struct {
		Board mb.Board
		Error string
		Prompt string
	}
	r := &response{Board: g.Board}

	if g.Response != nil {
		r.Prompt = string(g.Response.Prompt)
		if g.Response.Error != nil {
			r.Error = g.Response.Error.Error()
		}
	}
	
	b, err := json.Marshal(r)
	if err != nil {
		fmt.Fprint(w, err)
	} else {
		fmt.Fprint(w, bytes.NewBuffer(b).String())
	}
}

func mbLogHandler(w http.ResponseWriter, r *http.Request) {
	b, err := json.Marshal(g.Log)
	if err != nil {
		fmt.Fprint(w, err)
	} else {
		fmt.Fprint(w, bytes.NewBuffer(b).String())
	}	
}

func main() {
	g = mb.NewGame()
	g.StartGame()
	
    http.HandleFunc("/", appHandler)
    http.HandleFunc("/mb/board/", mbBoardHandler)
    http.HandleFunc("/mb/log/", mbLogHandler)
    http.ListenAndServe(":8080", nil)
}
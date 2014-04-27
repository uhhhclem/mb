package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"mb"
	"strings"
)

var g *mb.Game

func appHandler(w http.ResponseWriter, r *http.Request) {
	filename := r.URL.Path[len("/app/"):]
	filename = strings.ToLower("../client/app/" + filename)
	if strings.HasSuffix(filename, ".js") {
		w.Header()["Content-Type"] = []string{"application/javascript"}
	}
	if strings.HasSuffix(filename, ".css") {
		w.Header()["Content-Type"] = []string{"text/css"}
	}
	log.Printf("%s\n", filename)
	if body, err := ioutil.ReadFile(filename); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	} else {
		fmt.Fprint(w, string(body))
	}
}

func mbBoardHandler(w http.ResponseWriter, r *http.Request) {
	b, err := json.Marshal(g.Board)
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
	

    http.HandleFunc("/app/", appHandler)
    http.HandleFunc("/mb/board/", mbBoardHandler)
    http.HandleFunc("/mb/log/", mbLogHandler)
    http.ListenAndServe(":8080", nil)
}
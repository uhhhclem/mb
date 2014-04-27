package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"mb"
)

var g *mb.Game

func viewHandler(w http.ResponseWriter, r *http.Request) {
	b, err := json.Marshal(g.Board)
	if err != nil {
		fmt.Fprint(w, err)
	} else {
		fmt.Fprint(w, bytes.NewBuffer(b).String())
	}
}

func main() {
	g = mb.NewGame()
	g.StartGame()
	

    http.HandleFunc("/", viewHandler)
    http.ListenAndServe(":8080", nil)
}
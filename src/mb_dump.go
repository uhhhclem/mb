package main

import (
	"bufio"
	"fmt"
	"mb"
	"os"
	"strings"

//	"encoding/json"
)

func main() {
	g := mb.NewGame()
	g.StartGame()
	for g.Response != nil {
		g.Request.Input = ""
		if g.Response.Error != nil {
			fmt.Printf("\nError: %s\n", g.Response.Error)
		}
		reader := bufio.NewReader(os.Stdin)
		// this is pretty hacky, but it'll do for keyboard input
		fmt.Print("\n" + g.Response.Prompt + "> ")
		s, _ := reader.ReadString('\n')
		line := strings.Split(s, "\r")[0]
		g.HandleRequest(mb.Request{Input: mb.Input(line)})
	}
	fmt.Println("\n\nEnd of game")
}

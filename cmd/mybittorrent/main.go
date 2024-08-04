package main

import (
	"encoding/json"
	"fmt"
	bencode "github.com/jackpal/bencode-go" // Available if you need it!
	"os"
	"strings"
)

func main() {
	command := os.Args[1]

	if command == "decode" {
		bencodedValue := strings.NewReader(os.Args[2])

		decoded, err := bencode.Decode(bencodedValue)
		if err != nil {
			fmt.Println(err)
			return
		}

		jsonOutput, _ := json.Marshal(decoded)
		fmt.Println(string(jsonOutput))
	} else if command == "info" {
		dat, err := os.ReadFile(os.Args[2])

		if err != nil {
			fmt.Println(err)
			return
		}

		torrent, err := NewTorrentFile(dat)

		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Printf(torrent.ToString())
	} else {
		fmt.Println("Unknown command: " + command)
		os.Exit(1)
	}
}

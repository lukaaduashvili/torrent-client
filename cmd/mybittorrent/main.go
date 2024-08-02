package main

import (
	"encoding/json"
	"fmt"
	"github.com/codecrafters-io/bittorrent-starter-go/cmd/bencode"
	"os"
	// bencode "github.com/jackpal/bencode-go" // Available if you need it!
)

func main() {
	command := os.Args[1]

	if command == "decode" {
		bencodedValue := os.Args[2]

		decoded, _, err := bencode.Decode(bencodedValue)
		if err != nil {
			fmt.Println(err)
			return
		}

		jsonOutput, _ := json.Marshal(decoded)
		fmt.Println(string(jsonOutput))
	} else if command == "info" {
		dat, err := os.ReadFile(os.Args[2])

		decoded, _, err := bencode.Decode(string(dat))
		if err != nil {
			fmt.Println(err)
			return
		}

		jsonOutput, _ := json.Marshal(decoded)
		torrent := TorrentFile{}

		err = json.Unmarshal(jsonOutput, &torrent)
		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Printf("Tracker URL: %s\n", torrent.Announce)
		fmt.Printf("Length: %d\n", torrent.Info.Length)
	} else {
		fmt.Println("Unknown command: " + command)
		os.Exit(1)
	}
}

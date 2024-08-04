package main

import (
	"encoding/json"
	"fmt"
	"github.com/codecrafters-io/bittorrent-starter-go/cmd/peer"
	"github.com/codecrafters-io/bittorrent-starter-go/cmd/torrentfile"
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

		torrent, err := torrentfile.NewTorrentFile(dat)

		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Printf(torrent.ToString())
	} else if command == "peers" {
		dat, err := os.ReadFile(os.Args[2])

		if err != nil {
			fmt.Println(err)
			return
		}

		torrent, err := torrentfile.NewTorrentFile(dat)

		if err != nil {
			fmt.Println(err)
			return
		}

		peerResource := peer.NewTrackerResource(*torrent)

		peerResource.GetPeers()

		for _, peerAddress := range peerResource.Peers {
			fmt.Printf("%s\n", peerAddress)
		}
	} else {
		fmt.Println("Unknown command: " + command)
		os.Exit(1)
	}
}

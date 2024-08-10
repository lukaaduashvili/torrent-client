package main

import (
	"encoding/json"
	"fmt"
	"github.com/codecrafters-io/bittorrent-starter-go/cmd/peer"
	"github.com/codecrafters-io/bittorrent-starter-go/cmd/torrentfile"
	bencode "github.com/jackpal/bencode-go" // Available if you need it!
	"os"
	"strconv"
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
	} else if command == "handshake" {
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

		//peerResource.GetPeers()

		peerAddress := os.Args[3]

		peerResource.InitiateHandshake(peerAddress)
	} else if command == "download_piece" {

		var torrentFile, outputPath string

		if os.Args[2] == "-o" {
			torrentFile = os.Args[4]
			outputPath = os.Args[3]
		} else {
			torrentFile = os.Args[3]
			outputPath = "."
		}

		dat, err := os.ReadFile(torrentFile)

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

		peerResource.InitiateHandshake(peerResource.Peers[0])

		ind, _ := strconv.Atoi(os.Args[5])
		data := peerResource.DownloadChunk(peerResource.Peers[0], ind)

		file, err := os.Create(outputPath)
		if err != nil {
			fmt.Println(err)
			return
		}
		defer file.Close()
		_, err = file.Write(data)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Printf("Piece downloaded to %s.\n", outputPath)
	} else {
		fmt.Println("Unknown command: " + command)
		os.Exit(1)
	}
}

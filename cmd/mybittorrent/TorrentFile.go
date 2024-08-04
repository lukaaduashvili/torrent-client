package main

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"github.com/jackpal/bencode-go"
)

type TorrentFile struct {
	Announce string     `bencode:"announce"`
	Info     InfoStruct `bencode:"info"`
	InfoHash [20]byte
}

type InfoStruct struct {
	Length      int    `bencode:"length"`
	Name        string `bencode:"name"`
	PieceLength int    `bencode:"piece length"`
	Pieces      string `bencode:"pieces"`
}

func NewTorrentFile(data []byte) (*TorrentFile, error) {
	torrent := TorrentFile{}

	err := bencode.Unmarshal(bytes.NewReader(data), &torrent)

	if err != nil {
		return nil, err
	}

	torrent.setInfoMD1Hash()

	return &torrent, nil
}

func (torrent *TorrentFile) setInfoMD1Hash() {
	var buf bytes.Buffer
	info := torrent.Info
	err := bencode.Marshal(&buf, info)

	if err != nil {
		fmt.Println(err)
		return
	}

	h := sha1.Sum(buf.Bytes())

	torrent.InfoHash = h
}

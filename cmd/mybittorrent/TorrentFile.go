package main

type TorrentFile struct {
	Announce string     `json:"announce"`
	Info     InfoStruct `json:"info"`
}

type InfoStruct struct {
	Length      int    `json:"length"`
	Name        string `json:"name"`
	PieceLength int    `json:"piece length"`
	Pieces      string `json:"pieces"`
}

package main

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"github.com/jackpal/bencode-go"
)

type TorrentFile struct {
	Announce    string
	Info        InfoStruct
	InfoHash    [20]byte
	PieceHashes [][20]byte
}

type InfoStruct struct {
	Length      int
	Name        string
	PieceLength int
	Pieces      string
}

type encodedTorrentFile struct {
	Announce string            `bencode:"announce"`
	Info     encodedInfoStruct `bencode:"info"`
}

type encodedInfoStruct struct {
	Length      int    `bencode:"length"`
	Name        string `bencode:"name"`
	PieceLength int    `bencode:"piece length"`
	Pieces      string `bencode:"pieces"`
}

func NewTorrentFile(data []byte) (*TorrentFile, error) {
	torrent := TorrentFile{}
	encodedFile := encodedTorrentFile{}

	err := bencode.Unmarshal(bytes.NewReader(data), &encodedFile)

	if err != nil {
		return nil, err
	}

	infoFile := InfoStruct{}
	infoFile.Length = encodedFile.Info.Length
	infoFile.Name = encodedFile.Info.Name
	infoFile.Pieces = encodedFile.Info.Pieces
	infoFile.PieceLength = encodedFile.Info.PieceLength

	torrent.Announce = encodedFile.Announce
	torrent.Info = infoFile

	torrent.InfoHash, err = getInfoMD1Hash(encodedFile.Info)

	if err != nil {
		return nil, err
	}

	torrent.PieceHashes, err = setPieceHashes(encodedFile.Info)

	if err != nil {
		return nil, err
	}

	return &torrent, nil
}

func getInfoMD1Hash(encodedInfoStruct encodedInfoStruct) ([20]byte, error) {
	sha := sha1.New()
	err := bencode.Marshal(sha, encodedInfoStruct)
	var hash [20]byte

	if err != nil {
		return hash, err
	}

	h := sha.Sum(nil)

	copy(hash[:], h[:20])

	return hash, nil
}

func setPieceHashes(encodedInfoStruct encodedInfoStruct) ([][20]byte, error) {
	hashLen := 20 // Length of SHA-1 hash
	buf := []byte(encodedInfoStruct.Pieces)

	numHashes := len(buf) / hashLen
	hashes := make([][20]byte, numHashes)

	if len(buf)%hashLen != 0 {
		err := fmt.Errorf("Received malformed pieces of length %d", len(buf))
		return hashes, err
	}

	for i := 0; i < numHashes; i++ {
		copy(hashes[i][:], buf[i*hashLen:(i+1)*hashLen])
	}

	return hashes, nil
}

func (torrent *TorrentFile) ToString() string {
	res := ""
	res += fmt.Sprintf("Tracker URL: %s\n", torrent.Announce)
	res += fmt.Sprintf("Length: %d\n", torrent.Info.Length)
	res += fmt.Sprintf("Info Hash: %x\n", torrent.InfoHash)
	res += fmt.Sprintf("Piece Length: %d\n", torrent.Info.PieceLength)
	res += fmt.Sprintf("Piece Hashes:\n")
	for _, value := range torrent.PieceHashes {
		res += fmt.Sprintf("%x\n", value)
	}

	return res
}

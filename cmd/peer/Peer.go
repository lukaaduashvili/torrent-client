package peer

import (
	"encoding/binary"
	"fmt"
	"github.com/codecrafters-io/bittorrent-starter-go/cmd/torrentfile"
	"github.com/jackpal/bencode-go"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
)

type TrackerResource struct {
	file            torrentfile.TorrentFile
	peerId          string
	port            int
	uploaded        int
	downloaded      int
	left            int
	compact         int
	RequestInterval int
	Peers           []string
}

func NewTrackerResource(file torrentfile.TorrentFile) *TrackerResource {
	resource := TrackerResource{}
	resource.file = file
	resource.peerId = randSeq(20)
	resource.port = 6881
	resource.uploaded = 0
	resource.downloaded = 0
	resource.left = file.Info.Length
	resource.compact = 1

	return &resource
}

func (resource *TrackerResource) GetPeers() {
	parm := url.Values{}

	parm.Add("info_hash", string(resource.file.InfoHash[:]))
	parm.Add("peer_id", resource.peerId)
	parm.Add("port", strconv.Itoa(resource.port))
	parm.Add("uploaded", strconv.Itoa(resource.uploaded))
	parm.Add("downloaded", strconv.Itoa(resource.downloaded))
	parm.Add("left", strconv.Itoa(resource.left))
	parm.Add("compact", strconv.Itoa(resource.compact))

	resp, err := http.Get(resource.file.Announce + "?" + parm.Encode())

	if err != nil {
		fmt.Printf("error making http request: %s\n", err)
		os.Exit(1)
	}

	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	res, err := bencode.Decode(strings.NewReader(string(body)))

	resource.RequestInterval = int(res.(map[string]interface{})["interval"].(int64))
	encodedPeers := []byte(res.(map[string]interface{})["peers"].(string))

	for i := 0; i < len(encodedPeers); i += 6 {
		currPeer := ""
		for j := 0; j < 4; j++ {
			currPeer += strconv.Itoa(int(encodedPeers[i+j]))
			if j != 3 {
				currPeer += "."
			}
		}
		currPeer += ":"
		currPeer += strconv.Itoa(int(binary.BigEndian.Uint16(encodedPeers[i+4 : i+6])))
		resource.Peers = append(resource.Peers, currPeer)
	}
}

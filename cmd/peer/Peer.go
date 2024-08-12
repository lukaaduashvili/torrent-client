package peer

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/codecrafters-io/bittorrent-starter-go/cmd/torrentfile"
	"github.com/jackpal/bencode-go"
	"io"
	"math"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
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
	PeerConnections map[string]net.Conn
}

const BITFIELD = 5
const INTERESTED = 2
const UNCHOKE = 1
const REQUEST = 6
const PIECE = 7
const BLOCK_SIZE = 16 * 1024

type Message struct {
	lengthPrefix uint32
	id           uint8
	index        uint32
	begin        uint32
	length       uint32
}

func NewTrackerResource(file torrentfile.TorrentFile) *TrackerResource {
	resource := TrackerResource{}
	resource.file = file
	resource.peerId = "00112233445566778899"
	resource.port = 6881
	resource.uploaded = 0
	resource.downloaded = 0
	resource.left = file.Info.Length
	resource.compact = 1
	resource.PeerConnections = make(map[string]net.Conn)
	resource.Peers = make([]string, 0)

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
		resource.PeerConnections[currPeer] = nil
	}
}

func (resource *TrackerResource) InitiateHandshake(conn net.Conn, peer string) error {
	var handshake []byte
	handshake = append(handshake, byte(19))
	handshake = append(handshake, []byte("BitTorrent protocol")...)
	handshake = append(handshake, []byte{0, 0, 0, 0, 0, 0, 0, 0}...)
	handshake = append(handshake, resource.file.InfoHash[:]...)
	handshake = append(handshake, resource.peerId...)
	// Send the byte array
	_, err := conn.Write(handshake)
	if err != nil {
		return fmt.Errorf("Error sending data: %s", err.Error())
	}

	buffer := make([]byte, 1024)

	_, err = conn.Read(buffer)
	if err != nil {
		return fmt.Errorf("Error receiving data during handshake: %s", err.Error())
	}

	fmt.Printf("Peer ID: %x\n", buffer[48:68])

	return nil
}

func (resource *TrackerResource) DownloadChunk(conn net.Conn, peer string, pieceIndex int) []byte {
	defer conn.Close()

	buf := make([]byte, 4)
	_, err := conn.Read(buf)

	if err != nil {
		fmt.Println(err)
		return nil
	}

	peerMessage := Message{}

	peerMessage.lengthPrefix = binary.BigEndian.Uint32(buf)

	payloadBuf := make([]byte, peerMessage.lengthPrefix)
	_, err = conn.Read(payloadBuf)

	if err != nil {
		fmt.Println(err)
		return nil
	}

	if len(payloadBuf) < 1 {
		fmt.Println("Invalid payload")
		return nil
	}

	peerMessage.id = payloadBuf[0]

	if peerMessage.id != 5 {
		fmt.Println("Did not receive bitfield message ending download")
		return nil
	}

	_, err = conn.Write([]byte{0, 0, 0, 1, 2})

	if err != nil {
		fmt.Println(err)
		return nil
	}

	_, err = conn.Read(buf)

	if err != nil {
		fmt.Println(err)
		return nil
	}

	peerMessage.lengthPrefix = binary.BigEndian.Uint32(buf)
	payloadBuf = make([]byte, peerMessage.lengthPrefix)

	_, err = conn.Read(payloadBuf)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	peerMessage.id = payloadBuf[0]

	if peerMessage.id != 1 {
		fmt.Println("Did not receive unchoke message ending download")
		return nil
	}

	//Send actual requests for the file chunks

	pieceSize := resource.file.Info.PieceLength
	numPieces := int(math.Ceil(float64(resource.file.Info.Length) / float64(pieceSize)))

	if pieceIndex == numPieces-1 {
		pieceSize = resource.file.Info.Length % resource.file.Info.PieceLength
	}

	blockCount := int(math.Ceil(float64(pieceSize) / float64(BLOCK_SIZE)))

	fmt.Printf("File Length: %d Piece Length: %d Piece Count: %d Block Size: %d Block Count: %d\n", resource.file.Info.Length, pieceSize, numPieces, BLOCK_SIZE, blockCount)

	var fileData []byte

	for i := 0; i < blockCount; i++ {
		blockLen := BLOCK_SIZE

		if i == blockCount-1 {
			blockLen = pieceSize - ((blockCount - 1) * BLOCK_SIZE)
		}

		peerMessage := Message{
			lengthPrefix: 13,
			id:           6,
			index:        uint32(pieceIndex),
			begin:        uint32(i * BLOCK_SIZE),
			length:       uint32(blockLen),
		}

		var buffer bytes.Buffer
		binary.Write(&buffer, binary.BigEndian, peerMessage)
		_, err = conn.Write(buffer.Bytes())
		if err != nil {
			fmt.Println(err)
			return nil
		}

		resultPrefix := make([]byte, 4)
		_, err = conn.Read(resultPrefix)

		if err != nil {
			fmt.Println(err)
			return nil
		}

		peerMessage = Message{}
		peerMessage.lengthPrefix = binary.BigEndian.Uint32(resultPrefix)
		payloadBuf := make([]byte, peerMessage.lengthPrefix)
		_, err = io.ReadFull(conn, payloadBuf)
		if err != nil {
			fmt.Println(err)
			return nil
		}
		peerMessage.id = payloadBuf[0]
		fmt.Printf("Received message: %v\n", peerMessage)
		fileData = append(fileData, payloadBuf[9:]...)
	}

	return fileData
}

func (resource *TrackerResource) DownloadFile() ([]byte, error) {
	var fileData []byte

	pieceSize := resource.file.Info.PieceLength
	numPieces := int(math.Ceil(float64(resource.file.Info.Length) / float64(pieceSize)))

	for pieceNumber := 0; pieceNumber < numPieces; pieceNumber++ {
		fmt.Printf("Downloading piece number %d\n", pieceNumber)
		for _, peerAddress := range resource.Peers {
			fmt.Printf("Attempting to download from peer %s\n", peerAddress)

			conn, err := net.DialTimeout("tcp", peerAddress, time.Second*1)

			if err != nil {
				fmt.Println(err)
				continue
			}

			err = resource.InitiateHandshake(conn, peerAddress)

			if err != nil {
				fmt.Println(err)
				continue
			}

			data := resource.DownloadChunk(conn, peerAddress, pieceNumber)
			if len(data) != 0 {
				fileData = append(fileData, data...)
				break
			}
		}
	}
	return fileData, nil
}

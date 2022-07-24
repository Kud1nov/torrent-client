package torrentfile

import (
	"bytes"
	"crypto/rand"
	"crypto/sha1"
	"fmt"
	"os"
	"strings"

	"github.com/Kud1nov/torrent-client/p2p"
	"github.com/Kud1nov/torrent-client/peers"

	"github.com/jackpal/bencode-go"
)

// Port to listen on
const Port uint16 = 6881

// TorrentFile encodes the metadata from a .torrent file
type TorrentFile struct {
	AnnounceList  []string
	InfoHash      [20]byte
	PieceHashes   [][20]byte
	PieceLength   int
	Length        int
	Name          string
	PeersInterval int
	MultipleFile  bool
}

type bencodeInfoFiles struct {
	Length int      `bencode:"length"`           // Длина файла в байтах
	MD5sum string   `bencode:"md5sum,omitempty"` // (optional) Соответствующая сумме MD5 файла
	Path   []string `bencode:"path"`             // Путь и имя файла. Пример, []string{"dir1", "dir2", "file.ext"} => "dir1/dir2/file.ext"
}

type bencodeInfo struct {
	Files       []bencodeInfoFiles `bencode:"files,omitempty"`
	Pieces      string             `bencode:"pieces"`
	PieceLength int                `bencode:"piece length"`
	Length      int                `bencode:"length,omitempty"`
	Name        string             `bencode:"name"`
}

type bencodeTorrent struct {
	Announce     string      `bencode:"announce"`
	AnnounceList [][]string  `bencode:"announce-list,omitempty"`
	Encoding     string      `bencode:"encoding,omitempty"`
	Info         bencodeInfo `bencode:"info"`
}

// DownloadToFile downloads a torrent and writes it to a file
func (t *TorrentFile) DownloadToFile(path string) error {
	var peerID [20]byte
	_, err := rand.Read(peerID[:])
	if err != nil {
		return err
	}

	var Peers []peers.Peer

	for i := range t.AnnounceList {
		peersChunk, err := t.requestPeers(t.AnnounceList[i], peerID, Port)
		if err != nil {
			return err
		}

		Peers = append(Peers, peersChunk...)
	}

	torrent := p2p.Torrent{
		Peers:       Peers,
		PeerID:      peerID,
		InfoHash:    t.InfoHash,
		PieceHashes: t.PieceHashes,
		PieceLength: t.PieceLength,
		Length:      t.Length,
		Name:        t.Name,
	}
	buf, err := torrent.Download()
	if err != nil {
		return err
	}

	outFile, err := os.Create(path)
	if err != nil {
		return err
	}
	defer outFile.Close()
	_, err = outFile.Write(buf)
	if err != nil {
		return err
	}
	return nil
}

// Open parses a torrent file
func Open(path string) (TorrentFile, error) {
	file, err := os.Open(path)
	if err != nil {
		return TorrentFile{}, err
	}

	defer func() {
		_ = file.Close()
	}()

	bto := bencodeTorrent{}
	err = bencode.Unmarshal(file, &bto)
	if err != nil {
		return TorrentFile{}, err
	}
	return bto.toTorrentFile()
}

func (i *bencodeInfo) hash() ([20]byte, error) {
	var buf bytes.Buffer
	err := bencode.Marshal(&buf, *i)
	if err != nil {
		return [20]byte{}, err
	}
	h := sha1.Sum(buf.Bytes())
	return h, nil
}

func (i *bencodeInfo) splitPieceHashes() ([][20]byte, error) {
	hashLen := 20 // Length of SHA-1 hash
	buf := []byte(i.Pieces)
	if len(buf)%hashLen != 0 {
		err := fmt.Errorf("Received malformed pieces of length %d", len(buf))
		return nil, err
	}
	numHashes := len(buf) / hashLen
	hashes := make([][20]byte, numHashes)

	for i := 0; i < numHashes; i++ {
		copy(hashes[i][:], buf[i*hashLen:(i+1)*hashLen])
	}
	return hashes, nil
}

func (bto *bencodeTorrent) toTorrentFile() (TorrentFile, error) {
	infoHash, err := bto.Info.hash()
	if err != nil {
		return TorrentFile{}, err
	}
	pieceHashes, err := bto.Info.splitPieceHashes()
	if err != nil {
		return TorrentFile{}, err
	}

	t := TorrentFile{
		AnnounceList: []string{},
		InfoHash:     infoHash,
		PieceHashes:  pieceHashes,
		PieceLength:  bto.Info.PieceLength,
		Length:       bto.Info.Length,
		Name:         bto.Info.Name,
		MultipleFile: false,
	}

	for _, announceUrl := range bto.AnnounceList {
		if strings.Contains(announceUrl[0], ".local") {
			continue
		}
		t.AnnounceList = append(t.AnnounceList, announceUrl[0])
	}

	if len(t.AnnounceList) == 0 {
		t.AnnounceList = append(t.AnnounceList, bto.Announce)
	}

	if len(bto.Info.Files) == 0 {
		return t, nil
	}

	t.MultipleFile = false

	for _, fl := range bto.Info.Files {
		t.Length += fl.Length
	}

	return t, nil
}

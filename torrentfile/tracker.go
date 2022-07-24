package torrentfile

import (
	"crypto/tls"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/Kud1nov/torrent-client/peers"

	"github.com/jackpal/bencode-go"
)

type bencodeTrackerResp struct {
	Interval int    `bencode:"interval"`
	Peers    string `bencode:"peers"`
}

func (t *TorrentFile) buildTrackerURL(baseURL string, peerID [20]byte, port uint16) (string, error) {
	base, err := url.Parse(baseURL)
	if err != nil {
		return "", err
	}
	params := url.Values{
		"info_hash":  []string{string(t.InfoHash[:])},
		"peer_id":    []string{string(peerID[:])},
		"port":       []string{strconv.Itoa(int(Port))},
		"uploaded":   []string{"0"},
		"downloaded": []string{"0"},
		"left":       []string{strconv.Itoa(t.Length)},
		"compact":    []string{"1"},

		"numwant": []string{"200"},
	}
	base.RawQuery = params.Encode()

	return base.String(), nil
}

func (t *TorrentFile) requestPeers(baseURL string, peerID [20]byte, port uint16) ([]peers.Peer, error) {

	TrackerURL, err := t.buildTrackerURL(baseURL, peerID, port)
	if err != nil {
		return nil, err
	}
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	c := &http.Client{Transport: tr, Timeout: 15 * time.Second}
	resp, err := c.Get(TrackerURL)
	if err != nil {
		return nil, err
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	trackerResp := bencodeTrackerResp{}
	err = bencode.Unmarshal(resp.Body, &trackerResp)
	if err != nil {
		return nil, err
	}
	t.PeersInterval = trackerResp.Interval
	return peers.Unmarshal([]byte(trackerResp.Peers))
}

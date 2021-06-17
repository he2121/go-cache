package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/golang/protobuf/proto"

	"github.com/he2121/go-cache/consistent_hash"
	"github.com/he2121/go-cache/gocachepb"
)

const (
	defaultBasePath = "/go_cache/"
	defaultReplicas = 50
)

type HttpPool struct {
	selfPath    string // like 127.0.0.1:9999
	basePath    string // like /
	mu          sync.Mutex
	peers       *consistent_hash.Map
	httpGetters map[string]*httpGetter
}

func NewHttpPool(self string) *HttpPool {
	return &HttpPool{
		selfPath: self,
		basePath: defaultBasePath,
	}
}

func (p *HttpPool) Set(peers ...string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.peers = consistent_hash.New(defaultReplicas, nil)
	p.peers.Add(peers...)
	p.httpGetters = make(map[string]*httpGetter, len(peers))
	for _, peer := range peers {
		p.httpGetters[peer] = &httpGetter{baseURL: peer + defaultBasePath}
	}
}

func (p *HttpPool) PickPeer(key string) (PeerGetter, bool) {
	peer := p.peers.Get(key)
	if peer != "" && peer != p.selfPath {
		p.Log("Pick peer %s", peer)
		return p.httpGetters[peer], true
	}
	return nil, false
}

func (p *HttpPool) Log(format string, v ...interface{}) {
	log.Printf("[Server %s] %s", p.selfPath, fmt.Sprintf(format, v...))
}

func (p *HttpPool) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if !strings.HasPrefix(req.URL.Path, p.basePath) {
		panic("HTTPPool serving unexpected path: " + req.URL.Path)
	}
	p.Log("%s %s", req.Method, req.URL.Path)
	strs := strings.SplitN(req.URL.Path[len(p.basePath):], "/", 2)
	if len(strs) != 2 {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	groupName := strs[0]
	key := strs[1]
	group := GetGroup(groupName)
	if group == nil {
		http.Error(w, "no such group: "+groupName, http.StatusNotFound)
		return
	}
	byteView, err := group.Get(key)
	if err != nil {
		http.Error(w, "internal err", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	// ...
	// Write the value to the response body as a proto message.
	body, err := proto.Marshal(&gocachepb.Response{Value: byteView.ByteSlice()})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(body)
}

type httpGetter struct {
	baseURL string
}

func (h *httpGetter) Get(in *gocachepb.Request, out *gocachepb.Response) error {
	addr := fmt.Sprintf("%s%s/%s", h.baseURL, url.QueryEscape(in.Group), url.QueryEscape(in.Key))
	resp, err := http.Get(addr)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server return: %v", resp.Status)
	}
	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if err = proto.Unmarshal(bytes, out); err != nil {
		return fmt.Errorf("decoding response body: %v", err)
	}
	return nil
}

var _ PeerGetter = (*httpGetter)(nil)

var _ PeerPicker = (*HttpPool)(nil)

package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"
)

var defaultBasePath = "/go_cache/"

type HttpPool struct {
	selfPath string // like 127.0.0.1:9999
	basePath string	// like /
}

func NewHttpPool(self string) *HttpPool {
	return &HttpPool{
		selfPath: self,
		basePath: defaultBasePath,
	}
}

func (p *HttpPool) Log(format string, v ...interface{})  {
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
	w.Write(byteView.b)
}


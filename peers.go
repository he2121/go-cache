package main

import (
	"github.com/he2121/go-cache/gocachepb"
)

type PeerPicker interface {
	PickPeer(key string) (PeerGetter, bool)
}

type PeerGetter interface {
	Get(in *gocachepb.Request, out *gocachepb.Response) error
}

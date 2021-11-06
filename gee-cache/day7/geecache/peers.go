package geecache

import pb "Gee/gee-cache/day7/geecache/geecachepb"

// 根据key获取对应的节点PeerGetter

type PeerPicker interface {
	PeerPick(key string) (PeerGetter, bool)
}

// 找到Group对应group中key的value

type PeerGetter interface {
	Get(in *pb.Request, out *pb.Response) (error)
}

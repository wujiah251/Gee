package geecache

// 根据key获取对应的节点PeerGetter

type PeerPicker interface {
	PeerPick(key string) (PeerGetter, bool)
}

// 找到Group对应group中key的value

type PeerGetter interface {
	Get(group string, key string) ([]byte, error)
}

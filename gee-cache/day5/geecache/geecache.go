package geecache

import (
	"fmt"
	"log"
	"sync"
)

//是
//接收 key --> 检查是否被缓存 -----> 返回缓存值 ⑴
//|  否                         是
//|-----> 是否应当从远程节点获取 -----> 与远程节点交互 --> 返回缓存值 ⑵
//|  否
//|-----> 调用`回调函数`，获取值并添加到缓存 --> 返回缓存值 ⑶

// 负责与外部交互，控制缓存存储和获取的主流程

// 缓存不存在的回调函数，由用户来传递

// group is a cache namespace

type Group struct {
	name      string
	getter    Getter
	mainCache cache
	peers PeerPicker // 节点获取句柄
}

var (
	mu     sync.RWMutex
	groups = make(map[string]*Group)
)

type Getter interface {
	Get(key string) ([]byte, error)
}

type GetterFunc func(key string) ([]byte, error)

func (f GetterFunc) Get(key string) ([]byte, error) {
	return f(key)
}

func NewGroup(name string, cacheByes int64, getter Getter) *Group {
	if getter == nil {
		panic("nil Getter")
	}
	mu.Lock()
	defer mu.Unlock()
	g := &Group{
		name:      name,
		getter:    getter,
		mainCache: cache{cacheBytes: cacheByes},
	}
	groups[name] = g
	return g
}

func GetGroup(name string) *Group {
	mu.RLock()
	g := groups[name]
	mu.RUnlock()
	return g
}

func (g *Group) Get(key string) (ByteView, error) {
	if key == "" {
		return ByteView{}, fmt.Errorf("key is required")
	}

	if v, ok := g.mainCache.get(key); ok {
		log.Println("[GeeCache] hit")
		return v, nil
	}
	return g.load(key)
}

// 注册节点获取句柄
func (g*Group)RegisterPeers(peers PeerPicker){
	if g.peers != nil{
		panic("RegisterPeerPicker called more than once")
	}
	g.peers = peers
}


func (g *Group) load(key string) (value ByteView, err error) {
	if g.peers != nil{
		if peer, ok := g.peers.PeerPick(key);ok{
			if value, err = g.getFromPeer(peer,key);err ==nil{
				return value,nil
			}
			log.Println("[GeeCache]Failed to get from peer",err)
		}
	}
	return g.getLocally(key)
}

func (g *Group) getLocally(key string) (ByteView, error) {
	bytes, err := g.getter.Get(key)
	if err != nil {
		return ByteView{}, err
	}
	value := ByteView{b: cloneBytes(bytes)}
	g.populateCache(key, value)
	return value, nil
}

func (g *Group) populateCache(key string, value ByteView) {
	g.mainCache.add(key, value)
}

func (g *Group)getFromPeer(peer PeerGetter,key string)(ByteView,error){
	bytes, err := peer.Get(g.name,key)
	if err != nil{
		return ByteView{},err
	}
	return ByteView{b:bytes},nil
}

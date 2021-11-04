package geecache

import (
	"Gee/gee-cache/day6/geecache/singleflight"
	"fmt"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
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
	peers     PeerPicker // 节点获取句柄
	loader    *singleflight.Group
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
		loader:    &singleflight.Group{},
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

// 先读取mainCache
// 读取成功返回
// 读取失败则load
func (g *Group) Get(key string) (ByteView, error) {
	if key == "" {
		return ByteView{}, fmt.Errorf("key is required")
	}

	if v, ok := g.mainCache.get(key); ok {
		// 缓存命中，直接返回
		logrus.Infof("[%s mainCache get key]\n", g.name)
		return v, nil
	}
	return g.load(key)
}

// 注册节点获取句柄
func (g *Group) RegisterPeers(peers PeerPicker) {
	if g.peers != nil {
		panic("RegisterPeerPicker called more than once")
	}
	g.peers = peers
}

func (g *Group) load(key string) (value ByteView, err error) {
	viewi, err := g.loader.Do(key, func() (interface{}, error) {
		// 先尝试从远程获取key-value
		if g.peers != nil {
			if peer, ok := g.peers.PeerPick(key); ok {
				if value, err = g.getFromPeer(peer, key); err == nil {
					return value, nil
				}
				logrus.Infof("[GeeCache]Failed to get from peer,err=%v\n", err)
			}
		}
		// 本地读取key-value，实际上是调用getter
		return g.getLocally(key)
	})
	if err != nil {
		return viewi.(ByteView), nil
	}
	return
}

func (g *Group) getLocally(key string) (ByteView, error) {
	// 先sleep一下,模拟缓存失效
	time.Sleep(time.Second)
	bytes, err := g.getter.Get(key)
	if err != nil {
		return ByteView{}, err
	}
	value := ByteView{b: cloneBytes(bytes)}
	// 讲缓存添加到本地
	g.populateCache(key, value)
	return value, nil
}

func (g *Group) populateCache(key string, value ByteView) {
	g.mainCache.add(key, value)
}

func (g *Group) getFromPeer(peer PeerGetter, key string) (ByteView, error) {
	bytes, err := peer.Get(g.name, key)
	if err != nil {
		return ByteView{}, err
	}
	return ByteView{b: bytes}, nil
}

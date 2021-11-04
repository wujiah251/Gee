package geecache

import (
	"Gee/gee-cache/day6/geecache/consistenthash"

	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"

	"github.com/sirupsen/logrus"
)

const (
	defaultBasePath = "/_geecache/"
	defaultReplicas = 50
)

type HTTPPool struct {
	self     string
	basePath string // such as http://127.0.0.1:8080
	sync.Mutex
	peers       *consistenthash.Map // 一致性hash
	httpGetters map[string]*HTTPGetter
}

func NewHTTPPool(self string) *HTTPPool {
	return &HTTPPool{
		self:     self,
		basePath: defaultBasePath,
	}
}

// 向httpPool中注册节点

func (p *HTTPPool) Set(peers ...string) {
	p.Lock()
	defer p.Unlock()
	p.peers = consistenthash.New(defaultReplicas, nil)
	p.peers.Add(peers...)
	p.httpGetters = make(map[string]*HTTPGetter, len(peers))
	for _, peer := range peers {
		p.httpGetters[peer] = &HTTPGetter{baseURL: peer + p.basePath}
	}
}

// 获取key对应节点的客户端

func (p *HTTPPool) PeerPick(key string) (PeerGetter, bool) {
	p.Lock()
	defer p.Unlock()
	if peer := p.peers.Get(key); peer != "" && peer != p.self {
		logrus.Infof("[HTTPPool,name=%s]pick peer=%v\n", p.self, peer)
		logrus.Debugf("[HTTPPool,name=%s],status=%+v\n", p.self, *p)
		return p.httpGetters[peer], true
	}
	return nil, false
}

// 处理HTTP请求的接口
func (p *HTTPPool) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.URL.Path, p.basePath) {
		panic("HTTPPool serving unexpected path: " + r.URL.Path)
	}
	logrus.Infof("[HTTPPool(%s) recv req]Methpd=%v,url=%+v\n", p.self, r.Method, *(r.URL))
	parts := strings.SplitN(r.URL.Path[len(p.basePath):], "/", 2)
	if len(parts) != 2 {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	groupName := parts[0]
	key := parts[1]

	group := GetGroup(groupName)
	if group == nil {
		http.Error(w, "no such group: "+groupName, http.StatusNotFound)
		return
	}
	view, err := group.Get(key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(view.ByteSlice())
}

type HTTPGetter struct {
	baseURL string
}

func (g *HTTPGetter) Get(group string, key string) ([]byte, error) {
	url := fmt.Sprintf(
		"%s%s/%s",
		g.baseURL,
		group,
		key)
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned=%v", res.Status)
	}
	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body error=%v", err)
	}
	return bytes, nil
}

var _ PeerGetter = (*HTTPGetter)(nil)

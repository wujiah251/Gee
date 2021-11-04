package main

import (
	"Gee/gee-cache/day5/geecache"
	"flag"
	"fmt"
	"net/http"

	"github.com/sirupsen/logrus"
)

/*
> curl http://localhost:9999/_geecache/scores/Tom
630
> curl http://localhost:9999/_geecache/scores/kkk
kkk not exist
*/

var db = map[string]string{
	"Tom":  "630",
	"Jack": "589",
	"Sam":  "567",
}

func init() {
	logrus.SetLevel(logrus.InfoLevel)
}

func main() {
	var port int
	var api bool
	flag.IntVar(&port, "port", 8001, "Geecache server port")
	flag.BoolVar(&api, "api", false, "Start a api server?")
	flag.Parse()

	apiAddr := "http://localhost:9999"
	addrMap := map[int]string{
		8001: "http://localhost:8001",
		8002: "http://localhost:8002",
		8003: "http://localhost:8003",
	}
	var addrs []string
	for _, v := range addrMap {
		addrs = append(addrs, v)
	}

	gee := createGroup()
	if api {
		// 启动api服务器
		go startAPIServer(apiAddr, gee)
	}
	// 启动缓存服务器
	startCacheServer(addrMap[port], addrs, gee)
}

func getterFunc(key string) ([]byte, error) {
	logrus.Println("[SlowDB] search key", key)
	if v, ok := db[key]; ok {
		return []byte(v), nil
	}
	return nil, fmt.Errorf("%s not exist", key)
}

func createGroup() *geecache.Group {
	return geecache.NewGroup("score",
		2<<10,
		geecache.GetterFunc(getterFunc))
}

// 创建api服务器
func startAPIServer(apiAddr string, gee *geecache.Group) {
	http.Handle("/api", http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			key := r.URL.Query().Get("key")
			view, err := gee.Get(key)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/octet-stream")
			w.Write(view.ByteSlice())
		}))
	logrus.Println("fonted server is running at", apiAddr)
	logrus.Fatal(http.ListenAndServe(apiAddr[7:], nil))
	logrus.Infof("Start ApiServer,api addr = %v\n", apiAddr)
}

func startCacheServer(addr string, addrs []string, gee *geecache.Group) {
	peers := geecache.NewHTTPPool(addr)
	peers.Set(addrs...)      // 添加节点地址
	gee.RegisterPeers(peers) // 注册客户端
	logrus.Println("geecache is running at", addr)
	logrus.Fatal(http.ListenAndServe(addr[7:], peers))
	logrus.Infof("Start CacheServer, addr=%v\n", addr)
}

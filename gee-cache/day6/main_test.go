package main

import (
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"testing"
	"time"
)

const (
	TreadNums = 200
	Iters     = 10000
	BaseURL   = "http://localhost:9999/api?key="
)

var (
	KeyVec = []string{"Tom", "Jack", "Sam", "kkk"}
)

func RandReq() {
	index := rand.Int() % len(KeyVec)
	url := BaseURL + KeyVec[index]
	res, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	data, err := ioutil.ReadAll(res.Body)
	res.Body.Close() // 关闭连接
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("%s", data)
}

func TestPressure(t *testing.T) {
	start := time.Now()
	wg := &sync.WaitGroup{}
	wg.Add(1)
	for i := 0; i < TreadNums; i++ {
		go func(wg *sync.WaitGroup) {
			wg.Wait()
			RandReq()
		}(wg)
	}
	wg.Done()

	log.Printf("Time=%fs\n", time.Since(start).Seconds())

}

package singleflight

import "sync"

// 表示正在进行中，或已经结束的请求。
type call struct {
	sync.WaitGroup // 维护正在请求的个数，后来的key要等待前面的key释放
	val            interface{}
	err            error
}

// singleflight的主数据结构
type Group struct {
	sync.Mutex
	// 根据key来记录call，如果call结束，则删除key
	m map[string]*call
}

func (g *Group) Do(key string, fn func() (interface{}, error)) (interface{}, error) {
	g.Lock()
	if g.m == nil {
		g.m = make(map[string]*call)
	}
	if c, ok := g.m[key]; ok {
		g.Unlock()
		c.Wait() //等待所有当前正在请求的key结束
		return c.val, c.err
	}
	c := new(call)
	c.Add(1)
	g.m[key] = c
	g.Unlock()

	c.val, c.err = fn()
	c.Done()

	g.Lock()
	delete(g.m, key)
	g.Unlock()

	return c.val, c.err
}

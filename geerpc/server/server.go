package server

/*
客户端与服务端的通信需要协商一些内容，
例如 HTTP 报文，分为 header 和 body 2 部分，
body 的格式和长度通过 header 中的 Content-Type 和 Content-Length 指定，
服务端通过解析 header 就能够知道如何从 body 中读取需要的信息。
对于 RPC 协议来说，这部分协商是需要自主设计的。
为了提升性能，一般在报文的最开始会规划固定的字节，来协商相关的信息。
比如第1个字节用来表示序列化方式，第2个字节表示压缩方式，
第3-6字节表示 header 的长度，7-10 字节表示 body 的长度。
*/

import (
	"Gee/geerpc/codec"
	"Gee/geerpc/service"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"reflect"
	"strings"
	"sync"
)

const MagicNumber = 0x3bef5c

type Option struct {
	MagicNumber int
	CodecType   codec.Type
}

var DefaultOption = &Option{
	MagicNumber: MagicNumber,
	CodecType:   codec.GobType,
}

/*报文格式
| Option{MagicNumber: xxx, CodecType: xxx} | Header{ServiceMethod ...} | Body interface{} |
| <------      固定 JSON 编码      ------>  | <-------   编码方式由 CodeType 决定   ------->|
*/

// 服务端实现
type Server struct {
	// 一个server可以集成多个service
	serviceMap sync.Map
}

type request struct {
	h            *codec.Header
	argv, replyv reflect.Value
	mType        *service.MethodType
	svc          *service.Service
}

func NewServer() *Server {
	return &Server{}
}

var DefaultServer = NewServer()

// Service相关方法
// 注册一个service
func (server *Server) Register(rcvr interface{}) error {
	s := service.NewService(rcvr)
	if _, dup := server.serviceMap.LoadOrStore(s.Name, s); dup {
		return errors.New("rpc: service already defined: " + s.Name)
	}
	return nil
}

func Register(rcvr interface{}) error {
	return DefaultServer.Register(rcvr)
}

func (server *Server) findService(serviceMethod string) (
	svc *service.Service, mthType *service.MethodType, err error) {
	dot := strings.LastIndex(serviceMethod, ".")
	if dot == -1 {
		err = errors.New("rpc server: service/method request ill-formed: " + serviceMethod)
		return
	}
	serviceName, methodName := serviceMethod[:dot], serviceMethod[dot+1:]
	svci, ok := server.serviceMap.Load(serviceName)
	if !ok {
		err = fmt.Errorf("rpc server: service[%v] is not found", serviceName)
		return
	}
	svc = svci.(*service.Service)
	if mthType, ok = svc.Method[methodName]; !ok {
		err = fmt.Errorf("rpc server: method[%v] is not found", methodName)
		return
	}
	return
}

func (server *Server) Accept(lis net.Listener) {
	for {
		conn, err := lis.Accept()
		if err != nil {
			log.Println("rpc server: accept error:", err)
		}
		go server.ServeConn(conn)
	}
}

func Accept(lis net.Listener) { DefaultServer.Accept(lis) }

// 处理连接
func (server *Server) ServeConn(conn io.ReadWriteCloser) {
	defer func() { _ = conn.Close() }()
	var opt Option

	// 先解析opt
	if err := json.NewDecoder(conn).Decode(&opt); err != nil {
		log.Println("rpc server: options error: ", err)
	}
	if opt.MagicNumber != MagicNumber {
		log.Printf("rpc server: invalid magic number %x", opt.MagicNumber)
		return
	}

	// 获取编解码器
	f := codec.NewCodecFuncMap[opt.CodecType]
	if f == nil {
		log.Printf("rpc server: invalid codec type %s", opt.CodecType)
		return
	}

	server.serveCodec(f(conn))
}

var invalidRequest = struct{}{}

func (server *Server) serveCodec(cc codec.Codec) {
	sending := new(sync.Mutex) // 用于保证发送完整的response
	wg := new(sync.WaitGroup)  // 等待所有请求都处理完
	for {
		req, err := server.readRequest(cc)
		if err != nil {
			if req != nil {
				break
			}
			req.h.Error = err.Error()
			server.sendResponse(cc, req.h, invalidRequest, sending)
		}
		wg.Add(1)
		go server.handleRequest(cc, req, sending, wg)
	}
	wg.Wait()
	_ = cc.Close()
}

func (server *Server) readRequest(cc codec.Codec) (*request, error) {
	h, err := server.readRequestHeader(cc)
	if err != nil {
		return nil, err
	}
	req := &request{h: h}
	req.svc, req.mType, err = server.findService(h.ServiceMethod)
	if err != nil {
		return req, err
	}
	req.argv = req.mType.NewArgv()
	req.replyv = req.mType.NewReplyv()

	argvi := req.argv.Interface()
	if req.argv.Type().Kind() != reflect.Ptr {
		argvi = req.argv.Addr().Interface()
	}

	// 读取参数
	if err = cc.ReadBody(argvi); err != nil {
		log.Println("rpc server: read body err:", err)
		return req, err
	}
	return req, nil
}

func (server *Server) readRequestHeader(cc codec.Codec) (*codec.Header, error) {
	var h codec.Header
	if err := cc.ReadHeader(&h); err != nil {
		if err != io.EOF && err != io.ErrUnexpectedEOF {
			log.Println("rpc server: read header error:", err)
		}
		return nil, err
	}
	return &h, nil
}

func (server *Server) handleRequest(cc codec.Codec,
	req *request, sending *sync.Mutex, wg *sync.WaitGroup) {

	defer wg.Done()
	log.Println(req.h, req.argv.Elem())

	if err := req.svc.Call(req.mType, req.argv, req.replyv); err != nil {
		req.h.Error = err.Error()
		server.sendResponse(cc, req.h, req.replyv.Interface(), sending)
		return
	}

	server.sendResponse(cc, req.h, req.replyv.Interface(), sending)
}
func (server *Server) sendResponse(cc codec.Codec, h *codec.Header,
	body interface{}, sending *sync.Mutex) {
	sending.Lock()
	defer sending.Unlock()
	log.Printf("body=%+v\n", body)
	if err := cc.Write(h, body); err != nil {
		log.Println("rpc server: write response error:", err)
	}
}

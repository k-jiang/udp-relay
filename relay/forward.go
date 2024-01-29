package main

import (
	"net"
	"sync"
	"time"
)

type Forwarder struct {
	src         Address
	dst         Address
	pool        *ConnectionPool
	connTimeout int
	maxConn     int
}

type ConnectionPool struct {
	sync.Mutex
	connections map[string]net.Conn
}

func NewForwarder(src Address, dst Address, connTimeout, maxConn int) *Forwarder {
	return &Forwarder{
		src:         src,
		dst:         dst,
		connTimeout: connTimeout,
		maxConn:     maxConn,
	}
}

func (f *Forwarder) Start() {
	// Prepare connection pool for upstream
	f.pool = &ConnectionPool{
		connections: make(map[string]net.Conn),
	}
	// Listen for client (src) connection
	lAddr, err := net.ResolveUDPAddr("udp", f.src.String())
	panicIfErr(err)
	lConn, err := net.ListenUDP("udp", lAddr)
	panicIfErr(err)
	for {
		// Accept client (src) connection
		var buf [65507]byte
		n, srcAddr, err := lConn.ReadFromUDP(buf[0:])
		if err != nil {
			log.Println(`[E]`, err.Error())
			continue
		}
		go func(buf [65507]byte, n int, srcAddr *net.UDPAddr) {
			// Connect to upstream server, if not int connection poll
			f.pool.Lock()
			dstConn, found := f.pool.connections[srcAddr.String()]
			if !found {
				log.Printf(`[I] Connect from client udp://%s <-> upstream udp://%s`+"\n", srcAddr.String(), f.dst.String())
				dstAddr, err := net.ResolveUDPAddr("udp", f.dst.String())
				if err != nil {
					log.Println(`[E]`, err.Error())
					f.pool.Unlock()
					return
				}
				// respect max connections
				if len(f.pool.connections) > f.maxConn {
					log.Printf(`[E] Maximun upstream connections reached. Dropping new client connection udp://%s`, srcAddr.String())
					f.pool.Unlock()
					return
				}
				dstConn, err = net.DialUDP("udp", nil, dstAddr)
				if err != nil {
					log.Println(`[E]`, err.Error())
					f.pool.Unlock()
					return
				}
				f.pool.connections[srcAddr.String()] = dstConn

				// Relay buf from upstream -> client
				go func(srcAddr *net.UDPAddr) {
					var buf [65507]byte
					for {
						dstConn.SetReadDeadline(time.Now().Add(time.Duration(f.connTimeout) * time.Second))
						n, err := dstConn.Read(buf[0:])
						if err != nil {
							log.Printf(`[I] Disconnect from client udp://%s <-> upstream udp://%s`+"\n", srcAddr.String(), dstConn.RemoteAddr().String())
							f.pool.Lock()
							delete(f.pool.connections, srcAddr.String())
							f.pool.Unlock()
							dstConn.Close()
							return
						}
						if n > 0 {
							lConn.WriteToUDP(buf[:n], srcAddr)
						}
					}
				}(srcAddr)
			}
			f.pool.Unlock()

			// Relay buf from client -> upstream
			dstConn.Write(buf[:n])
		}(buf, n, srcAddr)
	}
}

func panicIfErr(err error) {
	if err != nil {
		panic(err)
	}
}

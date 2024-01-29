package main

import (
	"flag"
	"fmt"
	_log "log"
	"os"
	"strconv"
	"strings"
)

var log = _log.New(os.Stdout, "", _log.LstdFlags)

type Address struct {
	Host string
	Port int
}

func (a *Address) String() string {
	return fmt.Sprintf(`%s:%d`, a.Host, a.Port)
}

func main() {
	listenHost := flag.String("lHost", "0.0.0.0", "listen host")
	listenPort := flag.String("lPort", "8080", "listen port ( can be a port range like 1024-5555 too )")
	remoteHost := flag.String("rHost", "8.8.8.8", "remote host")
	remotePort := flag.Int("rPort", 53, "remote port ( will be equal to lPort when doing range forward )")
	connTimeout := flag.Int("timeout", 60, "connection timeout in seconds")
	maxConn := flag.Int("maxConn", 1000, "maximum connections")
	help := flag.Bool("help", false, "print help")
	h := flag.Bool("h", false, "")

	flag.Parse()
	if *help || *h {
		flag.PrintDefaults()
		return
	}

	if *connTimeout < 0 {
		panic("invalid connection timeout, it should be bigger or equal to 0")
	}
	src := Address{}
	dst := Address{
		Host: *remoteHost,
		Port: *remotePort,
	}

	if strings.Contains(*listenPort, "-") {
		pRange := strings.Split(*listenPort, "-")
		if len(pRange) > 2 {
			log.Println("[E] invalid port range format, should be like 1024-5555")
			os.Exit(1)
		}
		start, err := strconv.Atoi(pRange[0])
		if err != nil || start < 1 || start > 65534 {
			log.Println(`[E] invalid port range start, should be a Number between 1 and 65534`)
			os.Exit(1)
		}
		end, err := strconv.Atoi(pRange[1])
		if err != nil || end <= start || end > 65534 {
			log.Println(`[E] invalid port range start, should be a Number between 1 and 65534 ( start should be smaller than end too )`)
			os.Exit(1)
		}
		for port := start; port <= end; port++ {
			log.Printf(`[I] src udp://%s:%d <-> dst udp://%s:%d`+"\n", *listenHost, port, *remoteHost, port)
			go startForwarder(Address{
				Host: *listenHost,
				Port: port,
			}, Address{
				Host: *remoteHost,
				Port: port,
			}, *connTimeout, *maxConn)
		}
	} else {
		n, err := strconv.Atoi(*listenPort)
		if err != nil {
			log.Println("[E] invalid listenPort, should be a Number between 1 and 65534")
			os.Exit(1)
		}
		src.Host = *listenHost
		src.Port = n
		log.Printf(`[I] Start on src udp://%s <-> dst udp://%s`+"\n", src.String(), dst.String())
		go startForwarder(src, dst, *connTimeout, *maxConn)
	}
	select {}
}

func startForwarder(src, dst Address, connTimeout, maxConn int) {
	forwarder := NewForwarder(src, dst, connTimeout, maxConn)
	forwarder.Start()
}

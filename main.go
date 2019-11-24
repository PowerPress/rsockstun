package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/hashicorp/yamux"
	"time"

	"strconv"
	"strings"

)

var session *yamux.Session
var agentpassword string
var sconn net.Conn

func main() {

	listen := flag.String("listen", "", "listen port for receiver address:port")
	certificate := flag.String("cert", "", "certificate file")
	socks := flag.String("socks", "127.0.0.1:1080", "socks address:port")
	connect := flag.String("connect", "", "connect address:port")
	proxy := flag.String("proxy", "", "proxy address:port")
	optproxytimeout := flag.String("proxytimeout", "", "proxy response timeout (ms)")
	proxyauthstring := flag.String("proxyauth", "", "proxy auth Domain/user:Password ")
	optuseragent := flag.String("useragent", "", "User-Agent")
	optpassword := flag.String("pass", "", "Connect password")
	optredirecturl := flag.String("rurl", "", "redirect url. Ex: http://mail.com/login")
	frontDomain := flag.String("frontDomain", "www.google.com", "Fake domain for eSNI fronting")
	recn := flag.Int("recn", 3, "reconnection limit")
	//ymx := flag.Bool("ymx", true, "use yamux")

	rect := flag.Int("rect", 30, "reconnection delay")
	version := flag.Bool("version", false, "version information")
	flag.Usage = func() {
		fmt.Println("rsockstun - reverse socks5 server/client with eSNI domain fronting support")
		fmt.Println("")
		fmt.Println("Usage:")
		fmt.Println("0) Generate self-signed SSL certificate: openssl: openssl req -new -x509 -keyout server.key -out server.crt -days 365 -nodes")
		fmt.Println("1) Start rsockstun -listen :8080 -socks 127.0.0.1:1080 -cert server on the server.")
		fmt.Println("1a) Start listening with websocket: rsockstun -listen wss:0.0.0.0 -socks 127.0.0.1:1080 -cert server -frontDomain www.apple.com on the server.")
		fmt.Println("2) Start rsockstun -connect client:8080 on the client inside LAN.")
		fmt.Println("2a) Create cloudflare account and set your domain nameservers to it")
		fmt.Println("2b) Start connecting via websocket: rsockstun -connect wss:client on the client inside LAN.")
		fmt.Println("3) Connect to 127.0.0.1:1080 on the server with any socks5 client to access into LAN.")
		fmt.Println("X) Enjoy. :]")
	}

	flag.Parse()


	if *version {
		fmt.Println("rsockstun - reverse socks5 server/client")
		os.Exit(0)
	}

	if *listen != "" {
		log.Println("Starting to listen for clients")

		if *optproxytimeout != "" {
			opttimeout,_ := strconv.Atoi(*optproxytimeout)
			proxytout = time.Millisecond * time.Duration(opttimeout)
		} else {
			proxytout = time.Millisecond * 1000
		}

		if *optredirecturl != "" {
			rurl = *optredirecturl
		} else {
			rurl = "https://www.microsoft.com/"
		}

		if *optpassword != "" {
			agentpassword = *optpassword
		} else {
			agentpassword = "RocksDefaultRequestRocksDefaultRequestRocksDefaultRequestRocks!!"
		}

		if (strings.Contains(*listen,"ws:") || strings.Contains(*listen,"wss:")){
			go listenForWsClients(*listen, *certificate)
			log.Fatal(listenForSocks(*socks ))

		}else{
			//do not use websocket
			go listenForClients(*listen, *certificate)
			log.Fatal(listenForSocks(*socks ))
		}
	}

	if *connect != "" {

		if *optproxytimeout != "" {
			opttimeout,_ := strconv.Atoi(*optproxytimeout)
			proxytimeout = time.Millisecond * time.Duration(opttimeout)
		} else {
			proxytimeout = time.Millisecond * 1000
		}

		if *proxyauthstring != "" {
			domain = strings.Split(*proxyauthstring, "/")[0]
			username = strings.Split(strings.Split(*proxyauthstring, "/")[1],":")[0]
			password = strings.Split(strings.Split(*proxyauthstring, "/")[1],":")[1]
		} else {
			domain = ""
			username = ""
			password = ""
		}

		if *optpassword != "" {
			agentpassword = *optpassword
		} else {
			agentpassword = "RocksDefaultRequestRocksDefaultRequestRocksDefaultRequestRocks!!"
		}

		if *optuseragent != "" {
			useragent = *optuseragent
		} else {
			useragent = "Mozilla/5.0 (Windows NT 6.1; Trident/7.0; rv:11.0) like Gecko"
		}
		//log.Fatal(connectForSocks(*connect,*proxy))

		//try to connect to server for recn times
		for i := 1; i <= *recn; i++ {
			log.Printf("Connecting to the far end. Try %d of %d",i,*recn)
			var error1 error
			if strings.Contains(*connect,"ws:") || strings.Contains(*connect,"wss:") {
				error1 = connectForWsSocks(*connect, *proxy, *frontDomain)
			}else {
				error1 = connectForSocks(*connect, *proxy)
			}

			log.Print(error1)
			log.Printf("Sleeping for %d sec...",*rect)
			tsleep := time.Second * time.Duration(*rect)
			time.Sleep(tsleep)
		}


		log.Fatal("Ending...")
	}

	fmt.Fprintf(os.Stderr, "You must specify a listen port or a connect address")
	os.Exit(1)
}

package main

import (
	"flag"
	"os"

	"net/http"
	"strings"

	"log"

	"github.com/elazarl/goproxy"
	"github.com/miekg/dns"
)

const (
	dnsDefaultHost    = "127.0.0.1"
	port              = "53"
	redirectionOrigin = "neverssl.com"
	redirectionTarget = "httpforever.com"
	proxyAddr         = "127.0.0.1:53120"
)

var (
	dnsLoggerInfo    *log.Logger
	dnsLoggerFatal   *log.Logger
	proxyLoggerInfo  *log.Logger
	proxyLoggerFatal *log.Logger
)

// HINT: Creation du looger
func init() { // Utiliser zeroLog
	dnsLoggerInfo = log.New(os.Stdout, "[DNS - Info] ", log.LstdFlags)
	dnsLoggerFatal = log.New(os.Stdout, "[DNS - Error] ", log.LstdFlags)
	proxyLoggerInfo = log.New(os.Stdout, "[PROXY - Info] ", log.LstdFlags)
	proxyLoggerFatal = log.New(os.Stdout, "[PROXY - Error] ", log.LstdFlags)
}

func main() {
	// Set & Parse Flags with default value
	hostFlagPtr := flag.String("h", dnsDefaultHost, "flag to define the host of the server")
	portFlagPtr := flag.String("p", port, "flag to define the port of the service")
	updateFlagPtr := flag.Bool("u", false, "flag to force the fetch of adservers.txt")
	flag.Parse()

	// Set an adress for the dns server.
	address := *hostFlagPtr + ":" + *portFlagPtr

	// Initialize the mux (Rooting sytem for incoming http adress. Stands on top of the server). It will be in charge of the requests dispatch.
	// Attach the blackList
	mux := dns.NewServeMux()
	// Update the adservers.txt list
	if *updateFlagPtr {
		fetchList()
	}
	// Attacher les handlers
	setBlackList(mux)
	mux.HandleFunc(redirectionOrigin, redirectRequest)
	mux.HandleFunc(".", forwardRequest)

	// Initialize the dns server.
	dnsServer := dns.Server{Addr: address, Net: "udp", Handler: mux, NotifyStartedFunc: func() {
		dnsLoggerInfo.Println("The Dns Server is listenning at", address)
	}}

	// Start the proxy in a go-routine. It will achieve the redirection by changing the header of the request.
	proxy := goproxy.NewProxyHttpServer()
	proxy.OnRequest().DoFunc(func(r *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
		if strings.HasSuffix(r.Host, redirectionOrigin) {
			r.Host = redirectionTarget
			proxyLoggerInfo.Println("Header request modified : \n", r)
		}
		return r, nil
	})

	go func() {
		proxyLoggerInfo.Println("The Proxy Server is listenning at", proxyAddr)
		if err := http.ListenAndServe(proxyAddr, proxy); err != nil {
			proxyLoggerFatal.Fatal(err)
		}
	}()

	// Start the dns server.
	if err := dnsServer.ListenAndServe(); err != nil {
		dnsLoggerFatal.Fatal(err)
	}
}

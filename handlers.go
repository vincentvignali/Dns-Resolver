package main

import (
	"bufio"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/miekg/dns"
)

const (
	url            = "https://hosts.anudeep.me/mirror/adservers.txt"
	filepath       = "./adservers.txt"
	redirectionIp6 = "2a00:1450:4007:80b::2003"
	dnsToForward   = "8.8.8.8:53"
)

var redirectionIp4 = net.IP{172, 67, 181, 181}

// HINT: Retrieve the blackList from : "https://hosts.anudeep.me/mirror/adservers.txt"
func fetchList() {
	dnsLoggerInfo.Printf("Fetch the list from %v", url)
	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		log.Fatal(err)
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	// Check server response
	if resp.StatusCode != http.StatusOK {
		log.Fatal(resp.Status)
	}

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	dnsLoggerInfo.Println("Fetch Done")
}

func setBlackList(mux *dns.ServeMux) {
	// Loop over all lines, populate the list Array with domains only
	f, _ := os.Open(filepath)
	scanner := bufio.NewScanner(f)
	list := []string{}
	for scanner.Scan() {
		line := strings.Join(strings.Split(scanner.Text(), "0.0.0.0 "), "")
		list = append(list, line)
	}
	// remove the 10 first Line: irrelevant information from file
	list = list[10:]
	// Attach the HandleFunction for each domain
	for i := 0; i < len(list); i++ {
		mux.HandleFunc(list[i], blockRequest)
	}
}

func blockRequest(w dns.ResponseWriter, r *dns.Msg) {
	domain := r.Question[0].Name
	dnsLoggerInfo.Println("BLOCKED REQUEST :", domain)
	m := dns.Msg{}
	m.SetReply(r)
	// Code request for "REFUSED"
	m.MsgHdr.Rcode = 5
	// Send back the answer to the client
	if err := w.WriteMsg(&m); err != nil {
		dnsLoggerFatal.Fatalf("ERROR : Answer could not be sent back to the client :\n => %v\n", err)
	}
}

func redirectRequest(w dns.ResponseWriter, r *dns.Msg) {
	// Build the answer
	domain := r.Question[0].Name
	dnsLoggerInfo.Println("INTERCEPTED REQUEST : \n", domain)
	m := dns.Msg{}
	m.SetReply(r)
	m.Answer = []dns.RR{}
	for index, question := range r.Question {
		switch questionType := question.Qtype; questionType {
		case dns.TypeA:
			m.Answer[index] = &dns.A{A: redirectionIp4, Hdr: dns.RR_Header{Name: domain, Rrtype: questionType, Class: dns.ClassINET, Ttl: 3600}}
		case dns.TypeAAAA:
			m.Answer[index] = &dns.AAAA{AAAA: net.ParseIP(redirectionIp6), Hdr: dns.RR_Header{Name: domain, Rrtype: questionType, Class: 1, Ttl: 3600}}
		default:
			m.Answer[index] = &dns.A{A: net.IP{172, 67, 181, 181}, Hdr: dns.RR_Header{Name: domain, Rrtype: 1, Class: 1, Ttl: 5}}
		}
		dnsLoggerInfo.Println("CUSTOM ANSWER : \n", m.Answer[index])
	}
	// Send back the answer to the client
	if err := w.WriteMsg(&m); err != nil {
		dnsLoggerFatal.Fatalf("ERROR : Answer could not be sent back to the client :\n : %v\n", err)
	}
}

func forwardRequest(w dns.ResponseWriter, r *dns.Msg) {
	for _, question := range r.Question {
		domain := question.Name
		dnsLoggerInfo.Println("FORWARDED REQUEST : ", domain)
		// Forward to google dns service
		if response, err := dns.Exchange(r, dnsToForward); err != nil {
			dnsLoggerFatal.Fatalf("ERROR : Query could not be forwarded to the Google dns resolver :\n %v\n", err)
		} else {
			if err := w.WriteMsg(response); err != nil {
				dnsLoggerFatal.Fatalf("ERROR : Query could not be forwarded to the Google dns resolver :\n %v\n", err)
			}
		}
	}
}

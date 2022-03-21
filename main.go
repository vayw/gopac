package main

import (
	"bytes"
	"embed"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"text/template"
)

var version = "undefined"
var debug bool

var proxy_scheme, proxy_address, proxy_port string

var PACfile string
var DOMAINS []string

//go:embed templates/proxy.pac.tmpl
var f embed.FS

func main() {
	flag.BoolVar(&debug, "debug", false, "debug mode")
	port := flag.Int("port", 6033, "http port")
	addr := flag.String("address", "0.0.0.0", "ip address")
	config := flag.String("domains-file", "gopac.conf.json", "json config file")
	flag.StringVar(&proxy_scheme, "proxy-scheme", "SOCKS5", "proxy scheme")
	flag.StringVar(&proxy_address, "proxy-address", "0.0.0.0", "proxy address")
	flag.StringVar(&proxy_port, "proxy-port", "8080", "proxy port")
	flag.Parse()

	loadDomains(*config)
	preparePAC()

	httpport := fmt.Sprintf("%s:%d", *addr, *port)
	http.HandleFunc("/proxy.pac", getPAC)
	http.HandleFunc("/version", showVersion)

	http.ListenAndServe(httpport, nil)
}

func preparePAC() {
	templ := template.Must(template.New("").ParseFS(f, "templates/*.tmpl"))

	var tpl bytes.Buffer
	data := struct {
		Address string
		Port    string
		Scheme  string
		Domains []string
	}{proxy_address, proxy_port, proxy_scheme, DOMAINS}

	if err := templ.ExecuteTemplate(&tpl, "proxy.pac.tmpl", data); err != nil {
		log.Panic(err)
	}
	PACfile = tpl.String()
}

func getPAC(w http.ResponseWriter, r *http.Request) {
	log.Println("http:", r.Method, r.URL, r.RemoteAddr)
	w.Header().Set("ContentType", "application/x-ns-proxy-autoconfig")
	fmt.Fprint(w, PACfile)
}

func showVersion(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, version)
}

type DomainsList struct {
	Domains []string `json:"domains"`
}

func loadDomains(fname string) {
	// read file
	data, err := ioutil.ReadFile(fname)
	if err != nil {
		log.Panic(err)
	}

	var dlist DomainsList
	err = json.Unmarshal(data, &dlist)
	if err != nil {
		log.Panic("error:", err)
	}
	DOMAINS = dlist.Domains
}

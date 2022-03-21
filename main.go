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

var PACfile string
var Conf Config

//go:embed templates/proxy.pac.tmpl
var f embed.FS

func main() {
	flag.BoolVar(&debug, "debug", false, "debug mode")
	port := flag.Int("port", 6033, "http port")
	addr := flag.String("address", "0.0.0.0", "ip address")
	config := flag.String("config-file", "gopac.conf.json", "json config file")
	flag.Parse()

	loadConfig(*config)
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
		C Config
	}{Conf}

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

type Config struct {
	Domains []string `json:"domains"`
	Proxies []struct {
		Address string `json:"address"`
		Port    string `json:"port"`
		Scheme  string `json:"scheme"`
	} `json:"proxies"`
}

func loadConfig(fname string) {
	// read file
	data, err := ioutil.ReadFile(fname)
	if err != nil {
		log.Panic(err)
	}

	err = json.Unmarshal(data, &Conf)
	if err != nil {
		log.Panic("error:", err)
	}
}

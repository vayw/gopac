package main

import (
	"embed"
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

var version = "undefined"
var debug bool

var proxy_scheme, proxy_address, proxy_port string

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

	httpport := fmt.Sprintf("%s:%d", *addr, *port)
	router := setupRouter()

	router.Run(httpport)
}

func setupRouter() *gin.Engine {
	if debug == false {
		gin.SetMode(gin.ReleaseMode)
	}
	router := gin.Default()
	templ := template.Must(template.New("").ParseFS(f, "templates/*.tmpl"))
	router.SetHTMLTemplate(templ)

	router.GET("/proxy.pac", getPAC)
	router.GET("/version", showVersion)

	return router
}

func getPAC(c *gin.Context) {
	fmt.Println(proxy_scheme, proxy_port)
	c.HTML(http.StatusOK, "proxy.pac.tmpl", gin.H{
		"address": proxy_address,
		"port":    proxy_port,
		"scheme":  proxy_scheme,
		"domains": DOMAINS,
	})
}

func showVersion(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"version": version})
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

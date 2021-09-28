package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

var user string
var pass string

func main() {
	host := flag.String("h", "127.0.0.1", "interface to serve on")
	port := flag.String("p", "8080", "port to serve on")
	userpass := flag.String("u", "", "user:pass for basic auth")
	flag.Usage = usage
	flag.Parse()

	user = strings.SplitN(*userpass, ":", 2)[0]
	pass = strings.SplitN(*userpass, ":", 2)[1]
	http.HandleFunc("/exec", requireAuth(webExec))
	log.Printf("[!] This server provides code execution as a service")
	log.Printf("Server started listening on http://%s:%s", *host, *port)
	log.Fatal(http.ListenAndServe(*host+":"+*port, nil))
}
func requireAuth(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logRequest(w, r)
		user, pass, _ := r.BasicAuth()
		if !checkCreds(user, pass) {
			w.Header().Set("WWW-Authenticate", "Basic realm=\"Can I help you?\"")
			http.Error(w, "Unauthorized", 401)
			return
		}
		fn(w, r)
	}
}
func checkCreds(u, p string) bool {
	if u == user && p == pass {
		log.Printf("Successful auth for user %s", u)
		return true
	}
	log.Printf("Got auth request for user %s", u)
	return false
}

func webExec(w http.ResponseWriter, r *http.Request) {

	if err := r.ParseForm(); err != nil {
		log.Print(err)
	}
	cmd := ""
	for k, v := range r.Form {
		if k == "cmd" {
			cmd = v[0]
		}
	}

	lsCmd := exec.Command("bash", "-c", cmd)
	lsOut, err := lsCmd.Output()
	if err != nil {
		log.Printf("%s\n", err)
		http.Error(w, "Server error", 503)
	}
	fmt.Fprintf(w, string(lsOut))
}

func usage() {
	fmt.Printf("Usage: ./%s\n", os.Args[0])
	flag.PrintDefaults()
	os.Exit(1)
}

func logRequest(w http.ResponseWriter, r *http.Request) {
	log.Printf("%s, %s, User-Agent = %s\n", r.Method, r.RemoteAddr, r.Header.Get("User-Agent"))
	log.Printf("\t%s\n", r.RequestURI)
	if err := r.ParseForm(); err != nil {
		log.Print(err)
	}
	for k, v := range r.Form {
		log.Printf("\tForm[%q] = %q\n", k, v)
	}
}

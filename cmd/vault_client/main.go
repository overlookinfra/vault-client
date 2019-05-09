package main

import (
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/puppetlabs/vault-client/pkg/client"
)

func main() {
	log.Print("Grabbing shared cert from vault")
	cert, err := client.GetCert()
	if err != nil {
		panic(err)
	}

	if err := os.MkdirAll("/etc/ssl/certs/puppet-discovery", os.ModeDir|os.ModePerm); err != nil {
		panic(err)
	}

	log.Print("Writing cert files to /etc/ssl/certs/puppet-discovery")
	if err := ioutil.WriteFile(
		"/etc/ssl/certs/puppet-discovery/shared.ca",
		[]byte(cert.CA),
		os.ModePerm); err != nil {
		panic(err)
	}

	if err := ioutil.WriteFile(
		"/etc/ssl/certs/puppet-discovery/shared.crt",
		[]byte(cert.Cert),
		os.ModePerm); err != nil {
		panic(err)
	}

	if err := ioutil.WriteFile(
		"/etc/ssl/certs/puppet-discovery/shared.key",
		[]byte(cert.PrivateKey),
		os.ModePerm); err != nil {
		panic(err)
	}

	flag.Parse()
	args := flag.Args()

	if len(args) > 0 && args[0] == "serve" {
		http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
			w.Header().Set("Content-Type", "text/plain")
			w.Write([]byte("This is an example served up with a vault-backed shared cert which was written to /etc/ssl/certs/puppet-discovery.\n"))
		})
		log.Print("Serving an example app at https://192.168.33.11:9443...")
		if err := http.ListenAndServeTLS(
			":9443",
			"/etc/ssl/certs/puppet-discovery/shared.crt",
			"/etc/ssl/certs/puppet-discovery/shared.key",
			nil); err != nil {
			panic(err)
		}
	}
}

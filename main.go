package main

import (
	_ "embed"
	"errors"
	"flag"
	"log"
	"os"
	"strings"

	"github.com/tube-sh/tubed/daemon"
	"github.com/tube-sh/tubed/tunnel"
)

//go:embed assets/frpc
var frpc []byte

// init variables
var tubedDir string = "/etc/tubed/"

var signal = flag.String("s", "", `Send signal to the daemon:
  quit — graceful shutdown
  stop — fast shutdown
  reload — reloading the configuration file`)

var proxy string
var token string
var baseUrl string

func main() {

	// get flags
	flag.StringVar(&proxy, "proxy", "", "proxy configuration")
	flag.StringVar(&baseUrl, "apiurl", "https://api.tube.sh", "set tube API URL (default is https://api.tube.sh")
	skipTLSVerify := flag.Bool("skiptlsverify", false, "disable TLS certificate check on API request")

	flag.Parse()

	// set proxy if requested
	if proxy != "" {
		os.Setenv("http_proxy", proxy)
	}

	if *skipTLSVerify == true {
		os.Setenv("skiptlsverify", "true")
	}

	// main if no stop signal
	if *signal != "stop" {

		// get tunnel token from file
		log.Println("get tubed tunnel token from " + tubedDir + "token")
		f, err := os.ReadFile(tubedDir + "token")
		if err != nil {
			log.Fatal(err)
		}
		token = strings.TrimSuffix(string(f), "\n")

		// if no tubed.bootstrapped file, do bootstrap
		if _, err := os.Stat(tubedDir + "tubed.bootstrapped"); errors.Is(err, os.ErrNotExist) {
			log.Println("start tubed bootstrap...")

			// bootstrap tunnel
			tunnel.Bootstrap(tubedDir, baseUrl, token, frpc)

			// create empty file to tag as bootstrapped
			log.Println("create tubed.bootstrapped file to " + tubedDir + "tubed.bootstrapped")
			_, err = os.Create(tubedDir + "tubed.bootstrapped")
			if err != nil {
				log.Fatal(err)
			}

			log.Println("tubed bootstrap finished successfully")
		}

		log.Println("- - - - - - - - - - - - - - -")
		log.Println("starting tubed daemon...")
		log.Println("logfile location is /var/log/tubed.log")
	}

	// run daemon
	daemon.Run(proxy, tubedDir, baseUrl, token, signal)

	if *signal == "stop" {
		log.Println("tubed daemon terminated")
	}

}

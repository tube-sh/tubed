package daemon

import (
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/tube-sh/tubed/tunnel"
)

// init variables
var (
	stop = make(chan struct{})
	done = make(chan struct{})
)
var pid int

func Worker(proxy string, tubedDir string, baseUrl string, token string) {

	// set http_proxy env var if requested
	if proxy != "" {
		log.Println("set http_proxy env var")
		os.Setenv("http_proxy", proxy)
	}

	// run frpc service (tunnel mount)
	log.Println("starting frpc subprocess")
	cmd := exec.Command("/etc/tubed/frpc", "-c", "/etc/tubed/frpc.ini")
	cmd.Stdout = os.Stdout
	err := cmd.Start()
	if err != nil {
		log.Fatal(err)
	}
	pid = cmd.Process.Pid
	log.Printf("frpc subprocess pid is %d\n", cmd.Process.Pid)

	// read current tubed fingerprint
	f, err := os.ReadFile(tubedDir + "tubed.fgp")
	if err != nil {
		log.Fatal(err)
	}
	tubedfgp := string(f)
	log.Println("tubed fingerprint is " + tubedfgp)

	var i int = 0

LOOP:
	for {
		// pull tunnel config every 10s
		if i == 10 {
			i = 0
			// get tunnel fingerprint on API
			fgp := tunnel.GetFingerprint(baseUrl, token)

			// if different fingerprint than tubed --> pull new config
			if fgp != tubedfgp {
				log.Println("new tunnel config detected on server --> pulling")

				// pull new config
				tunnel.PullConfig(tubedDir, baseUrl, token)

				// update tubed fingerprint
				tubedfgp = fgp
				log.Println("new config successfully pulled --> tubed fingerprint is " + tubedfgp)
			}
		}

		// check if frpc pid ok (api call + pid)

		time.Sleep(time.Second) // this is work to be done by worker.

		select {
		case <-stop:
			break LOOP
		default:
		}

		i++
	}
	done <- struct{}{}
}

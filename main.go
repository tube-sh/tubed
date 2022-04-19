package main

import (
	"archive/zip"
	"bytes"
	"crypto/tls"
	_ "embed"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/sevlyar/go-daemon"
)

//go:embed frpc
var frpc []byte

var pid int
var tubedDir string = "/etc/tubed/"
var frpcConfLocation string = "/etc/tubed/"

//var base_url string = "https://api.tube.sh"
var base_url string = "http://localhost:8000"

type tunnelboostrap struct {
	TunnelToken string `json:"tunnel_token"`
}

var (
	signal = flag.String("s", "", `Send signal to the daemon:
  quit — graceful shutdown
  stop — fast shutdown
  reload — reloading the configuration file`)
)
var proxy string

// Add proxy to frpc.ini config file
func addProxy(proxy string) {
	input, err := ioutil.ReadFile(frpcConfLocation + "frpc.ini")
	if err != nil {
		log.Fatalln(err)
	}

	lines := strings.Split(string(input), "\n")

	for i, line := range lines {
		if strings.Contains(line, "[common]") {
			lines[i] = "[common]\nhttp_proxy = " + proxy
		}
	}
	output := strings.Join(lines, "\n")
	err = ioutil.WriteFile(frpcConfLocation+"frpc.ini", []byte(output), 0644)
	if err != nil {
		log.Fatalln(err)
	}
}

func processTubeResp(body []byte) {
	// write whole the body in frpc.zip
	err := ioutil.WriteFile("/tmp/frpc.zip", body, 0644)
	if err != nil {
		log.Println("Error while writing frpc.zip file:", err)
	}

	// read archive
	archive, err := zip.OpenReader("/tmp/frpc.zip")
	if err != nil {
		panic(err)
	}
	defer archive.Close()

	dst := frpcConfLocation
	for _, f := range archive.File {
		filePath := filepath.Join(dst, f.Name)

		if !strings.HasPrefix(filePath, filepath.Clean(dst)+string(os.PathSeparator)) {
			fmt.Println("invalid file path")
			return
		}
		if f.FileInfo().IsDir() {
			fmt.Println("creating directory...")
			os.MkdirAll(filePath, os.ModePerm)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
			panic(err)
		}

		dstFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			panic(err)
		}

		fileInArchive, err := f.Open()
		if err != nil {
			panic(err)
		}

		if _, err := io.Copy(dstFile, fileInArchive); err != nil {
			panic(err)
		}

		dstFile.Close()
		fileInArchive.Close()
	}

}

func initHttpClient() *http.Client {
	// Send req using http Client
	//var proxyUrl *url.URL
	var tr *http.Transport
	if proxy != "" {
		proxyUrl, _ := url.Parse(proxy)
		tr = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			Proxy:           http.ProxyURL(proxyUrl),
		}
	}

	client := &http.Client{Transport: tr}
	return client
}

func bootstrap(token string) {

	// write frpc binary in tubed dir
	err := os.WriteFile(tubedDir+"frpc", frpc, 0755)
	if err != nil {
		log.Fatalln(err)
	}

	// create /etc/frp directory if not exists
	err = os.Mkdir(frpcConfLocation, 0755)
	if err != nil && !os.IsExist(err) {
		log.Fatalln(err)
	}

	// build tube bootstrap URL
	urls := base_url + "/v1/tunnel/bootstrap"

	reqbody := &tunnelboostrap{
		TunnelToken: token,
	}

	// build body
	jsonBody, _ := json.Marshal(reqbody)

	// Create a new request using http
	req, err := http.NewRequest("POST", urls, bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	// Send req using http Client
	client := initHttpClient()
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Error on response:", err)
	}

	// read response
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("Error while reading the response bytes:", err)
	}

	processTubeResp(body)
}

func pullConfig(token string) {
	// build tube bootstrap URL
	url := base_url + "/v1/tunnel/pullconfig"

	reqbody := &tunnelboostrap{
		TunnelToken: token,
	}

	// build body
	jsonBody, _ := json.Marshal(reqbody)

	// Create a new request using http
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	// Send req using http Client
	client := initHttpClient()
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Error on response:", err)
	}

	// read response
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("Error while reading the response bytes:", err)
	}

	processTubeResp(body)

	// call frpc local api to hot reload config
	_, err = http.Get("http://127.0.0.1:7400/api/reload")
	if err != nil {
		log.Fatalln(err)
	}
}

func getTunnelFingerprint() string {
	// build tube bootstrap URL
	url := base_url + "/v1/tunnel/fingerprint"

	reqbody := &tunnelboostrap{
		TunnelToken: token,
	}

	// build body
	jsonBody, _ := json.Marshal(reqbody)

	// Create a new request using http
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	// Send req using http Client
	client := initHttpClient()
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Error on response:", err)
	}

	// read response
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("Error while reading the response bytes:", err)
	}

	type tunnelfingerprint struct {
		TunnelFingerprint string `json:"tunnel_fingerprint"`
	}

	var fgp tunnelfingerprint
	if err := json.Unmarshal(body, &fgp); err != nil {
		log.Fatal("Response unmarshal failed: " + err.Error())
	}

	return fgp.TunnelFingerprint
}

var token string

func main() {

	flag.StringVar(&proxy, "proxy", "", "proxy configuration")
	flag.Parse()

	// get tunnel token from file
	log.Println("get tubed tunnel token from " + tubedDir + "token")
	f, err := os.ReadFile(tubedDir + "token")
	if err != nil {
		panic(err)
	}
	token = strings.TrimSuffix(string(f), "\n")

	// if no tubed.bootstrapped file, do bootstrap
	if _, err := os.Stat(tubedDir + "tubed.bootstrapped"); errors.Is(err, os.ErrNotExist) {
		log.Println("start tubed bootstrap...")

		// bootstrap tunnel
		bootstrap(token)

		// create empty file to tag as bootstrapped
		log.Println("create tubed.bootstrapped file to " + tubedDir + "tubed.bootstrapped")
		_, err = os.Create(tubedDir + "tubed.bootstrapped")
		if err != nil {
			log.Fatal(err)
		}

		log.Println("tubed bootstrap finished successfully")
	}

	daemon.AddCommand(daemon.StringFlag(signal, "stop"), syscall.SIGTERM, termHandler)
	//daemon.AddCommand(daemon.StringFlag(signal, "reload"), syscall.SIGHUP, reloadHandler)

	if *signal != "stop" {
		log.Println("- - - - - - - - - - - - - - -")
		log.Println("starting tubed daemon...")
		log.Println("logfile location is /var/log/tubed.log")
	}

	runDaemon()

	if *signal == "stop" {
		log.Println("tubed daemon terminated")
	}

}

func runDaemon() {
	cntxt := &daemon.Context{
		PidFileName: "/var/run/tubed.pid",
		PidFilePerm: 0644,
		LogFileName: "/var/log/tubed.log",
		LogFilePerm: 0640,
		WorkDir:     "./",
		Umask:       027,
		//Args:        []string{"[go-daemon tubed]"},
		Args: nil,
	}

	if len(daemon.ActiveFlags()) > 0 {
		d, err := cntxt.Search()
		if err != nil {
			log.Fatalf("Unable send signal to the daemon: %s", err.Error())
		}
		daemon.SendCommands(d)
		return
	}

	d, err := cntxt.Reborn()
	if err != nil {
		log.Fatalln(err)
	}
	if d != nil {
		return
	}
	defer cntxt.Release()

	log.Println("tubed daemon started")

	go worker()

	err = daemon.ServeSignals()
	if err != nil {
		log.Printf("Error: %s", err.Error())
	}

	log.Println("tubed daemon terminated")
}

var (
	stop = make(chan struct{})
	done = make(chan struct{})
)

func worker() {

	if proxy != "" {
		//log.Println("add proxy config to frpc.ini")
		//addProxy(proxy)

		log.Println("set http_proxy env var")
		os.Setenv("http_proxy", proxy)
	}

	cmd := exec.Command("/etc/tubed/frpc", "-c", "/etc/tubed/frpc.ini")
	cmd.Stdout = os.Stdout
	err := cmd.Start()
	if err != nil {
		log.Fatal(err)
	}
	pid = cmd.Process.Pid
	log.Printf("frpc subprocess pid is %d\n", cmd.Process.Pid)

	var i int = 0
	var tubedfgp string

LOOP:
	for {
		// pull tunnel config every 10s
		if i == 10 {
			i = 0
			// get tunnel fingerprint on API
			fgp := getTunnelFingerprint()

			// if different fingerprint than tubed --> pull new config
			if fgp != tubedfgp {
				log.Println("new tunnel config detected on server --> pulling")

				// pull new config
				pullConfig(token)

				// update tubed fingerprint
				tubedfgp = fgp
				log.Println("new config successfully pulled --> update tubed fingerprint")
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

func termHandler(sig os.Signal) error {
	err := syscall.Kill(pid, 9)
	if err == nil {
		log.Println("frpc subprocess stopped")
	}
	stop <- struct{}{}
	if sig == syscall.SIGQUIT {
		<-done
	}
	return daemon.ErrStop
}

func reloadHandler(sig os.Signal) error {
	log.Println("configuration reloaded")
	return nil
}

package main

import (
	"archive/zip"
	"bytes"
	_ "embed"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

//go:embed frpc
var frpc []byte

//go:embed frpc.service
var frpcService []byte

var frpcConfLocation string = "/etc/frp/"
var base_url string = "http://127.0.0.1:8000"

type tunnelboostrap struct {
	TunnelToken string `json:"tunnel_token"`
}

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

	dst := "/etc/frp"
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

	// add proxy to config if requested
	if proxy != "" {
		addProxy(proxy)
	}
}

func bootstrap() {

	// write frpc binary in /usr/bin
	err := os.WriteFile("/usr/bin/frpc", frpc, 0755)
	if err != nil {
		log.Fatalln(err)
	}

	// write frpc.service in /lib/systemd/system
	err = os.WriteFile("/lib/systemd/system/frpc.service", frpcService, 0644)
	if err != nil {
		log.Fatalln(err)
	}

	// create /etc/frp directory if not exists
	err = os.Mkdir(frpcConfLocation, 0755)
	fmt.Println(os.IsExist(err))
	if err != nil && !os.IsExist(err) {
		log.Fatalln(err)
	}

	// build tube bootstrap URL
	url := base_url + "/v1/tunnel/bootstrap"

	reqbody := &tunnelboostrap{
		TunnelToken: token,
	}

	// build body
	jsonBody, _ := json.Marshal(reqbody)

	// Create a new request using http
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	// Send req using http Client
	client := &http.Client{}
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

	//fmt.Println(string(body))
	processTubeResp(body)
}

func update() {
	// build tube bootstrap URL
	url := base_url + "/v1/tunnel/pull"

	reqbody := &tunnelboostrap{
		TunnelToken: token,
	}

	// build body
	jsonBody, _ := json.Marshal(reqbody)

	// Create a new request using http
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	// Send req using http Client
	client := &http.Client{}
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

	_, err = http.Get("http://127.0.0.1:7400/api/reload")
	if err != nil {
		log.Fatalln(err)
	}
}

var token string
var proxy string
var help bool

func main() {

	// check if run as root
	if os.Getuid() != 0 {
		log.Fatalln("tubed must be run as root")
	}

	// get command options
	flag_b := flag.NewFlagSet("bootstrap", flag.ExitOnError)
	flag_b.StringVar(&token, "token", "", "tunnel token")
	flag_b.StringVar(&proxy, "proxy", "", "proxy configuration")

	flag_p := flag.NewFlagSet("update", flag.ExitOnError)
	flag_p.StringVar(&token, "token", "", "tunnel token")
	flag_p.StringVar(&proxy, "proxy", "", "proxy configuration")

	//flag.BoolVar(&help, "help", false, "Display Help")
	//flag.Parse()

	if len(os.Args) < 2 {
		fmt.Println("expected 'bootstrap' or 'update' subcommands")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "bootstrap":
		flag_b.Parse(os.Args[2:])
		bootstrap()
	case "update":
		flag_p.Parse(os.Args[2:])
		update()
	default:
		fmt.Println("expected 'bootstrap' or 'update' subcommands")
		os.Exit(1)
	}

	//out, _ := exec.Command("./frpc_binary").Output()
	//fmt.Printf("Output: %s\n", out)

	// tubed bootstrap -token *** -proxy ...
	// tubed expose -name app -subdomain app -ip x.x.x.x -port 0
	// tubed run
	// tubed install-daemon
	/*
			[app]
		subdomain = httpbin
		local_ip = 10.170.182.201
		local_port = 8080
		type = http
	*/
}

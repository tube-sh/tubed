package main

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

//go:embed frpc
var frpc []byte

//go:embed frpc.service
var frpcService []byte

var frpcConfLocation string = "/etc/frp/"

func getTunnelConfig(token string) {

	// build tube bootstrap URL
	url := "http://127.0.0.1:8000/tunnelbootstrap"
cur
	// build body
	jsonBody, _ := json.Marshal(map[string]string{
		"tunnel_token": token,
	})

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
	fmt.Println(body)

	// write whole the body in frpc.ini
	err = ioutil.WriteFile(frpcConfLocation+"frpc.ini", body, 0644)
	if err != nil {
		log.Println("Error while writing frpc.ini file:", err)
	}
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

var token string
var proxy string
var help bool

func main() {
	// get command options
	flag.StringVar(&token, "token", "", "tunnel configuration token")
	flag.StringVar(&proxy, "proxy", "", "proxy configuration")
	flag.BoolVar(&help, "help", false, "Display Help")
	flag.Parse()

	// check if help was called explicitly
	if help {
		fmt.Println(">> Display help screen")
		os.Exit(1)
	}

	// check if run as root
	if os.Getuid() != 0 {
		log.Fatalln("tubed must be run as root")
	}

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

	// get tunnel configuration from tube server
	getTunnelConfig(token)

	// add proxy to config if requested
	if proxy != "" {
		addProxy(proxy)
	}

	//out, _ := exec.Command("./frpc_binary").Output()
	//fmt.Printf("Output: %s\n", out)
}

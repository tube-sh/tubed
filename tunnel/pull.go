package tunnel

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/tube-sh/tubed/utils"
)

func PullConfig(tubedDir string, baseUrl string, token string) {
	// variables
	url := baseUrl + "/v1/tunnel/pullconfig"

	// build request body
	reqbody := &tunnelboostrap{
		TunnelToken: token,
	}
	jsonBody, _ := json.Marshal(reqbody)

	// call api to get tunnel config
	resp := utils.PostJSON(url, jsonBody)

	// write response in frpc.zip
	err := ioutil.WriteFile("/tmp/frpc.zip", resp, 0644)
	if err != nil {
		log.Println("Error while writing frpc.zip file:", err)
	}

	// extract frpc.zip file to tubed dir
	utils.ExtractZipFile("/tmp/frpc.zip", tubedDir)

	// call frpc local api to hot reload config
	_, err = http.Get("http://127.0.0.1:7400/api/reload")
	if err != nil {
		log.Fatalln(err)
	}
}

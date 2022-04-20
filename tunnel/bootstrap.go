package tunnel

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"

	"github.com/tube-sh/tubed/utils"
)

func Bootstrap(tubedDir string, baseUrl string, token string, frpcBinary []byte) {

	// write frpc binary in tubed dir
	err := os.WriteFile(tubedDir+"frpc", frpcBinary, 0755)
	if err != nil {
		log.Fatalln(err)
	}

	// variables
	url := baseUrl + "/v1/tunnel/bootstrap"

	// build request body
	reqbody := &tunnelboostrap{
		TunnelToken: token,
	}
	jsonBody, _ := json.Marshal(reqbody)

	// call api to get tunnel bootstrap config
	resp := utils.PostJSON(url, jsonBody)

	// write response in frpc.zip
	err = ioutil.WriteFile("/tmp/frpc.zip", resp, 0644)
	if err != nil {
		log.Println("Error while writing frpc.zip file:", err)
	}

	// extract frpc.zip file to tubed dir
	utils.ExtractZipFile("/tmp/frpc.zip", tubedDir)
}

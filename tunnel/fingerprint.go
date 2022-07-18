package tunnel

import (
	"encoding/json"
	"log"

	"github.com/tube-sh/tubed/utils"
)

func GetFingerprint(baseUrl string, token string) string {
	// variables
	url := baseUrl + "/v1/tunnel/fingerprint"

	// build request body
	reqbody := &tunnelboostrap{
		TunnelToken: token,
	}
	jsonBody, _ := json.Marshal(reqbody)

	// call api to get config fingerprint
	resp := utils.PostJSON(url, jsonBody)

	// process response
	var fgp tunnelfingerprint
	if err := json.Unmarshal(resp, &fgp); err != nil {
		log.Fatal("Response unmarshal failed: " + err.Error())
	}

	// return fingerprint
	return fgp.TunnelFingerprint
}

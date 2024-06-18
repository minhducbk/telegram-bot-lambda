package ip

import (
	"encoding/json"
	"net/http"
)

type IPResponse struct {
	IP string `json:"ip"`
}

func GetPublicIP() (string, error) {
	resp, err := http.Get("https://api.ipify.org?format=json")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var ipResp IPResponse
	if err := json.NewDecoder(resp.Body).Decode(&ipResp); err != nil {
		return "", err
	}

	return ipResp.IP, nil
}

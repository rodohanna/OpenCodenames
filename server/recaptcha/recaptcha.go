package recaptcha

// This implementation is borrowed and slightly modified from:
// https://github.com/dpapathanasiou/go-recaptcha

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"time"
)

// Response response from Google
type Response struct {
	Success     bool      `json:"success"`
	Score       float64   `json:"score"`
	Action      string    `json:"action"`
	ChallengeTS time.Time `json:"challenge_ts"`
	Hostname    string    `json:"hostname"`
	ErrorCodes  []string  `json:"error-codes"`
}

const recaptchaServerName = "https://www.google.com/recaptcha/api/siteverify"

var recaptchaPrivateKey string

func check(remoteip, response string) (r Response, err error) {
	resp, err := http.PostForm(recaptchaServerName,
		url.Values{"secret": {recaptchaPrivateKey}, "remoteip": {remoteip}, "response": {response}})
	if err != nil {
		log.Printf("Post error: %s\n", err)
		return
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Read error: could not read body: %s", err)
		return
	}
	err = json.Unmarshal(body, &r)
	if err != nil {
		log.Printf("Read error: got invalid JSON: %s", err)
		return
	}
	return
}

// Confirm kicks off the check
func Confirm(remoteip, response string) (result bool, err error) {
	resp, err := check(remoteip, response)
	result = resp.Success
	return
}

// Init sets the private key
func Init(key string) {
	recaptchaPrivateKey = key
}

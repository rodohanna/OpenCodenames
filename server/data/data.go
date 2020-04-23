package data

import (
	"bufio"
	"io/ioutil"
	"log"
	"os"
	"sync"
)

// WordList a slice of strings containing all possible words
type WordList []string

var (
	once         sync.Once
	instance     WordList
	recaptchaKey string
)

// GetWordList returns the word list
func GetWordList() WordList {
	once.Do(func() {
		file, err := os.Open("./data/wordlist.txt")
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()
		scanner := bufio.NewScanner(file)
		instance = make([]string, 0)
		for scanner.Scan() {
			instance = append(instance, scanner.Text())
		}
		if err := scanner.Err(); err != nil {
			log.Fatal(err)
		}

	})
	return instance
}

// GetReCAPTCHAKey returns the token necessary to check ReCAPTCHA tests
func GetReCAPTCHAKey() string {
	once.Do(func() {
		key, err := ioutil.ReadFile("./recaptcha-key.txt")
		if err != nil {
			log.Fatal(err)
		}
		recaptchaKey = string(key)
	})
	return recaptchaKey
}

package data

import (
	"bufio"
	"log"
	"os"
	"sync"
)

// WordList a slice of strings containing all possible words
type WordList []string

var (
	once     sync.Once
	instance WordList
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

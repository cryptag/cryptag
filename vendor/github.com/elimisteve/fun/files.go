package fun

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

func OpenAndRead(filename string) (string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer file.Close()

	contents, err := ioutil.ReadAll(file)
	if err != nil {
		return "", err
	}
	return string(contents), nil
}

func MaybeFatal(err error) {
    if err != nil {
		log.Fatal(err)
	}
}

func MaybeFatalAt(where string, err error) {
    if err != nil {
		log.Fatalf("Error near %s: %v\n", where, err)
	}
}

// OpenAndAppend opens the file with the given filename and appends
// the given string
func OpenAndAppend(filename string, toAppend string) error {
	// Open file with given filename w/append-only permissions; create
	// it if it doesn't already exist
	flags := os.O_APPEND | os.O_WRONLY | os.O_CREATE
	file, err := os.OpenFile(filename, flags, 0666)
	if err != nil {
		return fmt.Errorf("Error opening log file: %v", err)
	}
	defer file.Close()
	if _, err = file.Write([]byte(toAppend)); err != nil {
		return fmt.Errorf("Error writing to log file: %v", err)
	}
	return nil
}

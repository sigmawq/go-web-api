package main

import "os"
import "math/rand"

func fileExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func createFileIfDoesntExist(path string) {
	exists, err := fileExists(path)
	if err != nil {
		panic(err)
	}

	if !exists {
		_, err := os.Create(path)
		if err != nil {
			panic(err)
		}
	}
}

func randomString() string {
	min := 3
	max := 8
	length := min + rand.Intn(max-min)
	buff := make([]byte, length)
	for i, _ := range buff {
		buff[i] = 65 + byte(rand.Intn(122-65)) // Basically all ASCII letters + special chars
	}

	return string(buff)
}
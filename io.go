package main

import (
	"bufio"
	"os"
)

func (doc *DocStruct) handleEventLoad() error {
	// open file and read line by line
	f, err := os.Open(doc.filename)
	if err != nil {
		return err
	}
	defer f.Close()
	doc.text = make([]LineType, 0, 256)
	scanner := bufio.NewScanner(f) // default delimiter is new line
	for scanner.Scan() {
		doc.text = append(doc.text, LineType(scanner.Text()))
	}
	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}

func (doc *DocStruct) handleEventSave() error {
	f, err := os.Create("data.txt")
	if err != nil {
		return err
	}
	defer f.Close()
	for _, line := range doc.text {
		_, err = f.WriteString(string(line))
		if err != nil {
			return err
		}
	}
	return nil
}

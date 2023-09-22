package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"unicode"
	// bencode "github.com/jackpal/bencode-go" // Available if you need it!
)

func decodeBencode(bencodedString string) (interface{}, error) {
	benStrLen := len(bencodedString)
	if unicode.IsDigit(rune(bencodedString[0])) {
		var firstColonIndex int

		for i := 0; i < benStrLen; i++ {
			if bencodedString[i] == ':' {
				firstColonIndex = i
				break
			}
		}

		lengthStr := bencodedString[:firstColonIndex]

		length, err := strconv.Atoi(lengthStr)
		if err != nil {
			return "", err
		}

		return bencodedString[firstColonIndex+1 : firstColonIndex+1+length], nil

	} else if bencodedString[0] == 'i' && benStrLen >= 3 {

		decodedNumberString := bencodedString[1 : benStrLen-1]
		number, err := strconv.Atoi(decodedNumberString)
		if err != nil {
			return "", err
		}

		return number, nil
	} else {
		return "", fmt.Errorf("only strings are supported at the moment")
	}
}

func main() {
	command := os.Args[1]

	if command == "decode" {

		bencodedValue := os.Args[2]

		decoded, err := decodeBencode(bencodedValue)
		if err != nil {
			fmt.Println(err)
			return
		}

		jsonOutput, _ := json.Marshal(decoded)
		fmt.Println(string(jsonOutput))

	} else {
		fmt.Println("Unknown command: " + command)
		os.Exit(1)
	}
}

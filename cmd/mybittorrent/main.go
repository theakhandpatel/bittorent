package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"unicode"
	// bencode "github.com/jackpal/bencode-go" // Available if you need it!
)

var ErrPoorEncoded error = fmt.Errorf("poorly encoded string")

func decodeBencode(bencodedString string, curIndex int) (interface{}, int, error) {
	if curIndex >= len(bencodedString) {
		return nil, -1, ErrPoorEncoded
	}
	var result interface{}
	var err error

	if unicode.IsDigit(rune(bencodedString[curIndex])) {
		result, curIndex, err = decodeForString(bencodedString, curIndex)
	} else if bencodedString[curIndex] == 'i' {
		result, curIndex, err = decodeForNumber(bencodedString, curIndex)
	} else if bencodedString[curIndex] == 'l' {
		result, curIndex, err = decodeForList(bencodedString, curIndex)
	} else if bencodedString[curIndex] == 'd' {
		result, curIndex, err = decodeForDictionary(bencodedString, curIndex)
	} else {
		err = fmt.Errorf("only strings are supported at the moment")
	}

	if err != nil {
		return nil, -1, err
	}
	return result, curIndex, nil
}

func decodeForString(bencodedString string, curIndex int) (string, int, error) {
	var firstColonIndex int = curIndex

	for i := curIndex; i < len(bencodedString); i++ {
		if bencodedString[i] == ':' {
			firstColonIndex = i
			break
		}
	}
	if firstColonIndex == curIndex {
		return "", -1, ErrPoorEncoded
	}
	lengthStr := bencodedString[curIndex:firstColonIndex]
	length, err := strconv.Atoi(lengthStr)
	if err != nil {
		return "", -1, err
	}

	stringEndIndex := firstColonIndex + 1 + length
	if stringEndIndex > len(bencodedString) {
		return "", -1, ErrPoorEncoded
	}

	return bencodedString[firstColonIndex+1 : stringEndIndex], stringEndIndex - 1, nil
}

func decodeForNumber(bencodedString string, curIndex int) (int, int, error) {
	benStrLen := len(bencodedString) - curIndex
	if benStrLen <= 2 {
		return 0, -1, ErrPoorEncoded
	}

	eIndex := findOccurrenceAfterIndex(bencodedString, 'e', curIndex)

	if eIndex == -1 {
		return 0, -1, ErrPoorEncoded
	}
	decodedNumberString := bencodedString[curIndex+1 : eIndex]
	number, err := strconv.Atoi(decodedNumberString)
	if err != nil {
		return 0, eIndex, err
	}

	return number, eIndex, nil
}

// l5:helloi52ee
func decodeForList(bencodedString string, curIndex int) (interface{}, int, error) {
	var result []interface{}
	for curIndex = curIndex + 1; bencodedString[curIndex] != 'e'; {
		item, itr, err := decodeBencode(bencodedString, curIndex)
		if err != nil {
			return nil, -1, err
		}
		curIndex = itr + 1
		result = append(result, item)
	}

	return result, curIndex, nil
}

func decodeForDictionary(bencodedString string, curIndex int) (interface{}, int, error) {
	result := make(map[string]interface{})
	for curIndex = curIndex + 1; bencodedString[curIndex] != 'e'; {
		key, itr, err := decodeForString(bencodedString, curIndex)
		if err != nil {
			return nil, -1, err
		}
		itr = itr + 1
		value, itr, err := decodeBencode(bencodedString, itr)
		if err != nil {
			return nil, -1, err
		}
		curIndex = itr + 1
		result[key] = value
	}
	return result, curIndex, nil
}

func findOccurrenceAfterIndex(str string, char rune, curIndex int) int {
	for i := curIndex; i < len(str); i++ {
		if rune(str[i]) == char {
			return i
		}
	}
	return -1 // Character not found after curIndex
}

type TorrentMeta struct {
	Announce string
	Info     struct {
		Length       int
		Name         string
		Piece_length int
		Pieces       string
	}
}

func parseTorrent(filePath string) (TorrentMeta, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return TorrentMeta{}, err
	}
	fileContent, err := io.ReadAll(file)
	if err != nil {
		return TorrentMeta{}, err
	}
	decodedContent, _, err := decodeBencode(string(fileContent), 0)
	if err != nil {
		return TorrentMeta{}, err
	}

	metaDict := decodedContent.(map[string]interface{})
	metaInfoDict := metaDict["info"].(map[string]interface{})

	return TorrentMeta{
		Announce: metaDict["announce"].(string),
		Info: struct {
			Length       int
			Name         string
			Piece_length int
			Pieces       string
		}{
			Length:       metaInfoDict["length"].(int),
			Name:         metaInfoDict["name"].(string),
			Piece_length: metaInfoDict["piece length"].(int),
			Pieces:       metaInfoDict["pieces"].(string),
		},
	}, nil
}

func main() {
	command := os.Args[1]

	if command == "decode" {

		bencodedValue := os.Args[2]

		decoded, _, err := decodeBencode(bencodedValue, 0)
		if err != nil {
			fmt.Println(err)
			return
		}

		jsonOutput, _ := json.Marshal(decoded)
		if string(jsonOutput) != "null" {
			fmt.Println(string(jsonOutput))
		} else {
			fmt.Println("[]")
		}

	} else if command == "info" {
		torrentFilePath := os.Args[2]
		torrentMetaData, err := parseTorrent(torrentFilePath)
		if err != nil {
			fmt.Println("Error parsing file")
		}
		fmt.Printf("Tracker URL: %s\nLength: %d\n", torrentMetaData.Announce, torrentMetaData.Info.Length)
	} else {
		fmt.Println("Unknown command: " + command)
		os.Exit(1)
	}
}

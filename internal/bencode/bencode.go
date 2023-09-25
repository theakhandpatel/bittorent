package bencode

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"unicode"
)

var ErrPoorEncoded error = fmt.Errorf("poorly encoded string")

func Decode(bencodedString string) (interface{}, error) {
	if len(bencodedString) < 2 {
		return "", ErrPoorEncoded
	}
	decoded, _, err := decodeBencode(bencodedString, 0)
	return decoded, err
}

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
} // Encode encodes data into Bencode format.

func Encode(data interface{}) (string, error) {
	switch data := data.(type) {
	case string:
		return encodeString(data), nil
	case int:
		return encodeNumber(data), nil
	case []interface{}:
		return encodeList(data)
	case map[string]interface{}:
		return encodeDictionary(data)
	default:
		return "", fmt.Errorf("unsupported data type")
	}
}

func encodeNumber(number int) string {
	return fmt.Sprintf("i%de", number)
}

func encodeString(text string) string {
	return fmt.Sprintf("%d:%s", len(text), text)
}

func encodeList(list []interface{}) (string, error) {
	var encodedList []string

	for _, item := range list {
		encodedItem, err := Encode(item)
		if err != nil {
			return "", err
		}
		encodedList = append(encodedList, encodedItem)
	}
	return "l" + strings.Join(encodedList, "") + "e", nil
}

func encodeDictionary(dict map[string]interface{}) (string, error) {
	var encodedDict []string
	var keys []string
	for k := range dict {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, key := range keys {
		encodedValue, err := Encode(dict[key])
		if err != nil {
			return "", err
		}
		encodedDict = append(encodedDict, encodeString(key)+encodedValue)
	}
	return "d" + strings.Join(encodedDict, "") + "e", nil
}

func findOccurrenceAfterIndex(str string, char rune, curIndex int) int {
	for i := curIndex; i < len(str); i++ {
		if rune(str[i]) == char {
			return i
		}
	}
	return -1 // Character not found after curIndex
}

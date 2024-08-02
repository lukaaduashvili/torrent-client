package bencode

import (
	"fmt"
	"strconv"
	"unicode"
)

func decodeString(bencodedString string) (res string, length int, err error) {
	var firstColonIndex int

	for i := 0; i < len(bencodedString); i++ {
		if bencodedString[i] == ':' {
			firstColonIndex = i
			break
		}
	}

	lengthStr := bencodedString[:firstColonIndex]

	length, err = strconv.Atoi(lengthStr)
	if err != nil {
		return "", 0, err
	}

	return bencodedString[firstColonIndex+1 : firstColonIndex+1+length], length + firstColonIndex + 1, nil
}

func decodeInt(bencodedString string) (res int, length int, err error) {
	var startIdx int
	var endIdx int
	for i := 0; i < len(bencodedString); i++ {
		if bencodedString[i] == 'i' {
			startIdx = i + 1
		}
		if bencodedString[i] == 'e' {
			endIdx = i
			break
		}
	}

	res, err = strconv.Atoi(bencodedString[startIdx:endIdx])

	return res, endIdx + 1, err
}

func decodeList(bencodedString string) (res []interface{}, length int, err error) {
	result := make([]interface{}, 0)

	for i := 1; i < len(bencodedString); {
		val, length, err := Decode(bencodedString[i:])
		if err != nil {
			return nil, 0, err
		}

		result = append(result, val)

		i += length

		if i >= len(bencodedString) {
			return nil, 0, fmt.Errorf("List decoding out of bounds")
		}
		if bencodedString[i] == 'e' {
			return result, i, nil
		}
	}

	return nil, 0, fmt.Errorf("List missing ending character")
}

// Make this private and decode into structs
func Decode(bencodedString string) (interface{}, int, error) {
	if unicode.IsDigit(rune(bencodedString[0])) {
		return decodeString(bencodedString)
	} else if bencodedString[0] == 'i' {
		return decodeInt(bencodedString)
	} else if bencodedString[0] == 'l' {
		return decodeList(bencodedString)
	} else {
		return "", 0, fmt.Errorf("only strings are supported at the moment")
	}
}

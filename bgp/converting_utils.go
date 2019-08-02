package bgp

import (
	"fmt"
	"strconv"
	"strings"
)

func toUint32(value string) uint32 {
	number, _ := strconv.ParseUint(value, 10, 32)
	return uint32(number)
}

func ToUint32(value string) uint32 {
	number, _ := strconv.ParseUint(value, 10, 32)
	return uint32(number)
}

func stringsToNumbers(strings []string) (numbers []uint32) {
	numbers = []uint32{}

	for _, value := range strings {
		number, _ := strconv.ParseUint(value, 10, 32)
		numbers = append(numbers, uint32(number))
	}

	return numbers
}

func StringsToNumbers(strings []string) (numbers []uint32) {
	numbers = []uint32{}

	for _, value := range strings {
		number, _ := strconv.ParseUint(value, 10, 32)
		numbers = append(numbers, uint32(number))
	}

	return numbers
}

func NumbersToString(numbers []uint32, delimeter string) string {
	return strings.Trim(strings.Replace(fmt.Sprint(numbers), " ", delimeter, -1), "[]")
}

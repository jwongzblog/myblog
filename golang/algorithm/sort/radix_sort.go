package main

import (
	"fmt"
	"math"
	"strconv"
)

var stack = map[int][]int{
	0: []int{},
	1: []int{},
	2: []int{},
	3: []int{},
	4: []int{},
	5: []int{},
	6: []int{},
	7: []int{},
	8: []int{},
	9: []int{},
}

func digitSort(array []int, digit int) []int {
	var newArray = make([]int, 0)

	for _, value := range array {
		numDigit := (value / int(math.Pow10(digit))) % 10
		stack[numDigit] = append(stack[numDigit], value)
	}

	fmt.Print(stack, "\n")

	for key := 0; key < 10; key++ {
		for range stack[key] {
			newArray = append(newArray, stack[key][0])
			stack[key] = stack[key][1:]
		}
	}

	fmt.Print(stack, newArray, "\n")
	return newArray
}

func radixSort(array []int) []int {
	max := 0
	for _, value := range array {
		nLen := len(strconv.Itoa(value))
		if nLen > max {
			max = nLen
		}
	}

	var result = make([]int, 0)
	for i := 0; i < max; i++ {
		if i == 0 {
			result = digitSort(array, i)
		} else {
			result = digitSort(result, i)
		}
	}

	return result
}

func main() {
	array := []int{17, 15, 12, 3, 3777, 2, 3, 19, 88, 3, 1, 22222, 4, 1, 3, 20}
	result := radixSort(array)

	fmt.Print(result)
}

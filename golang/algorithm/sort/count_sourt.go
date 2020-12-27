package main

import "fmt"

func countSort(array []int) []int {
	aLen := len(array)
	if aLen < 2 {
		return array
	}

	var result = make([]int, aLen)
	max := 0
	for _, value := range array {
		if max < value {
			max = value
		}
	}

	if max == 0 {
		return array
	}

	var b = make([]int, max+1)
	for _, value := range array {
		b[value] += 1
	}

	fmt.Print(b, "\n")

	for index, _ := range b {
		if index == 0 {
			continue
		}
		b[index] = b[index] + b[index-1]
	}

	fmt.Print(b, len(b), "\n")
	for _, value := range array {
		result[b[value]-1] = value
		b[value] = b[value] - 1
	}

	return result
}

// func main() {
// 	array := []int{17, 15, 12, 3, 3, 2, 3, 19, 88, 3, 1, 2, 4, 1, 3, 20}
// 	result := countSort(array)

// 	fmt.Print(result)
// }

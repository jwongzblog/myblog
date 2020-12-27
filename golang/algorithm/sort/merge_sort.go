package main

import (
	"fmt"
)

func mergeSort(array []int) []int {
	num := len(array)
	if num < 2 {
		return array
	}

	split := num / 2
	a := mergeSort(array[0:split])
	b := mergeSort(array[split:])

	return merge(a, b)
}

func merge(a, b []int) []int {
	var c = make([]int, 0)

	for len(a) != 0 && len(b) != 0 {
		if a[0] <= b[0] {
			c = append(c, a[0])
			a = a[1:]
			continue
		} else {
			c = append(c, b[0])
			b = b[1:]
			continue
		}

		fmt.Print(c, "\n")
	}

	for len(a) != 0 {
		c = append(c, a[0])
		a = a[1:]
	}

	for len(b) != 0 {
		c = append(c, b[0])
		b = b[1:]
	}

	return c
}

// func main() {
// 	b := []int{17, 15, 12, 3, 3, 2, 3, 19, 88, 3, 1, 2, 4, 1, 3, 20}
// 	result := mergeSort(b)

// 	fmt.Print(result)
// }

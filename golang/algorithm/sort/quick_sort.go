package main

import "fmt"

func quickSort(array []int, start, end int) {
	fmt.Print(start, end, "\n")
	if start < end {
		pos := partition(array, start, end)
		quickSort(array, start, pos-1)
		quickSort(array, pos+1, end)
	}
}

func partition(array []int, start, end int) int {
	x := array[end]
	i := start
	for j := start; j < end; j++ {
		if array[j] < x {

			swap(array, j, i)
			i += 1
		}
	}

	swap(array, i, end)

	//fmt.Print(i, array, "\n")
	return i
}

func swap(array []int, a, b int) {
	fmt.Print(array, "\n")
	array[a], array[b] = array[b], array[a]
	fmt.Print(array, "\n")
}

// func main() {
// 	array := []int{17, 15, 12, 3, 3, 2, 3, 19, 88, 3, 1, 2, 4, 1, 3, 20}
// 	quickSort(array, 0, len(array)-1)

// 	fmt.Print(array)
// }

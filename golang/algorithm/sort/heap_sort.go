package main

import "fmt"

func heapSort(array []int) {
	aLen := len(array)
	if aLen < 2 {
		return
	}

	middle := aLen / 2

	for i := middle; i > -1; i-- {
		buildMaxHeap(array, i, aLen-1)
	}

	for i := aLen - 1; i > -1; i-- {
		swap(array, 0, i)
		buildMaxHeap(array, 0, i-1)
	}
}

func buildMaxHeap(array []int, i, end int) {
	lLeaf := 2*i + 1
	if lLeaf > end {
		return
	}

	tmp := lLeaf
	rLeaf := 2*i + 2
	if rLeaf <= end && array[rLeaf] > array[lLeaf] {
		tmp = rLeaf
	}

	if array[i] > array[tmp] {
		return
	}

	swap(array, i, tmp)
	buildMaxHeap(array, tmp, end)
}

// func swap(array []int, a, b int) {
// 	fmt.Print(array, "\n")
// 	array[a], array[b] = array[b], array[a]
// 	fmt.Print(array, "\n")
// }

func main() {
	array := []int{17, 15, 12, 3, 3, 2, 3, 19, 88, 3, 1, 2, 4, 1, 3, 20}
	heapSort(array)

	fmt.Print(array)
}

package main

func shellSort(array []int) {
	aLen := len(array)
	if aLen < 2 {
		return
	}

	step := aLen / 2
	for step > 0 {
		for i := step; i < aLen; i++ {
			j := i
			for j >= step && array[j] < array[j-step] {
				array[j], array[j-step] = array[j-step], array[j]
				j -= step
			}
		}

		step /= 2
	}
}

// func main() {
// 	array := []int{17, 15, 12, 3, 3777, 2, 3, 19, 88, 3, 1, 22222, 4, 1, 3, 20}
// 	shellSort(array)

// 	fmt.Print(array)
// }

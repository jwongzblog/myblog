package main

func insert(key, index int, array []int) []int {
	var c = make([]int, 0)
	for _, value := range array[0:index] {
		c = append(c, value)
	}
	c = append(c, key)

	for _, value := range array[index:] {
		c = append(c, value)
	}

	return c
}

// func main() {
// 	var b = make([]int, 0)
// 	a := []int{17, 15, 12, 3, 3, 2, 3, 19, 88, 3, 1, 2, 4, 1, 3, 20}

// 	for _, a_key := range a {
// 		if len(b) == 0 {
// 			b = append(b, a_key)
// 			continue
// 		}

// 		for b_index, b_key := range b {
// 			fmt.Print(a_key, b_key, b, "\n")
// 			if a_key <= b_key {
// 				b = insert(a_key, b_index, b)
// 				break
// 			} else {
// 				fmt.Print(a_key, b[b_index], "\n")
// 				if b_index != len(b)-1 {

// 					if a_key <= b[b_index+1] {
// 						b = insert(a_key, b_index+1, b)
// 						break
// 					} else {
// 						continue
// 					}
// 				}

// 				b = append(b, a_key)
// 				break
// 			}
// 		}
// 	}

// 	fmt.Print(b)
// }

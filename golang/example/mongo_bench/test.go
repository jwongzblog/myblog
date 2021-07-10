// You can edit this code!
// Click here and start typing.
package main

import "fmt"

func hello(c chan int) {
	fmt.Print(<-c)
}

func thread() {
	c := make(chan int, 10)
	for {
		go hello(c)
		c <- 0
	}

}

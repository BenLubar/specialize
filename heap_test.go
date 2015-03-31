package main_test

import (
	"os"
	"strconv"

	"github.com/BenLubar/specialize"
)

func ExampleHeapSort() {
	var h main.IntPriorityQueue

	h.Add(1)
	h.Add(3)
	h.Add(6)
	h.Add(5)
	h.Add(2)
	h.Add(4)

	for h.Len() != 0 {
		// We have to avoid fmt.Println here as it uses reflection.
		os.Stdout.WriteString(strconv.Itoa(h.Remove()) + "\n")
	}

	// Output:
	// 1
	// 2
	// 3
	// 4
	// 5
	// 6
}

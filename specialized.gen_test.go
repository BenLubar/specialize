//+build !no_specialized

package main_test

import "github.com/BenLubar/specialize"
import "os"
import "strconv"

func ExampleHeapSort_Specialized() {
	var (
		t0  *main.IntPriorityQueue
		t7  *os.File
		t8  int
		t9  string
		t10 string
		t12 main.IntPriorityQueue
		t13 int
		t14 bool
	)

	//b0: // entry
	t0 = new(main.IntPriorityQueue) // h
	specialized_main__pmain_dIntPriorityQueue_dAdd_a_pmain_dIntPriorityQueue_aint(t0, (int)(1))
	specialized_main__pmain_dIntPriorityQueue_dAdd_a_pmain_dIntPriorityQueue_aint(t0, (int)(3))
	specialized_main__pmain_dIntPriorityQueue_dAdd_a_pmain_dIntPriorityQueue_aint(t0, (int)(6))
	specialized_main__pmain_dIntPriorityQueue_dAdd_a_pmain_dIntPriorityQueue_aint(t0, (int)(5))
	specialized_main__pmain_dIntPriorityQueue_dAdd_a_pmain_dIntPriorityQueue_aint(t0, (int)(2))
	specialized_main__pmain_dIntPriorityQueue_dAdd_a_pmain_dIntPriorityQueue_aint(t0, (int)(4))
	goto b3

b1: // for.body
	t7 = os.Stdout
	t8 = specialized_main__pmain_dIntPriorityQueue_dRemove_a_pmain_dIntPriorityQueue_rint(t0)
	t9 = strconv.Itoa(t8)
	t10 = t9 + (string)("\n")
	_, _ = (*os.File).WriteString(t7, t10)
	goto b3

b2: // for.done
	return

b3: // for.loop
	t12 = *t0
	t13 = (main.IntPriorityQueue).Len(t12)
	t14 = t13 != (int)(0)
	if t14 {
		goto b1
	} else {
		goto b2
	}
	// Output:
	// 1
	// 2
	// 3
	// 4
	// 5
	// 6
	//
}

func specialized_main__pmain_dIntPriorityQueue_dAdd_a_pmain_dIntPriorityQueue_aint(param_h *main.IntPriorityQueue, param_i int) {
	var (
		t0 *main.IntPriorityQueue
		t1 int
	)

	//b0: // entry
	t0 = (*main.IntPriorityQueue)(param_h)
	t1 = (int)(param_i)
	specialized_heap_Push_a_pmain_dIntPriorityQueue_aint(t0, t1)
	return
}

func specialized_heap_Push_a_pmain_dIntPriorityQueue_aint(param_h *main.IntPriorityQueue, param_x int) {
	var (
		t1 int
		t2 int
	)

	//b0: // entry
	specialized_main__pmain_dIntPriorityQueue_dPush_a_pmain_dIntPriorityQueue_aint(param_h, param_x)
	t1 = (*main.IntPriorityQueue).Len(param_h)
	t2 = t1 - (int)(1)
	specialized_heap_up_a_pmain_dIntPriorityQueue_aint(param_h, t2)
	return
}

func specialized_heap_up_a_pmain_dIntPriorityQueue_aint(param_h *main.IntPriorityQueue, param_j int) {
	var (
		t0 int
		t1 int
		t2 int
		t3 bool
		t5 bool
	)

	//b0: // entry
	t0 = param_j // j
	goto b1

b1: // for.body
	t1 = t0 - (int)(1)
	t2 = t1 / (int)(2)
	t3 = t2 == t0
	if t3 {
		goto b2
	} else {
		goto b4
	}

b2: // if.then
	return

b3: // if.done
	(*main.IntPriorityQueue).Swap(param_h, t2, t0)
	t0 = t2 // j
	goto b1

b4: // cond.false
	t5 = (*main.IntPriorityQueue).Less(param_h, t0, t2)
	if t5 {
		goto b3
	} else {
		goto b2
	}
}

func specialized_main__pmain_dIntPriorityQueue_dPush_a_pmain_dIntPriorityQueue_aint(param_h *main.IntPriorityQueue, param_v int) {
	var (
		t0 main.IntPriorityQueue
		t1 int
		t2 *[1]int
		t3 *int
		t4 []int
		t5 main.IntPriorityQueue
	)

	//b0: // entry
	t0 = *param_h
	t1 = param_v
	t2 = new([1]int) // varargs
	t3 = &t2[(int)(0)]
	*t3 = t1
	t4 = t2[:]
	t5 = append(t0, t4...)
	*param_h = t5
	return
}

func specialized_main__pmain_dIntPriorityQueue_dRemove_a_pmain_dIntPriorityQueue_rint(param_h *main.IntPriorityQueue) int {
	var (
		t0 *main.IntPriorityQueue
		t1 int
		t2 int
	)

	//b0: // entry
	t0 = (*main.IntPriorityQueue)(param_h)
	t1 = specialized_heap_Pop_a_pmain_dIntPriorityQueue_rint(t0)
	t2 = t1
	return t2
}

func specialized_heap_Pop_a_pmain_dIntPriorityQueue_rint(param_h *main.IntPriorityQueue) int {
	var (
		t0 int
		t1 int
		t4 int
	)

	//b0: // entry
	t0 = (*main.IntPriorityQueue).Len(param_h)
	t1 = t0 - (int)(1)
	(*main.IntPriorityQueue).Swap(param_h, (int)(0), t1)
	specialized_heap_down_a_pmain_dIntPriorityQueue_aint_aint(param_h, (int)(0), t1)
	t4 = specialized_main__pmain_dIntPriorityQueue_dPop_a_pmain_dIntPriorityQueue_rint(param_h)
	return t4
}

func specialized_heap_down_a_pmain_dIntPriorityQueue_aint_aint(param_h *main.IntPriorityQueue, param_i int, param_n int) {
	var (
		t0  int
		t3  int
		t4  int
		t5  bool
		t6  int
		t7  bool
		t8  bool
		t9  int
		t10 bool
		t11 bool
	)

	//b0: // entry
	t0 = param_i // i
	_ = (int)(0) // j
	_ = (int)(0) // j2
	goto b1

b1: // for.body
	t3 = (int)(2) * t0
	t4 = t3 + (int)(1)
	t5 = t4 >= param_n
	if t5 {
		goto b2
	} else {
		goto b4
	}

b2: // for.done
	return

b3: // if.done
	t6 = t4 + (int)(1)
	t7 = t6 < param_n
	if t7 {
		goto b7
	} else {
		t9 = t4 // j
		goto b6
	}

b4: // cond.false
	t8 = t4 < (int)(0)
	if t8 {
		goto b2
	} else {
		goto b3
	}

b5: // if.then
	t9 = t6 // j
	goto b6

b6: // if.done
	t10 = (*main.IntPriorityQueue).Less(param_h, t9, t0)
	if t10 {
		goto b8
	} else {
		goto b2
	}

b7: // cond.true
	t11 = (*main.IntPriorityQueue).Less(param_h, t4, t6)
	if t11 {
		t9 = t4 // j
		goto b6
	} else {
		goto b5
	}

b8: // if.done
	(*main.IntPriorityQueue).Swap(param_h, t0, t9)
	t0 = t9 // i
	_ = t9  // j
	_ = t6  // j2
	goto b1
}

func specialized_main__pmain_dIntPriorityQueue_dPop_a_pmain_dIntPriorityQueue_rint(param_h *main.IntPriorityQueue) int {
	var (
		t0 main.IntPriorityQueue
		t1 []int
		t2 int
		t3 int
		t4 main.IntPriorityQueue
		t5 main.IntPriorityQueue
		t6 main.IntPriorityQueue
		t7 *int
		t8 int
		t9 int
	)

	//b0: // entry
	t0 = *param_h
	t1 = ([]int)(t0)
	t2 = len(t1)
	t3 = t2 - (int)(1)
	t4 = *param_h
	t5 = t4[:t3]
	t6 = *param_h
	t7 = &t6[t3]
	t8 = *t7
	*param_h = t5
	t9 = (int)(t8)
	return t9
}

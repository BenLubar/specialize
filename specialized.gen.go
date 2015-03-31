//+build !no_specialized

package main

func (param_h IntPriorityQueue) Init_Specialized() {
	var (
		t0 *IntPriorityQueue
		t1 *IntPriorityQueue
	)

	//b0: // entry
	t0 = new(IntPriorityQueue) // h
	*t0 = param_h
	t1 = (*IntPriorityQueue)(t0)
	specialized_heap_Init_a_pIntPriorityQueue(t1)
	return
}

func specialized_heap_Init_a_pIntPriorityQueue(param_h *IntPriorityQueue) {
	var (
		t0 int
		t1 int
		t2 int
		t4 int
		t5 int
		t6 bool
	)

	//b0: // entry
	t0 = (*IntPriorityQueue).Len(param_h)
	t1 = t0 / (int)(2)
	t2 = t1 - (int)(1)
	t5 = t2 // i
	goto b3

b1: // for.body
	specialized_heap_down_a_pIntPriorityQueue_aint_aint(param_h, t5, t0)
	t4 = t5 - (int)(1)
	t5 = t4 // i
	goto b3

b2: // for.done
	return

b3: // for.loop
	t6 = t5 >= (int)(0)
	if t6 {
		goto b1
	} else {
		goto b2
	}
}

func specialized_heap_down_a_pIntPriorityQueue_aint_aint(param_h *IntPriorityQueue, param_i int, param_n int) {
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
	t10 = (*IntPriorityQueue).Less(param_h, t9, t0)
	if t10 {
		goto b8
	} else {
		goto b2
	}

b7: // cond.true
	t11 = (*IntPriorityQueue).Less(param_h, t4, t6)
	if t11 {
		t9 = t4 // j
		goto b6
	} else {
		goto b5
	}

b8: // if.done
	(*IntPriorityQueue).Swap(param_h, t0, t9)
	t0 = t9 // i
	_ = t9  // j
	_ = t6  // j2
	goto b1
}

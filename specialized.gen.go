//+build !no_specialized

package main

func (param_h *FuncCallHeap) Add_Specialized(param_fc *FuncCall) {
	var (
		t0 *FuncCallHeap
		t1 *FuncCall
	)

	//b0: // entry
	t0 = (*FuncCallHeap)(param_h)
	t1 = (*FuncCall)(param_fc)
	specialized_heap_Push_a_pFuncCallHeap_a_pFuncCall(t0, t1)
	return
}

func (param_h *FuncCallHeap) Init_Specialized() {
	var (
		t0 *FuncCallHeap
	)

	//b0: // entry
	t0 = (*FuncCallHeap)(param_h)
	specialized_heap_Init_a_pFuncCallHeap(t0)
	return
}

func (param_h *FuncCallHeap) Next_Specialized() *FuncCall {
	var (
		t0 *FuncCallHeap
		t1 *FuncCall
		t2 *FuncCall
	)

	//b0: // entry
	t0 = (*FuncCallHeap)(param_h)
	t1 = specialized_heap_Pop_a_pFuncCallHeap_r_pFuncCall(t0)
	t2 = t1
	return t2
}

func (param_h *IntPriorityQueue) Add_Specialized(param_i int) {
	var (
		t0 *IntPriorityQueue
		t1 int
	)

	//b0: // entry
	t0 = (*IntPriorityQueue)(param_h)
	t1 = (int)(param_i)
	specialized_heap_Push_a_pIntPriorityQueue_aint(t0, t1)
	return
}

func (param_h *IntPriorityQueue) Remove_Specialized() int {
	var (
		t0 *IntPriorityQueue
		t1 int
		t2 int
	)

	//b0: // entry
	t0 = (*IntPriorityQueue)(param_h)
	t1 = specialized_heap_Pop_a_pIntPriorityQueue_rint(t0)
	t2 = t1
	return t2
}

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

func specialized_heap_Init_a_pFuncCallHeap(param_h *FuncCallHeap) {
	var (
		t0 int
		t1 int
		t2 int
		t4 int
		t5 int
		t6 bool
	)

	//b0: // entry
	t0 = (*FuncCallHeap).Len(param_h)
	t1 = t0 / (int)(2)
	t2 = t1 - (int)(1)
	t5 = t2 // i
	goto b3

b1: // for.body
	specialized_heap_down_a_pFuncCallHeap_aint_aint(param_h, t5, t0)
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

func specialized_heap_Pop_a_pFuncCallHeap_r_pFuncCall(param_h *FuncCallHeap) *FuncCall {
	var (
		t0 int
		t1 int
		t4 *FuncCall
	)

	//b0: // entry
	t0 = (*FuncCallHeap).Len(param_h)
	t1 = t0 - (int)(1)
	(*FuncCallHeap).Swap(param_h, (int)(0), t1)
	specialized_heap_down_a_pFuncCallHeap_aint_aint(param_h, (int)(0), t1)
	t4 = specialized__l_pFuncCallHeap_dPop_a_pFuncCallHeap_r_pFuncCall(param_h)
	return t4
}

func specialized__l_pFuncCallHeap_dPop_a_pFuncCallHeap_r_pFuncCall(param_h *FuncCallHeap) *FuncCall {
	var (
		t0 FuncCallHeap
		t1 []*FuncCall
		t2 int
		t3 int
		t4 FuncCallHeap
		t5 **FuncCall
		t6 *FuncCall
		t7 FuncCallHeap
		t8 FuncCallHeap
		t9 *FuncCall
	)

	//b0: // entry
	t0 = *param_h
	t1 = ([]*FuncCall)(t0)
	t2 = len(t1)
	t3 = t2 - (int)(1)
	t4 = *param_h
	t5 = &t4[t3]
	t6 = *t5
	t7 = *param_h
	t8 = t7[:t3]
	*param_h = t8
	t9 = (*FuncCall)(t6)
	return t9
}

func specialized_heap_Pop_a_pIntPriorityQueue_rint(param_h *IntPriorityQueue) int {
	var (
		t0 int
		t1 int
		t4 int
	)

	//b0: // entry
	t0 = (*IntPriorityQueue).Len(param_h)
	t1 = t0 - (int)(1)
	(*IntPriorityQueue).Swap(param_h, (int)(0), t1)
	specialized_heap_down_a_pIntPriorityQueue_aint_aint(param_h, (int)(0), t1)
	t4 = specialized__l_pIntPriorityQueue_dPop_a_pIntPriorityQueue_rint(param_h)
	return t4
}

func specialized__l_pIntPriorityQueue_dPop_a_pIntPriorityQueue_rint(param_h *IntPriorityQueue) int {
	var (
		t0 IntPriorityQueue
		t1 []int
		t2 int
		t3 int
		t4 IntPriorityQueue
		t5 IntPriorityQueue
		t6 IntPriorityQueue
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

func specialized_heap_Push_a_pFuncCallHeap_a_pFuncCall(param_h *FuncCallHeap, param_x *FuncCall) {
	var (
		t1 int
		t2 int
	)

	//b0: // entry
	specialized__l_pFuncCallHeap_dPush_a_pFuncCallHeap_a_pFuncCall(param_h, param_x)
	t1 = (*FuncCallHeap).Len(param_h)
	t2 = t1 - (int)(1)
	specialized_heap_up_a_pFuncCallHeap_aint(param_h, t2)
	return
}

func specialized__l_pFuncCallHeap_dPush_a_pFuncCallHeap_a_pFuncCall(param_h *FuncCallHeap, param_x *FuncCall) {
	var (
		t0 FuncCallHeap
		t1 *FuncCall
		t2 *[1]*FuncCall
		t3 **FuncCall
		t4 []*FuncCall
		t5 FuncCallHeap
	)

	//b0: // entry
	t0 = *param_h
	t1 = param_x
	t2 = new([1]*FuncCall) // varargs
	t3 = &t2[(int)(0)]
	*t3 = t1
	t4 = t2[:]
	t5 = append(t0, t4...)
	*param_h = t5
	return
}

func specialized_heap_Push_a_pIntPriorityQueue_aint(param_h *IntPriorityQueue, param_x int) {
	var (
		t1 int
		t2 int
	)

	//b0: // entry
	specialized__l_pIntPriorityQueue_dPush_a_pIntPriorityQueue_aint(param_h, param_x)
	t1 = (*IntPriorityQueue).Len(param_h)
	t2 = t1 - (int)(1)
	specialized_heap_up_a_pIntPriorityQueue_aint(param_h, t2)
	return
}

func specialized__l_pIntPriorityQueue_dPush_a_pIntPriorityQueue_aint(param_h *IntPriorityQueue, param_v int) {
	var (
		t0 IntPriorityQueue
		t1 int
		t2 *[1]int
		t3 *int
		t4 []int
		t5 IntPriorityQueue
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

func specialized_heap_down_a_pFuncCallHeap_aint_aint(param_h *FuncCallHeap, param_i int, param_n int) {
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
	t10 = (*FuncCallHeap).Less(param_h, t9, t0)
	if t10 {
		goto b8
	} else {
		goto b2
	}

b7: // cond.true
	t11 = (*FuncCallHeap).Less(param_h, t4, t6)
	if t11 {
		t9 = t4 // j
		goto b6
	} else {
		goto b5
	}

b8: // if.done
	(*FuncCallHeap).Swap(param_h, t0, t9)
	t0 = t9 // i
	_ = t9  // j
	_ = t6  // j2
	goto b1
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

func specialized_heap_up_a_pFuncCallHeap_aint(param_h *FuncCallHeap, param_j int) {
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
	(*FuncCallHeap).Swap(param_h, t2, t0)
	t0 = t2 // j
	goto b1

b4: // cond.false
	t5 = (*FuncCallHeap).Less(param_h, t0, t2)
	if t5 {
		goto b3
	} else {
		goto b2
	}
}

func specialized_heap_up_a_pIntPriorityQueue_aint(param_h *IntPriorityQueue, param_j int) {
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
	(*IntPriorityQueue).Swap(param_h, t2, t0)
	t0 = t2 // j
	goto b1

b4: // cond.false
	t5 = (*IntPriorityQueue).Less(param_h, t0, t2)
	if t5 {
		goto b3
	} else {
		goto b2
	}
}

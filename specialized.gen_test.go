//+build !no_specialized

package main_test

import "testing"
import "github.com/BenLubar/specialize"

func BenchmarkHeapInit100_Specialized(param_b *testing.B) {
	var (
		t0 []int
	)

	//b0: // entry
	t0 = heapData100
	specialized__lbenchmarkHeapInit_a_ptesting_dB_a_s_eint(param_b, t0)
	return
}

func BenchmarkHeapInit10K_Specialized(param_b *testing.B) {
	var (
		t0 []int
	)

	//b0: // entry
	t0 = heapData10K
	specialized__lbenchmarkHeapInit_a_ptesting_dB_a_s_eint(param_b, t0)
	return
}

func BenchmarkHeapInit1M_Specialized(param_b *testing.B) {
	var (
		t0 []int
	)

	//b0: // entry
	t0 = heapData1M
	specialized__lbenchmarkHeapInit_a_ptesting_dB_a_s_eint(param_b, t0)
	return
}

func BenchmarkHeapPop100_Specialized(param_b *testing.B) {
	var (
		t0 []int
	)

	//b0: // entry
	t0 = heapData100
	specialized__lbenchmarkHeapPop_a_ptesting_dB_a_s_eint(param_b, t0)
	return
}

func BenchmarkHeapPop10K_Specialized(param_b *testing.B) {
	var (
		t0 []int
	)

	//b0: // entry
	t0 = heapData10K
	specialized__lbenchmarkHeapPop_a_ptesting_dB_a_s_eint(param_b, t0)
	return
}

func BenchmarkHeapPop1M_Specialized(param_b *testing.B) {
	var (
		t0 []int
	)

	//b0: // entry
	t0 = heapData1M
	specialized__lbenchmarkHeapPop_a_ptesting_dB_a_s_eint(param_b, t0)
	return
}

func BenchmarkHeapPush100_Specialized(param_b *testing.B) {
	var (
		t0 []int
	)

	//b0: // entry
	t0 = heapData100
	specialized__lbenchmarkHeapPush_a_ptesting_dB_a_s_eint(param_b, t0)
	return
}

func BenchmarkHeapPush10K_Specialized(param_b *testing.B) {
	var (
		t0 []int
	)

	//b0: // entry
	t0 = heapData10K
	specialized__lbenchmarkHeapPush_a_ptesting_dB_a_s_eint(param_b, t0)
	return
}

func BenchmarkHeapPush1M_Specialized(param_b *testing.B) {
	var (
		t0 []int
	)

	//b0: // entry
	t0 = heapData1M
	specialized__lbenchmarkHeapPush_a_ptesting_dB_a_s_eint(param_b, t0)
	return
}

func benchmarkHeapInit_Specialized(param_b *testing.B, param_data []int) {
	var (
		t0 int
		t1 main.IntSlice
		t5 int
		t6 int
		t7 *int
		t8 int
		t9 bool
	)

	//b0: // entry
	t0 = len(param_data)
	t1 = make(main.IntSlice, t0, t0)
	(*testing.B).ResetTimer(param_b)
	t6 = (int)(0) // i
	goto b3

b1: // for.body
	_ = copy(t1, param_data)
	specialized_main_main_dIntSlice_dInit_amain_dIntSlice(t1)
	t5 = t6 + (int)(1)
	t6 = t5 // i
	goto b3

b2: // for.done
	return

b3: // for.loop
	t7 = &param_b.N
	t8 = *t7
	t9 = t6 < t8
	if t9 {
		goto b1
	} else {
		goto b2
	}
}

func benchmarkHeapPop_Specialized(param_b *testing.B, param_data []int) {
	var (
		t0  int
		t1  main.IntSlice
		t4  *main.IntSlice
		t5  int
		t6  main.IntSlice
		t8  main.IntSlice
		t9  int
		t10 main.IntSlice
		t11 main.IntSlice
		t13 int
		t14 *int
		t15 int
		t16 bool
		t18 int
		t19 main.IntSlice
		t20 []int
		t21 int
		t22 bool
	)

	//b0: // entry
	t0 = len(param_data)
	t1 = make(main.IntSlice, t0, t0)
	_ = copy(t1, param_data)
	specialized_main_main_dIntSlice_dInit_amain_dIntSlice(t1)
	t4 = new(main.IntSlice) // h2
	t5 = len(param_data)
	t6 = make(main.IntSlice, (int)(0), t5)
	*t4 = t6
	(*testing.B).ResetTimer(param_b)
	t13 = (int)(0) // i
	goto b3

b1: // for.body
	t8 = *t4
	t9 = len(param_data)
	t10 = t8[:t9]
	*t4 = t10
	t11 = *t4
	_ = copy(t11, t1)
	goto b6

b2: // for.done
	return

b3: // for.loop
	t14 = &param_b.N
	t15 = *t14
	t16 = t13 < t15
	if t16 {
		goto b1
	} else {
		goto b2
	}

b4: // for.body
	_ = specialized_main__pmain_dIntSlice_dRemove_a_pmain_dIntSlice_rint(t4)
	goto b6

b5: // for.done
	t18 = t13 + (int)(1)
	t13 = t18 // i
	goto b3

b6: // for.loop
	t19 = *t4
	t20 = ([]int)(t19)
	t21 = len(t20)
	t22 = t21 != (int)(0)
	if t22 {
		goto b4
	} else {
		goto b5
	}
}

func benchmarkHeapPush_Specialized(param_b *testing.B, param_data []int) {
	var (
		t0  *main.IntSlice
		t1  int
		t2  main.IntSlice
		t4  main.IntSlice
		t5  main.IntSlice
		t6  int
		t7  int
		t8  *int
		t9  int
		t10 bool
		t11 int
		t12 int
		t13 bool
		t14 *int
		t15 int
		t17 int
	)

	//b0: // entry
	t0 = new(main.IntSlice) // h
	t1 = len(param_data)
	t2 = make(main.IntSlice, t1, t1)
	*t0 = t2
	(*testing.B).ResetTimer(param_b)
	t7 = (int)(0) // i
	goto b3

b1: // for.body
	t4 = *t0
	t5 = t4[:(int)(0)]
	*t0 = t5
	t6 = len(param_data)
	t11 = (int)(-1)
	goto b4

b2: // for.done
	return

b3: // for.loop
	t8 = &param_b.N
	t9 = *t8
	t10 = t7 < t9
	if t10 {
		goto b1
	} else {
		goto b2
	}

b4: // rangeindex.loop
	t12 = t11 + (int)(1)
	t13 = t12 < t6
	if t13 {
		goto b5
	} else {
		goto b6
	}

b5: // rangeindex.body
	t14 = &param_data[t12]
	t15 = *t14
	specialized_main__pmain_dIntSlice_dAdd_a_pmain_dIntSlice_aint(t0, t15)
	t11 = t12
	goto b4

b6: // rangeindex.done
	t17 = t7 + (int)(1)
	t7 = t17 // i
	goto b3
}

func specialized__lbenchmarkHeapInit_a_ptesting_dB_a_s_eint(param_b *testing.B, param_data []int) {
	var (
		t0 int
		t1 main.IntSlice
		t5 int
		t6 int
		t7 *int
		t8 int
		t9 bool
	)

	//b0: // entry
	t0 = len(param_data)
	t1 = make(main.IntSlice, t0, t0)
	(*testing.B).ResetTimer(param_b)
	t6 = (int)(0) // i
	goto b3

b1: // for.body
	_ = copy(t1, param_data)
	specialized_main_main_dIntSlice_dInit_amain_dIntSlice(t1)
	t5 = t6 + (int)(1)
	t6 = t5 // i
	goto b3

b2: // for.done
	return

b3: // for.loop
	t7 = &param_b.N
	t8 = *t7
	t9 = t6 < t8
	if t9 {
		goto b1
	} else {
		goto b2
	}
}

func specialized__lbenchmarkHeapPop_a_ptesting_dB_a_s_eint(param_b *testing.B, param_data []int) {
	var (
		t0  int
		t1  main.IntSlice
		t4  *main.IntSlice
		t5  int
		t6  main.IntSlice
		t8  main.IntSlice
		t9  int
		t10 main.IntSlice
		t11 main.IntSlice
		t13 int
		t14 *int
		t15 int
		t16 bool
		t18 int
		t19 main.IntSlice
		t20 []int
		t21 int
		t22 bool
	)

	//b0: // entry
	t0 = len(param_data)
	t1 = make(main.IntSlice, t0, t0)
	_ = copy(t1, param_data)
	specialized_main_main_dIntSlice_dInit_amain_dIntSlice(t1)
	t4 = new(main.IntSlice) // h2
	t5 = len(param_data)
	t6 = make(main.IntSlice, (int)(0), t5)
	*t4 = t6
	(*testing.B).ResetTimer(param_b)
	t13 = (int)(0) // i
	goto b3

b1: // for.body
	t8 = *t4
	t9 = len(param_data)
	t10 = t8[:t9]
	*t4 = t10
	t11 = *t4
	_ = copy(t11, t1)
	goto b6

b2: // for.done
	return

b3: // for.loop
	t14 = &param_b.N
	t15 = *t14
	t16 = t13 < t15
	if t16 {
		goto b1
	} else {
		goto b2
	}

b4: // for.body
	_ = specialized_main__pmain_dIntSlice_dRemove_a_pmain_dIntSlice_rint(t4)
	goto b6

b5: // for.done
	t18 = t13 + (int)(1)
	t13 = t18 // i
	goto b3

b6: // for.loop
	t19 = *t4
	t20 = ([]int)(t19)
	t21 = len(t20)
	t22 = t21 != (int)(0)
	if t22 {
		goto b4
	} else {
		goto b5
	}
}

func specialized__lbenchmarkHeapPush_a_ptesting_dB_a_s_eint(param_b *testing.B, param_data []int) {
	var (
		t0  *main.IntSlice
		t1  int
		t2  main.IntSlice
		t4  main.IntSlice
		t5  main.IntSlice
		t6  int
		t7  int
		t8  *int
		t9  int
		t10 bool
		t11 int
		t12 int
		t13 bool
		t14 *int
		t15 int
		t17 int
	)

	//b0: // entry
	t0 = new(main.IntSlice) // h
	t1 = len(param_data)
	t2 = make(main.IntSlice, t1, t1)
	*t0 = t2
	(*testing.B).ResetTimer(param_b)
	t7 = (int)(0) // i
	goto b3

b1: // for.body
	t4 = *t0
	t5 = t4[:(int)(0)]
	*t0 = t5
	t6 = len(param_data)
	t11 = (int)(-1)
	goto b4

b2: // for.done
	return

b3: // for.loop
	t8 = &param_b.N
	t9 = *t8
	t10 = t7 < t9
	if t10 {
		goto b1
	} else {
		goto b2
	}

b4: // rangeindex.loop
	t12 = t11 + (int)(1)
	t13 = t12 < t6
	if t13 {
		goto b5
	} else {
		goto b6
	}

b5: // rangeindex.body
	t14 = &param_data[t12]
	t15 = *t14
	specialized_main__pmain_dIntSlice_dAdd_a_pmain_dIntSlice_aint(t0, t15)
	t11 = t12
	goto b4

b6: // rangeindex.done
	t17 = t7 + (int)(1)
	t7 = t17 // i
	goto b3
}

func specialized_main__pmain_dIntSlice_dAdd_a_pmain_dIntSlice_aint(param_s *main.IntSlice, param_i int) {
	var (
		t0 *main.IntSlice
		t1 int
	)

	//b0: // entry
	t0 = (*main.IntSlice)(param_s)
	t1 = (int)(param_i)
	specialized_heap_Push_a_pmain_dIntSlice_aint(t0, t1)
	return
}

func specialized_heap_Push_a_pmain_dIntSlice_aint(param_h *main.IntSlice, param_x int) {
	var (
		t1 int
		t2 int
	)

	//b0: // entry
	specialized_main__pmain_dIntSlice_dPush_a_pmain_dIntSlice_aint(param_h, param_x)
	t1 = (*main.IntSlice).Len(param_h)
	t2 = t1 - (int)(1)
	specialized_heap_up_a_pmain_dIntSlice_aint(param_h, t2)
	return
}

func specialized_heap_up_a_pmain_dIntSlice_aint(param_h *main.IntSlice, param_j int) {
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
	(*main.IntSlice).Swap(param_h, t2, t0)
	t0 = t2 // j
	goto b1

b4: // cond.false
	t5 = (*main.IntSlice).Less(param_h, t0, t2)
	if t5 {
		goto b3
	} else {
		goto b2
	}
}

func specialized_main__pmain_dIntSlice_dPush_a_pmain_dIntSlice_aint(param_s *main.IntSlice, param_v int) {
	var (
		t0 main.IntSlice
		t1 int
		t2 *[1]int
		t3 *int
		t4 []int
		t5 main.IntSlice
	)

	//b0: // entry
	t0 = *param_s
	t1 = param_v
	t2 = new([1]int) // varargs
	t3 = &t2[(int)(0)]
	*t3 = t1
	t4 = t2[:]
	t5 = append(t0, t4...)
	*param_s = t5
	return
}

func specialized_main__pmain_dIntSlice_dRemove_a_pmain_dIntSlice_rint(param_s *main.IntSlice) int {
	var (
		t0 *main.IntSlice
		t1 int
		t2 int
	)

	//b0: // entry
	t0 = (*main.IntSlice)(param_s)
	t1 = specialized_heap_Pop_a_pmain_dIntSlice_rint(t0)
	t2 = t1
	return t2
}

func specialized_heap_Pop_a_pmain_dIntSlice_rint(param_h *main.IntSlice) int {
	var (
		t0 int
		t1 int
		t4 int
	)

	//b0: // entry
	t0 = (*main.IntSlice).Len(param_h)
	t1 = t0 - (int)(1)
	(*main.IntSlice).Swap(param_h, (int)(0), t1)
	specialized_heap_down_a_pmain_dIntSlice_aint_aint(param_h, (int)(0), t1)
	t4 = specialized_main__pmain_dIntSlice_dPop_a_pmain_dIntSlice_rint(param_h)
	return t4
}

func specialized_heap_down_a_pmain_dIntSlice_aint_aint(param_h *main.IntSlice, param_i int, param_n int) {
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
	t10 = (*main.IntSlice).Less(param_h, t9, t0)
	if t10 {
		goto b8
	} else {
		goto b2
	}

b7: // cond.true
	t11 = (*main.IntSlice).Less(param_h, t4, t6)
	if t11 {
		t9 = t4 // j
		goto b6
	} else {
		goto b5
	}

b8: // if.done
	(*main.IntSlice).Swap(param_h, t0, t9)
	t0 = t9 // i
	_ = t9  // j
	_ = t6  // j2
	goto b1
}

func specialized_main__pmain_dIntSlice_dPop_a_pmain_dIntSlice_rint(param_s *main.IntSlice) int {
	var (
		t0 main.IntSlice
		t1 []int
		t2 int
		t3 int
		t4 main.IntSlice
		t5 main.IntSlice
		t6 main.IntSlice
		t7 *int
		t8 int
		t9 int
	)

	//b0: // entry
	t0 = *param_s
	t1 = ([]int)(t0)
	t2 = len(t1)
	t3 = t2 - (int)(1)
	t4 = *param_s
	t5 = t4[:t3]
	t6 = *param_s
	t7 = &t6[t3]
	t8 = *t7
	*param_s = t5
	t9 = (int)(t8)
	return t9
}

func specialized_main_main_dIntSlice_dInit_amain_dIntSlice(param_s main.IntSlice) {
	var (
		t0 *main.IntSlice
		t1 *main.IntSlice
	)

	//b0: // entry
	t0 = new(main.IntSlice) // s
	*t0 = param_s
	t1 = (*main.IntSlice)(t0)
	specialized_heap_Init_a_pmain_dIntSlice(t1)
	return
}

func specialized_heap_Init_a_pmain_dIntSlice(param_h *main.IntSlice) {
	var (
		t0 int
		t1 int
		t2 int
		t4 int
		t5 int
		t6 bool
	)

	//b0: // entry
	t0 = (*main.IntSlice).Len(param_h)
	t1 = t0 / (int)(2)
	t2 = t1 - (int)(1)
	t5 = t2 // i
	goto b3

b1: // for.body
	specialized_heap_down_a_pmain_dIntSlice_aint_aint(param_h, t5, t0)
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

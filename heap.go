//go:generate specialize

package main

import "container/heap"

type IntPriorityQueue []int

// implement heap.Interface
func (h IntPriorityQueue) Len() int            { return len(h) }
func (h IntPriorityQueue) Less(i, j int) bool  { return h[i] < h[j] }
func (h IntPriorityQueue) Swap(i, j int)       { h[i], h[j] = h[j], h[i] }
func (h *IntPriorityQueue) Push(v interface{}) { *h = append(*h, v.(int)) }
func (h *IntPriorityQueue) Pop() (v interface{}) {
	n := len(*h) - 1
	*h, v = (*h)[:n], (*h)[n]
	return
}

// convenience methods
func (h IntPriorityQueue) Init()        { heap.Init(&h) }
func (h *IntPriorityQueue) Add(i int)   { heap.Push(h, i) }
func (h *IntPriorityQueue) Remove() int { return heap.Pop(h).(int) }

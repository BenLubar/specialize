//go:generate specialize

package main

import "container/heap"

type IntSlice []int

func (s IntSlice) Len() int            { return len(s) }
func (s IntSlice) Less(i, j int) bool  { return s[i] < s[j] }
func (s IntSlice) Swap(i, j int)       { s[i], s[j] = s[j], s[i] }
func (s *IntSlice) Push(v interface{}) { *s = append(*s, v.(int)) }
func (s *IntSlice) Pop() (v interface{}) {
	n := len(*s) - 1
	*s, v = (*s)[:n], (*s)[n]
	return
}

func (s IntSlice) Init()        { heap.Init(&s) }
func (s *IntSlice) Add(i int)   { heap.Push(s, i) }
func (s *IntSlice) Remove() int { return heap.Pop(s).(int) }

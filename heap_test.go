package main

import (
	"container/heap"
	"math/rand"
	"testing"
)

var (
	heapData100 = rand.New(rand.NewSource(0)).Perm(1e2)
	heapData10K = rand.New(rand.NewSource(0)).Perm(1e4)
	heapData1M  = rand.New(rand.NewSource(0)).Perm(1e6)
)

func benchmarkHeapPush(b *testing.B, data []int) {
	h := make(IntSlice, len(data))
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		h = h[:0]
		for _, n := range data {
			heap.Push(&h, n)
		}
	}
}

func BenchmarkHeapPush100(b *testing.B) { benchmarkHeapPush(b, heapData100) }
func BenchmarkHeapPush10K(b *testing.B) { benchmarkHeapPush(b, heapData10K) }
func BenchmarkHeapPush1M(b *testing.B)  { benchmarkHeapPush(b, heapData1M) }

func benchmarkHeapInit(b *testing.B, data []int) {
	h := make(IntSlice, len(data))
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		copy(h, data)
		heap.Init(&h)
	}
}

func BenchmarkHeapInit100(b *testing.B) { benchmarkHeapInit(b, heapData100) }
func BenchmarkHeapInit10K(b *testing.B) { benchmarkHeapInit(b, heapData10K) }
func BenchmarkHeapInit1M(b *testing.B)  { benchmarkHeapInit(b, heapData1M) }

func benchmarkHeapPop(b *testing.B, data []int) {
	h1 := make(IntSlice, len(data))
	copy(h1, data)
	heap.Init(&h1)
	h2 := make(IntSlice, 0, len(data))
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		h2 = h2[:len(data)]
		copy(h2, h1)
		for len(h2) != 0 {
			_ = heap.Pop(&h2).(int)
		}
	}
}

func BenchmarkHeapPop100(b *testing.B) { benchmarkHeapPop(b, heapData100) }
func BenchmarkHeapPop10K(b *testing.B) { benchmarkHeapPop(b, heapData10K) }
func BenchmarkHeapPop1M(b *testing.B)  { benchmarkHeapPop(b, heapData1M) }

package main_test

import (
	"math/rand"
	"sort"
	"testing"
	"testing/quick"

	"github.com/BenLubar/specialize"
)

var (
	heapData100 = rand.New(rand.NewSource(0)).Perm(1e2)
	heapData10K = rand.New(rand.NewSource(0)).Perm(1e4)
	heapData1M  = rand.New(rand.NewSource(0)).Perm(1e6)
)

func benchmarkHeapPush(b *testing.B, data []int) {
	h := make(main.IntSlice, len(data))
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		h = h[:0]
		for _, n := range data {
			h.Add(n)
		}
	}
}

func BenchmarkHeapPush100(b *testing.B) { benchmarkHeapPush(b, heapData100) }
func BenchmarkHeapPush10K(b *testing.B) { benchmarkHeapPush(b, heapData10K) }
func BenchmarkHeapPush1M(b *testing.B)  { benchmarkHeapPush(b, heapData1M) }

func benchmarkHeapInit(b *testing.B, data []int) {
	h := make(main.IntSlice, len(data))
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		copy(h, data)
		h.Init()
	}
}

func BenchmarkHeapInit100(b *testing.B) { benchmarkHeapInit(b, heapData100) }
func BenchmarkHeapInit10K(b *testing.B) { benchmarkHeapInit(b, heapData10K) }
func BenchmarkHeapInit1M(b *testing.B)  { benchmarkHeapInit(b, heapData1M) }

func benchmarkHeapPop(b *testing.B, data []int) {
	h1 := make(main.IntSlice, len(data))
	copy(h1, data)
	h1.Init()
	h2 := make(main.IntSlice, 0, len(data))
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		h2 = h2[:len(data)]
		copy(h2, h1)
		for len(h2) != 0 {
			_ = h2.Remove()
		}
	}
}

func BenchmarkHeapPop100(b *testing.B) { benchmarkHeapPop(b, heapData100) }
func BenchmarkHeapPop10K(b *testing.B) { benchmarkHeapPop(b, heapData10K) }
func BenchmarkHeapPop1M(b *testing.B)  { benchmarkHeapPop(b, heapData1M) }

func TestHeapSort(t *testing.T) {
	if err := quick.CheckEqual(func(data []int) []int {
		h := main.IntSlice(data)
		h.Init()

		out := make([]int, 0, len(data))

		for len(h) != 0 {
			out = append(out, h.Remove())
		}

		return out
	}, func(data []int) []int {
		sort.Ints(data)
		return data
	}, nil); err != nil {
		t.Error(err)
	}
}

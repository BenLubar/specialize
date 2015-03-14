package main

import (
	"container/list"
	"testing"
)

/*
var helloWorldList = makeHelloWorldList()

func makeHelloWorldList() *list.List {
	l := list.New()
	l.PushBack("Hello,")
	l.PushBack("World!")
	return l
}
*/
var mediumList = makeMediumList()

func makeMediumList() *list.List {
	l := list.New()
	for i := 0; i < 10000; i++ {
		l.PushBack(i)
	}
	return l
}

var bigList = makeBigList()

func makeBigList() *list.List {
	l := list.New()
	for i := 0; i < 1000000; i++ {
		l.PushBack(i)
	}
	return l
}

/*
func BenchmarkMakeListTiny(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = makeHelloWorldList()
	}
}
*/

func BenchmarkMakeListMedium(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = makeMediumList()
	}
}

func BenchmarkMakeListBig(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = makeBigList()
	}
}

/*
func BenchmarkIterateListTiny(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for e := helloWorldList.Front(); e != nil; e = e.Next() {
			_ = e.Value.(string)
		}
	}
}
*/

func BenchmarkIterateListMedium(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for e := mediumList.Front(); e != nil; e = e.Next() {
			_ = e.Value.(int)
		}
	}
}

func BenchmarkIterateListBig(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for e := bigList.Front(); e != nil; e = e.Next() {
			_ = e.Value.(int)
		}
	}
}

/*
func TestListOfDifferentTypes(t *testing.T) {
	l := makeHelloWorldList()
	l.PushBack(42)
	e := l.Front()
	if expected := "Hello,"; expected != e.Value {
		t.Error("Expected %#v, got %#v", expected, e.Value)
	}
	e = e.Next()
	if expected := "World!"; expected != e.Value {
		t.Error("Expected %#v, got %#v", expected, e.Value)
	}
	e = e.Next()
	if expected := 42; expected != e.Value {
		t.Error("Expected %#v, got %#v", expected, e.Value)
	}
	e = e.Next()
	if expected := (*list.Element)(nil); expected != e {
		t.Error("Expected %#v, got %#v", expected, e)
	}
}
*/

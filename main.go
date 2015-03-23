package main

import (
	"bytes"
	"container/heap"
	"fmt"
	"go/format"
	"log"
	"os"
	"strconv"
	"strings"
	"unicode"

	"golang.org/x/tools/go/callgraph"
	"golang.org/x/tools/go/callgraph/rta"
	"golang.org/x/tools/go/loader"
	"golang.org/x/tools/go/ssa"
	"golang.org/x/tools/go/types"
)

var iprog *loader.Program

func main() {
	log.SetFlags(0)
	log.SetPrefix("specialize: ")

	var conf loader.Config

	conf.ImportWithTests(".")

	var err error
	iprog, err = conf.Load()
	if err != nil {
		log.Fatal(err)
	}

	prog := ssa.Create(iprog, ssa.SanityCheckFunctions)
	prog.BuildAll()

	var functions []*ssa.Function

	for _, ipkg := range iprog.InitialPackages() {
		pkg := prog.Package(ipkg.Pkg)

		for _, m := range pkg.Members {
			if f, ok := m.(*ssa.Function); ok {
				functions = append(functions, f)
			}
		}
	}

	result := rta.Analyze(functions, true)
	var h FuncCallHeap
	for _, f := range functions {
		h = Analyze(f, result, h)
	}

	h.Init()

	seen := make(map[string]bool)
	for h.Len() != 0 {
		fcall := h.Next()

		if seen[fcall.Name()] {
			continue
		}
		seen[fcall.Name()] = true

		log.Println(fcall.Name())
		Rewrite(&h, fcall)
	}
}

type FuncCall struct {
	F    *ssa.Function
	Call []types.Type
	Ret  []types.Type
	name string
}

func mangle(buf []byte, s string) []byte {
	for _, r := range s {
		switch r {
		case '_':
			buf = append(buf, "__"...)
		case '*':
			buf = append(buf, "_p"...)
		case '.':
			buf = append(buf, "_d"...)
		case '[':
			buf = append(buf, "_s"...)
		case ']':
			buf = append(buf, "_e"...)
		default:
			if unicode.IsLetter(r) || unicode.IsNumber(r) {
				buf = append(buf, string(r)...)
			} else {
				buf = append(buf, "_u"...)
				for i := 7; i >= 0; i-- {
					const hex = "0123456789ABCDEF"
					buf = append(buf, hex[r>>uint(i<<2)&0xF])
				}
			}
		}
	}

	return buf
}

func (fc *FuncCall) Name() string {
	if fc.name != "" {
		return fc.name
	}

	sig := fc.F.Object().Type().(*types.Signature)

	if name := fc.F.Name(); sig.Recv() == nil && (strings.HasPrefix(name, "Test") || strings.HasPrefix(name, "Benchmark") || strings.HasPrefix(name, "Example")) {
		fc.name = name + "_Specialized"
		return fc.name
	}

	name := []byte("specialized_")
	var buf bytes.Buffer

	ipkg := iprog.InitialPackages()[0].Pkg
	if fc.F.Pkg.Object == ipkg {
		name = append(name, "_l"...)
	} else {
		name = mangle(name, fc.F.Pkg.Object.Name())
		name = append(name, '_')
	}
	if recv := sig.Recv(); recv != nil {
		buf.Reset()
		writeType(&buf, ipkg, recv.Type(), nil)
		name = mangle(name, buf.String())
		name = append(name, "_d"...)
	}
	name = mangle(name, fc.F.Name())
	for _, v := range fc.Call {
		name = append(name, "_a"...)
		buf.Reset()
		writeType(&buf, ipkg, v, nil)
		name = mangle(name, buf.String())
	}
	for _, v := range fc.Ret {
		name = append(name, "_r"...)
		buf.Reset()
		writeType(&buf, ipkg, v, nil)
		name = mangle(name, buf.String())
	}

	fc.name = string(name)
	return fc.name
}

type FuncCallHeap []*FuncCall

func (h FuncCallHeap) Len() int            { return len(h) }
func (h FuncCallHeap) Swap(i, j int)       { h[i], h[j] = h[j], h[i] }
func (h FuncCallHeap) Less(i, j int) bool  { return h[i].Name() < h[j].Name() }
func (h *FuncCallHeap) Push(x interface{}) { *h = append(*h, x.(*FuncCall)) }
func (h *FuncCallHeap) Pop() interface{} {
	n := len(*h) - 1
	x := (*h)[n]
	*h = (*h)[:n]
	return x
}

func (h *FuncCallHeap) Init()            { heap.Init(h) }
func (h *FuncCallHeap) Add(fc *FuncCall) { heap.Push(h, fc) }
func (h *FuncCallHeap) Next() *FuncCall  { return heap.Pop(h).(*FuncCall) }

func UsesReflection(n *callgraph.Node, seen map[int]bool) bool {
	if seen[n.ID] {
		return false
	}
	seen[n.ID] = true
	if n.Func.Package() != nil && n.Func.Package().Object.Path() == "reflect" {
		return true
	}
	for _, e := range n.Out {
		if UsesReflection(e.Callee, seen) {
			return true
		}
	}
	return false
}

func Analyze(f *ssa.Function, result *rta.Result, h FuncCallHeap) FuncCallHeap {
	for _, b := range f.Blocks {
		for _, instr := range b.Instrs {
			if cc, ok := instr.(ssa.CallInstruction); ok {
				c := cc.Common()
				cf := c.StaticCallee()
				if cf == nil {
					continue
				}
				if UsesReflection(result.CallGraph.Nodes[cf], make(map[int]bool)) {
					continue
				}
				anyInterface := false
				ret := c.Signature().Results()
				call := &FuncCall{
					F:    cf,
					Call: make([]types.Type, len(c.Args)),
				}
				for i, v := range c.Args {
					call.Call[i] = v.Type()
					if m, ok := v.(*ssa.MakeInterface); ok {
						anyInterface = true
						call.Call[i] = m.X.Type()
					}
				}
				if v, ok := instr.(ssa.Value); ok && ret.Len() == 1 {
					ref := *v.Referrers()
					if len(ref) > 0 {
						if a, ok := ref[0].(*ssa.TypeAssert); ok && len(ref) == 1 && !a.CommaOk {
							call.Ret = []types.Type{a.Type()}
						} else {
							for i, l := 0, ret.Len(); i < l; i++ {
								call.Ret = append(call.Ret, ret.At(i).Type())
							}
						}
					}
				}
				if anyInterface {
					h = append(h, call)
				}
			}
		}
	}
	return h
}

func writeName(buf *bytes.Buffer, pkg *types.Package, val ssa.Value) {
	switch v := val.(type) {
	case *ssa.Const:
		buf.WriteByte('(')
		writeType(buf, pkg, v.Type(), nil)
		buf.WriteString(")(")
		if v.Value == nil {
			buf.WriteString("nil")
		} else {
			buf.WriteString(v.Value.String())
		}
		buf.WriteByte(')')
	case *ssa.Parameter:
		buf.WriteString("param_")
		buf.WriteString(v.Name())
	case *ssa.Field:
		if shouldSkipAnonymous(v) {
			writeName(buf, pkg, v.X)
		} else {
			buf.WriteString(v.Name())
		}
	case *ssa.FieldAddr:
		if shouldSkipAnonymous(v) {
			writeName(buf, pkg, v.X)
		} else {
			buf.WriteString(v.Name())
		}
	default:
		buf.WriteString(v.Name())
	}
}

func writeGoto(buf *bytes.Buffer, pkg *types.Package, b *ssa.BasicBlock, i int) {
	for _, instr := range b.Succs[i].Instrs {
		if φ, ok := instr.(*ssa.Phi); ok {
			writeName(buf, pkg, φ)
			buf.WriteByte('=')
			for j, p := range b.Succs[i].Preds {
				if p == b {
					writeName(buf, pkg, φ.Edges[j])
					break
				}
			}
			buf.WriteByte('\n')
		} else {
			break
		}
	}
	buf.WriteString("goto b")
	buf.WriteString(strconv.Itoa(b.Succs[i].Index))
}

func field(typ types.Type, f int) *types.Var {
	switch t := typ.(type) {
	case *types.Named:
		return field(t.Underlying(), f)
	case *types.Pointer:
		return field(t.Elem(), f)
	case *types.Struct:
		return t.Field(f)
	default:
		panic("unreachable")
	}
}

func shouldSkipAnonymous(instr ssa.Instruction) bool {
	var x ssa.Value
	var f *types.Var

	switch i := instr.(type) {
	case *ssa.Field:
		x = i.X
		f = field(x.Type(), i.Field)
	case *ssa.FieldAddr:
		x = i.X
		f = field(x.Type(), i.Field)
	default:
		return false
	}

	if !f.Anonymous() {
		return false
	}

	for _, ref := range *instr.(ssa.Value).Referrers() {
		switch i := ref.(type) {
		case ssa.CallInstruction:
			c := i.Common()
			if c.Args[0] != instr.(ssa.Value) {
				return false
			}
		case *ssa.Field, *ssa.FieldAddr:
			// a.b.c is the same as a.c if b is anonymous.
		default:
			return false
		}
	}

	return true
}

func writeCall(buf *bytes.Buffer, pkg *types.Package, c *ssa.CallCommon, i ssa.Instruction, h *FuncCallHeap) {
	buf.WriteString("panic(\"specialize: TODO: writeCall\")")
}

func Rewrite(h *FuncCallHeap, fc *FuncCall) {
	fc.F.WriteTo(os.Stdout)

	pkg := iprog.InitialPackages()[0].Pkg

	var buf bytes.Buffer

	buf.WriteString("\n\nfunc ")
	buf.WriteString(fc.Name())
	buf.WriteByte('(')
	for i, p := range fc.F.Params {
		if i != 0 {
			buf.WriteByte(',')
		}
		writeName(&buf, pkg, p)
		buf.WriteByte(' ')
		writeType(&buf, pkg, fc.Call[i], nil)
	}
	buf.WriteByte(')')
	if len(fc.Ret) != 0 {
		buf.WriteString(" (")
		for i, r := range fc.Ret {
			if i != 0 {
				buf.WriteByte(',')
			}
			writeType(&buf, pkg, r, nil)
		}
		buf.WriteByte(')')
	}
	buf.WriteString(" {\n")

	buf.WriteString("var (")
	for _, b := range fc.F.Blocks {
		for _, instr := range b.Instrs {
			if v, ok := instr.(ssa.Value); ok && !shouldSkipAnonymous(instr) {
				name := v.Name()
				if t, ok := v.Type().(*types.Tuple); ok {
					if t.Len() == 0 {
						// nothing
					} else if t.Len() == 1 {
						buf.WriteByte('\n')
						buf.WriteString(name)
						buf.WriteByte(' ')
						writeType(&buf, pkg, t.At(0).Type(), nil)
					} else {
						for i, l := 0, t.Len(); i < l; i++ {
							buf.WriteByte('\n')
							buf.WriteString(name)
							buf.WriteByte('_')
							buf.WriteString(strconv.Itoa(i))
							buf.WriteByte(' ')
							writeType(&buf, pkg, t.At(i).Type(), nil)
						}
					}
				} else {
					buf.WriteByte('\n')
					buf.WriteString(name)
					buf.WriteByte(' ')
					writeType(&buf, pkg, v.Type(), nil)
				}
			}
		}
	}
	buf.WriteString("\n)\n")

	for _, b := range fc.F.Blocks {
		buf.WriteString("\nb")
		buf.WriteString(strconv.Itoa(b.Index))
		buf.WriteString(":")
		if b.Comment != "" {
			buf.WriteString(" // ")
			buf.WriteString(b.Comment)
		}
		buf.WriteByte('\n')
		for _, instr := range b.Instrs {
			if _, ok := instr.(*ssa.Phi); ok {
				// handled in writeGoto
				continue
			}

			if shouldSkipAnonymous(instr) {
				continue
			}

			if v, ok := instr.(ssa.Value); ok {
				name := v.Name()
				if t, ok := v.Type().(*types.Tuple); ok {
					if t.Len() == 0 {
						// nothing
					} else if t.Len() == 1 {
						buf.WriteString(name)
						buf.WriteByte('=')
					} else {
						for i, l := 0, t.Len(); i < l; i++ {
							if i != 0 {
								buf.WriteByte(',')
							}
							buf.WriteString(name)
							buf.WriteByte('_')
							buf.WriteString(strconv.Itoa(i))
						}
						buf.WriteByte('=')
					}
				} else {
					buf.WriteString(name)
					buf.WriteByte('=')
				}
			}

			switch i := instr.(type) {
			case *ssa.Alloc:
				fmt.Fprintf(&buf, "panic(%q)", fmt.Sprintf("unhandled type: %T", i))

			case *ssa.BinOp:
				writeName(&buf, pkg, i.X)
				buf.WriteString(i.Op.String())
				writeName(&buf, pkg, i.Y)

			case *ssa.Call:
				writeCall(&buf, pkg, i.Common(), i, h)

			case *ssa.ChangeInterface:
				buf.WriteByte('(')
				writeType(&buf, pkg, i.Type(), nil)
				buf.WriteString(")(")
				writeName(&buf, pkg, i.X)
				buf.WriteByte(')')

			case *ssa.ChangeType:
				buf.WriteByte('(')
				writeType(&buf, pkg, i.Type(), nil)
				buf.WriteString(")(")
				writeName(&buf, pkg, i.X)
				buf.WriteByte(')')

			case *ssa.Convert:
				buf.WriteByte('(')
				writeType(&buf, pkg, i.Type(), nil)
				buf.WriteString(")(")
				writeName(&buf, pkg, i.X)
				buf.WriteByte(')')

			case *ssa.Defer:
				buf.WriteString("defer ")
				writeCall(&buf, pkg, i.Common(), i, h)

			case *ssa.Extract:
				buf.WriteString(i.Tuple.Name())
				buf.WriteByte('_')
				buf.WriteString(strconv.Itoa(i.Index))

			case *ssa.Field:
				writeName(&buf, pkg, i.X)
				buf.WriteByte('.')
				f := field(i.X.Type(), i.Field)
				buf.WriteString(f.Name())

			case *ssa.FieldAddr:
				buf.WriteByte('&')
				writeName(&buf, pkg, i.X)
				buf.WriteByte('.')
				f := field(i.X.Type(), i.Field)
				buf.WriteString(f.Name())

			case *ssa.Go:
				buf.WriteString("go ")
				writeCall(&buf, pkg, i.Common(), i, h)

			case *ssa.If:
				buf.WriteString("if ")
				writeName(&buf, pkg, i.Cond)
				buf.WriteString(" {\n")
				writeGoto(&buf, pkg, b, 0)
				buf.WriteString("\n} else {\n")
				writeGoto(&buf, pkg, b, 1)
				buf.WriteString("\n}")

			case *ssa.Index:
				writeName(&buf, pkg, i.X)
				buf.WriteByte('[')
				writeName(&buf, pkg, i.Index)
				buf.WriteByte(']')

			case *ssa.IndexAddr:
				buf.WriteByte('&')
				writeName(&buf, pkg, i.X)
				buf.WriteByte('[')
				writeName(&buf, pkg, i.Index)
				buf.WriteByte(']')

			case *ssa.Jump:
				writeGoto(&buf, pkg, b, 0)

			case *ssa.Lookup:
				writeName(&buf, pkg, i.X)
				buf.WriteByte('[')
				writeName(&buf, pkg, i.Index)
				buf.WriteByte(']')

			case *ssa.MakeChan:
				buf.WriteString("make(")
				writeType(&buf, pkg, i.Type(), nil)
				buf.WriteByte(',')
				writeName(&buf, pkg, i.Size)
				buf.WriteByte(')')

			case *ssa.MakeClosure:
				panic("specialize: TODO: implement closures")

			case *ssa.MakeInterface:
				buf.WriteByte('(')
				writeType(&buf, pkg, i.Type(), nil)
				buf.WriteString(")(")
				writeName(&buf, pkg, i.X)
				buf.WriteByte(')')

			case *ssa.MakeMap:
				buf.WriteString("make(")
				writeType(&buf, pkg, i.Type(), nil)
				if i.Reserve != nil {
					buf.WriteByte(',')
					writeName(&buf, pkg, i.Reserve)
				}
				buf.WriteByte(')')

			case *ssa.MakeSlice:
				buf.WriteString("make(")
				writeType(&buf, pkg, i.Type(), nil)
				buf.WriteByte(',')
				writeName(&buf, pkg, i.Len)
				buf.WriteByte(',')
				writeName(&buf, pkg, i.Cap)
				buf.WriteByte(')')

			case *ssa.MapUpdate:
				writeName(&buf, pkg, i.Map)
				buf.WriteByte('[')
				writeName(&buf, pkg, i.Key)
				buf.WriteString("]=")
				writeName(&buf, pkg, i.Value)

			case *ssa.Next:
				panic("specialize: TODO: implement map/string iterators")

			case *ssa.Panic:
				buf.WriteString("panic(")
				writeName(&buf, pkg, i.X)
				buf.WriteByte(')')

			case *ssa.Return:
				buf.WriteString("return ")
				for i, r := range i.Results {
					if i != 0 {
						buf.WriteByte(',')
					}
					writeName(&buf, pkg, r)
				}

			case *ssa.RunDefers:
				panic("specialize: TODO: implement RunDefers")

			case *ssa.Select:
				panic("specialize: TODO: implement select")

			case *ssa.Send:
				writeName(&buf, pkg, i.Chan)
				buf.WriteString("<-")
				writeName(&buf, pkg, i.X)

			case *ssa.Slice:
				writeName(&buf, pkg, i.X)
				buf.WriteByte('[')
				if i.Low != nil {
					writeName(&buf, pkg, i.Low)
				}
				buf.WriteByte(':')
				if i.High != nil {
					writeName(&buf, pkg, i.High)
				}
				if i.Max != nil {
					buf.WriteByte(':')
					writeName(&buf, pkg, i.Max)
				}
				buf.WriteByte(']')

			case *ssa.Store:
				buf.WriteByte('*')
				writeName(&buf, pkg, i.Addr)
				buf.WriteByte('=')
				writeName(&buf, pkg, i.Val)

			case *ssa.TypeAssert:
				writeName(&buf, pkg, i.X)
				buf.WriteString(".(")
				writeType(&buf, pkg, i.AssertedType, nil)
				buf.WriteByte(')')

			case *ssa.UnOp:
				buf.WriteString(i.Op.String())
				writeName(&buf, pkg, i.X)

			default:
				panic("unreachable")
			}
			buf.WriteByte('\n')
		}
	}

	buf.WriteByte('}')

	b, err := format.Source(buf.Bytes())
	if err != nil {
		panic(err)
	}
	_, err = os.Stdout.Write(b)
	if err != nil {
		panic(err)
	}

	fmt.Println()
}

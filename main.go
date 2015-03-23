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

type Context struct {
	Pkg  *types.Package
	RTA  *rta.Result
	Heap FuncCallHeap
	Call *FuncCall
	bytes.Buffer
}

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

	for _, ipkg := range iprog.InitialPackages() {
		pkg := prog.Package(ipkg.Pkg)

		var functions []*ssa.Function

		for _, m := range pkg.Members {
			if f, ok := m.(*ssa.Function); ok {
				functions = append(functions, f)
			}
		}

		var ctx Context
		ctx.RTA = rta.Analyze(functions, true)
		ctx.Pkg = ipkg.Pkg
		fcctx.Pkg = ipkg.Pkg
		for _, f := range functions {
			ctx.Analyze(f)
		}

		ctx.Heap.Init()

		ctx.WriteString("package ")
		ctx.WriteString(ctx.Pkg.Name())
		ctx.WriteByte('\n')

		seen := make(map[string]bool)
		for ctx.Heap.Len() != 0 {
			ctx.Call = ctx.Heap.Next()

			if seen[ctx.Call.Name()] {
				continue
			}
			seen[ctx.Call.Name()] = true

			log.Println(ctx.Call.Name())
			Rewrite(&ctx)
		}

		b, err := format.Source(ctx.Bytes())
		if err != nil {
			panic(err)
		}
		_, err = os.Stdout.Write(b)
		if err != nil {
			panic(err)
		}

		fmt.Println()
	}
}

var fcctx Context

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

	if fc.F.Pkg.Object == fcctx.Pkg {
		name = append(name, "_l"...)
	} else {
		name = mangle(name, fc.F.Pkg.Object.Name())
		name = append(name, '_')
	}
	if recv := sig.Recv(); recv != nil {
		fcctx.Reset()
		fcctx.WriteType(recv.Type())
		name = mangle(name, fcctx.String())
		name = append(name, "_d"...)
	}
	name = mangle(name, fc.F.Name())
	for _, v := range fc.Call {
		name = append(name, "_a"...)
		fcctx.Reset()
		fcctx.WriteType(v)
		name = mangle(name, fcctx.String())
	}
	for _, v := range fc.Ret {
		name = append(name, "_r"...)
		fcctx.Reset()
		fcctx.WriteType(v)
		name = mangle(name, fcctx.String())
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

func usesReflection(n *callgraph.Node, seen map[int]bool) bool {
	if seen[n.ID] {
		return false
	}
	seen[n.ID] = true
	if n.Func.Package() != nil && n.Func.Package().Object.Path() == "reflect" {
		return true
	}
	for _, e := range n.Out {
		if usesReflection(e.Callee, seen) {
			return true
		}
	}
	return false
}

func (ctx *Context) Analyze(f *ssa.Function) {
	for _, b := range f.Blocks {
		for _, instr := range b.Instrs {
			if cc, ok := instr.(ssa.CallInstruction); ok {
				c := cc.Common()
				cf := c.StaticCallee()
				if cf == nil {
					continue
				}
				if usesReflection(ctx.RTA.CallGraph.Nodes[cf], make(map[int]bool)) {
					continue
				}
				if call := usesInterface(cc, cf); call != nil {
					ctx.Heap = append(ctx.Heap, call)
				}
			}
		}
	}
}

func usesInterface(cc ssa.CallInstruction, cf *ssa.Function) *FuncCall {
	c := cc.Common()

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
	if v, ok := cc.(ssa.Value); ok && ret.Len() > 0 {
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
		return call
	}
	return nil
}

func (ctx *Context) WriteNumber(n int) {
	ctx.WriteString(strconv.Itoa(n))
}

func (ctx *Context) WriteName(val ssa.Value) {
	switch v := val.(type) {
	case *ssa.Const:
		ctx.WriteByte('(')
		ctx.WriteType(v.Type())
		ctx.WriteString(")(")
		if v.Value == nil {
			ctx.WriteString("nil")
		} else {
			ctx.WriteString(v.Value.String())
		}
		ctx.WriteByte(')')
	case *ssa.Parameter:
		ctx.WriteString("param_")
		ctx.WriteString(v.Name())
	case *ssa.Field:
		if shouldSkipAnonymous(v) {
			ctx.WriteName(v.X)
		} else {
			ctx.WriteString(v.Name())
		}
	case *ssa.FieldAddr:
		if shouldSkipAnonymous(v) {
			ctx.WriteName(v.X)
		} else {
			ctx.WriteString(v.Name())
		}
	default:
		ctx.WriteString(v.Name())
	}
}

func (ctx *Context) WriteGoto(b *ssa.BasicBlock, succ int) {
	for _, instr := range b.Succs[succ].Instrs {
		if φ, ok := instr.(*ssa.Phi); ok {
			ctx.WriteName(φ)
			ctx.WriteByte('=')
			for j, p := range b.Succs[succ].Preds {
				if p == b {
					ctx.WriteName(φ.Edges[j])
					if φ.Comment != "" {
						ctx.WriteString(" // ")
						ctx.WriteString(φ.Comment)
					}
					break
				}
			}
			ctx.WriteByte('\n')
		} else {
			break
		}
	}
	ctx.WriteString("goto b")
	ctx.WriteNumber(b.Succs[succ].Index)
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

func (ctx *Context) WriteCall(cc ssa.CallInstruction) {
	c := cc.Common()

	if c.IsInvoke() {
		panic("specialize: TODO: writeCall(invoke)")
	} else {
		switch v := c.Value.(type) {
		case *ssa.Function:
			panic("specialize: TODO: writeCall(function)")

		case *ssa.MakeClosure:
			panic("specialize: TODO: implement closures")

		case *ssa.Builtin:
			panic("specialize: TODO: writeCall(builtin)")

		default:
			panic("specialize: TODO: writeCall(value)")
			_ = v
		}
	}
}

func Rewrite(ctx *Context) {
	ctx.Call.F.WriteTo(os.Stdout)

	ctx.WriteString("func ")
	ctx.WriteString(ctx.Call.Name())
	ctx.WriteByte('(')
	for i, p := range ctx.Call.F.Params {
		if i != 0 {
			ctx.WriteByte(',')
		}
		ctx.WriteName(p)
		ctx.WriteByte(' ')
		ctx.WriteType(ctx.Call.Call[i])
	}
	ctx.WriteByte(')')
	if len(ctx.Call.Ret) != 0 {
		ctx.WriteString(" (")
		for i, r := range ctx.Call.Ret {
			if i != 0 {
				ctx.WriteByte(',')
			}
			ctx.WriteType(r)
		}
		ctx.WriteByte(')')
	}
	ctx.WriteString(" {\n")

	ctx.WriteString("var (")
	for _, b := range ctx.Call.F.Blocks {
		for _, instr := range b.Instrs {
			if v, ok := instr.(ssa.Value); ok && !shouldSkipAnonymous(instr) {
				name := v.Name()
				if t, ok := v.Type().(*types.Tuple); ok {
					if t.Len() == 0 {
						// nothing
					} else if t.Len() == 1 {
						ctx.WriteByte('\n')
						ctx.WriteString(name)
						ctx.WriteByte(' ')
						ctx.WriteType(t.At(0).Type())
					} else {
						for i, l := 0, t.Len(); i < l; i++ {
							ctx.WriteByte('\n')
							ctx.WriteString(name)
							ctx.WriteByte('_')
							ctx.WriteNumber(i)
							ctx.WriteByte(' ')
							ctx.WriteType(t.At(i).Type())
						}
					}
				} else {
					ctx.WriteByte('\n')
					ctx.WriteString(name)
					ctx.WriteByte(' ')
					ctx.WriteType(v.Type())
				}
			}
		}
	}
	ctx.WriteString("\n)\n")

	for _, b := range ctx.Call.F.Blocks {
		ctx.WriteString("\nb")
		ctx.WriteNumber(b.Index)
		ctx.WriteString(":")
		if b.Comment != "" {
			ctx.WriteString(" // ")
			ctx.WriteString(b.Comment)
		}
		ctx.WriteByte('\n')
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
						ctx.WriteString(name)
						ctx.WriteByte('=')
					} else {
						for i, l := 0, t.Len(); i < l; i++ {
							if i != 0 {
								ctx.WriteByte(',')
							}
							ctx.WriteString(name)
							ctx.WriteByte('_')
							ctx.WriteNumber(i)
						}
						ctx.WriteByte('=')
					}
				} else {
					ctx.WriteString(name)
					ctx.WriteByte('=')
				}
			}

			switch i := instr.(type) {
			case *ssa.Alloc:
				ctx.WriteString("new(")
				ctx.WriteType(i.Type())
				ctx.WriteByte(')')
				if i.Comment != "" {
					ctx.WriteString(" // ")
					ctx.WriteString(i.Comment)
				}

			case *ssa.BinOp:
				ctx.WriteName(i.X)
				ctx.WriteString(i.Op.String())
				ctx.WriteName(i.Y)

			case *ssa.Call:
				ctx.WriteCall(i)

			case *ssa.ChangeInterface:
				ctx.WriteByte('(')
				ctx.WriteType(i.Type())
				ctx.WriteString(")(")
				ctx.WriteName(i.X)
				ctx.WriteByte(')')

			case *ssa.ChangeType:
				ctx.WriteByte('(')
				ctx.WriteType(i.Type())
				ctx.WriteString(")(")
				ctx.WriteName(i.X)
				ctx.WriteByte(')')

			case *ssa.Convert:
				ctx.WriteByte('(')
				ctx.WriteType(i.Type())
				ctx.WriteString(")(")
				ctx.WriteName(i.X)
				ctx.WriteByte(')')

			case *ssa.Defer:
				ctx.WriteString("defer ")
				ctx.WriteCall(i)

			case *ssa.Extract:
				ctx.WriteString(i.Tuple.Name())
				ctx.WriteByte('_')
				ctx.WriteNumber(i.Index)

			case *ssa.Field:
				ctx.WriteName(i.X)
				ctx.WriteByte('.')
				f := field(i.X.Type(), i.Field)
				ctx.WriteString(f.Name())

			case *ssa.FieldAddr:
				ctx.WriteByte('&')
				ctx.WriteName(i.X)
				ctx.WriteByte('.')
				f := field(i.X.Type(), i.Field)
				ctx.WriteString(f.Name())

			case *ssa.Go:
				ctx.WriteString("go ")
				ctx.WriteCall(i)

			case *ssa.If:
				ctx.WriteString("if ")
				ctx.WriteName(i.Cond)
				ctx.WriteString(" {\n")
				ctx.WriteGoto(b, 0)
				ctx.WriteString("\n} else {\n")
				ctx.WriteGoto(b, 1)
				ctx.WriteString("\n}")

			case *ssa.Index:
				ctx.WriteName(i.X)
				ctx.WriteByte('[')
				ctx.WriteName(i.Index)
				ctx.WriteByte(']')

			case *ssa.IndexAddr:
				ctx.WriteByte('&')
				ctx.WriteName(i.X)
				ctx.WriteByte('[')
				ctx.WriteName(i.Index)
				ctx.WriteByte(']')

			case *ssa.Jump:
				ctx.WriteGoto(b, 0)

			case *ssa.Lookup:
				ctx.WriteName(i.X)
				ctx.WriteByte('[')
				ctx.WriteName(i.Index)
				ctx.WriteByte(']')

			case *ssa.MakeChan:
				ctx.WriteString("make(")
				ctx.WriteType(i.Type())
				ctx.WriteByte(',')
				ctx.WriteName(i.Size)
				ctx.WriteByte(')')

			case *ssa.MakeClosure:
				panic("specialize: TODO: implement closures")

			case *ssa.MakeInterface:
				ctx.WriteByte('(')
				ctx.WriteType(i.Type())
				ctx.WriteString(")(")
				ctx.WriteName(i.X)
				ctx.WriteByte(')')

			case *ssa.MakeMap:
				ctx.WriteString("make(")
				ctx.WriteType(i.Type())
				if i.Reserve != nil {
					ctx.WriteByte(',')
					ctx.WriteName(i.Reserve)
				}
				ctx.WriteByte(')')

			case *ssa.MakeSlice:
				ctx.WriteString("make(")
				ctx.WriteType(i.Type())
				ctx.WriteByte(',')
				ctx.WriteName(i.Len)
				ctx.WriteByte(',')
				ctx.WriteName(i.Cap)
				ctx.WriteByte(')')

			case *ssa.MapUpdate:
				ctx.WriteName(i.Map)
				ctx.WriteByte('[')
				ctx.WriteName(i.Key)
				ctx.WriteString("]=")
				ctx.WriteName(i.Value)

			case *ssa.Next:
				panic("specialize: TODO: implement map/string iterators")

			case *ssa.Panic:
				ctx.WriteString("panic(")
				ctx.WriteName(i.X)
				ctx.WriteByte(')')

			case *ssa.Return:
				ctx.WriteString("return ")
				for i, r := range i.Results {
					if i != 0 {
						ctx.WriteByte(',')
					}
					ctx.WriteName(r)
				}

			case *ssa.RunDefers:
				panic("specialize: TODO: implement RunDefers")

			case *ssa.Select:
				panic("specialize: TODO: implement select")

			case *ssa.Send:
				ctx.WriteName(i.Chan)
				ctx.WriteString("<-")
				ctx.WriteName(i.X)

			case *ssa.Slice:
				ctx.WriteName(i.X)
				ctx.WriteByte('[')
				if i.Low != nil {
					ctx.WriteName(i.Low)
				}
				ctx.WriteByte(':')
				if i.High != nil {
					ctx.WriteName(i.High)
				}
				if i.Max != nil {
					ctx.WriteByte(':')
					ctx.WriteName(i.Max)
				}
				ctx.WriteByte(']')

			case *ssa.Store:
				ctx.WriteByte('*')
				ctx.WriteName(i.Addr)
				ctx.WriteByte('=')
				ctx.WriteName(i.Val)

			case *ssa.TypeAssert:
				ctx.WriteName(i.X)
				ctx.WriteString(".(")
				ctx.WriteType(i.AssertedType)
				ctx.WriteByte(')')

			case *ssa.UnOp:
				ctx.WriteString(i.Op.String())
				ctx.WriteName(i.X)

			default:
				panic("unreachable")
			}
			ctx.WriteByte('\n')
		}
	}

	ctx.WriteString("\n")
}

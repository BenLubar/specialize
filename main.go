package main

import (
	"bytes"
	"container/heap"
	"go/build"
	"go/format"
	"go/token"
	"io/ioutil"
	"log"
	"sort"
	"strconv"
	"unicode"

	"golang.org/x/tools/go/callgraph"
	"golang.org/x/tools/go/callgraph/rta"
	"golang.org/x/tools/go/loader"
	"golang.org/x/tools/go/ssa"
	"golang.org/x/tools/go/types"
)

var iprog *loader.Program

type Context struct {
	SSA     *ssa.Program
	Pkg     *types.Package
	RTA     *rta.Result
	Heap    FuncCallHeap
	Call    *FuncCall
	Type    map[ssa.Value]types.Type
	Imports map[string]bool
	bytes.Buffer
}

func main() {
	log.SetFlags(0)
	log.SetPrefix("specialize: ")

	var conf loader.Config
	bctx := build.Default
	bctx.BuildTags = append(bctx.BuildTags, "no_specialized")
	conf.Build = &bctx

	conf.ImportWithTests(".")

	var err error
	iprog, err = conf.Load()
	if err != nil {
		log.Fatal(err)
	}

	prog := ssa.Create(iprog, ssa.SanityCheckFunctions)
	prog.BuildAll()

	packages := PackageInfoSlice(iprog.InitialPackages())
	packages.Sort()

	for i, ipkg := range packages {
		pkg := prog.Package(ipkg.Pkg)

		var functions []*ssa.Function

		for _, m := range pkg.Members {
			if f, ok := m.(*ssa.Function); ok {
				functions = append(functions, f)
			}
			if t, ok := m.(*ssa.Type); ok {
				ms := prog.MethodSets.MethodSet(t.Type())
				for i, l := 0, ms.Len(); i < l; i++ {
					functions = append(functions, prog.Method(ms.At(i)))
				}
			}
		}

		var ctx Context
		ctx.SSA = prog
		ctx.RTA = rta.Analyze(functions, true)
		ctx.Pkg = ipkg.Pkg
		ctx.Imports = make(map[string]bool)
		fcctx.Pkg = ipkg.Pkg

		for _, f := range functions {
			if call := ctx.shouldRewrite(nil, f, make(map[ssa.CallInstruction]bool)); call != nil {
				ctx.Heap = append(ctx.Heap, call)
			}
		}
		ctx.Heap.Init()

		seen := make(map[string]bool)
		for ctx.Heap.Len() != 0 {
			ctx.Call = ctx.Heap.Next()

			if seen[ctx.Call.Name()] {
				continue
			}
			seen[ctx.Call.Name()] = true

			//log.Println(ctx.Call.Name())
			Rewrite(&ctx)
		}

		b := []byte("//+build !no_specialized\n\npackage ")
		b = append(b, ctx.Pkg.Name()...)
		for path := range ctx.Imports {
			b = append(b, "\nimport "...)
			b = strconv.AppendQuote(b, path)
		}
		b = append(b, '\n')
		b = append(b, ctx.Bytes()...)

		orig := b
		b, err = format.Source(b)
		if err != nil {
			log.Println("error formatting output:", err)
			b = orig
		}
		switch i {
		case 0:
			err = ioutil.WriteFile("specialized.gen.go", b, 0644)
		case 1:
			err = ioutil.WriteFile("specialized.gen_test.go", b, 0644)
		default:
			panic("internal error: we have three packages somehow?")
		}
		if err != nil {
			panic(err)
		}
	}
}

type PackageInfoSlice []*loader.PackageInfo

func (s PackageInfoSlice) Sort()              { sort.Sort(s) }
func (s PackageInfoSlice) Len() int           { return len(s) }
func (s PackageInfoSlice) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s PackageInfoSlice) Less(i, j int) bool { return s[i].Pkg.Path() < s[j].Pkg.Path() }

var fcctx Context

type FuncCall struct {
	F         *ssa.Function
	Call      []types.Type
	Ret       []types.Type
	Unmangled bool
	name      string
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
	sig := fc.F.Signature

	if fc.name != "" {
		return fc.name
	}

	if fc.Unmangled {
		fc.name = fc.F.Name()
		if recv := sig.Recv(); recv != nil {
			fcctx.Reset()
			fcctx.WriteType(recv.Type())
			fc.name = "(" + fcctx.String() + ")." + fc.name
		}
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
	if n == nil || seen[n.ID] {
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

func (ctx *Context) shouldRewrite(cc ssa.CallInstruction, cf *ssa.Function, seen map[ssa.CallInstruction]bool) *FuncCall {
	if seen[cc] {
		return nil
	}
	if usesReflection(ctx.RTA.CallGraph.Nodes[cf], make(map[int]bool)) {
		//log.Println("ignoring because it uses reflection:", cf)
		return nil
	}
	for _, b := range cf.Blocks {
		for _, instr := range b.Instrs {
			var f *types.Var
			switch i := instr.(type) {
			case *ssa.Field:
				f = field(ctx.TypeOf(i.X), i.Field)
			case *ssa.FieldAddr:
				f = field(ctx.TypeOf(i.X), i.Field)
			}

			if f != nil && f.Pkg() != ctx.Pkg && !f.Exported() {
				//log.Println("ignoring due to unexported field access:", cf)
				return nil // we can't do anything if there's an unexported field access.
			}

			var g *ssa.Global
			switch i := instr.(type) {
			case *ssa.UnOp:
				g, _ = i.X.(*ssa.Global)
			case *ssa.Store:
				g, _ = i.Addr.(*ssa.Global)
			}
			if g != nil && g.Object().Pkg() != ctx.Pkg && !g.Object().Exported() {
				//log.Println("ignoring due to unexported global access:", cf)
				return nil
			}
		}
	}

	seen[cc] = true

	anyInterface := false
	for _, b := range cf.Blocks {
		for _, instr := range b.Instrs {
			if ccc, ok := instr.(ssa.CallInstruction); ok {
				if f := ccc.Common().StaticCallee(); f != nil && ctx.shouldRewrite(ccc, f, seen) != nil {
					anyInterface = true
					break
				}
			}
		}
		if anyInterface {
			break
		}
	}

	if cc == nil {
		if anyInterface {
			return &FuncCall{F: cf, Unmangled: true}
		}
		//log.Println("ignoring because there is nothing to rewrite:", cf)
		return nil
	}

	c := cc.Common()

	if c.Signature().Variadic() {
		// TODO(BenLubar): also support rewriting variadic functions
		//log.Println("ignoring variadic function:", cf)
		return nil
	}

	ret := c.Signature().Results()

	call := &FuncCall{
		F: cf,
	}
	if c.IsInvoke() {
		call.Call = append(call.Call, ctx.TypeOf(c.Value))
	}
	for _, v := range c.Args {
		t := ctx.TypeOf(v)
		if m, ok := v.(*ssa.MakeInterface); ok {
			t = ctx.TypeOf(m.X)
			anyInterface = true
		} else if v.Type() != t {
			anyInterface = true
		}
		call.Call = append(call.Call, t)
	}
	if v, ok := cc.(ssa.Value); ok && ret.Len() > 0 {
		if t := ctx.TypeOf(v); ret.Len() == 1 && !types.Identical(t, v.Type()) {
			call.Ret = []types.Type{t}
			anyInterface = true
		} else {
			ref := *v.Referrers()
			if len(ref) > 0 {
				if a, ok := ref[0].(*ssa.TypeAssert); ok && len(ref) == 1 && !a.CommaOk {
					call.Ret = []types.Type{ctx.TypeOf(a)}
					anyInterface = true
				}
			}
			if call.Ret == nil {
				for i, l := 0, ret.Len(); i < l; i++ {
					call.Ret = append(call.Ret, ret.At(i).Type())
				}
			}
		}
	}

	if anyInterface {
		return call
	}
	//log.Println("ignoring non-interface function:", cf)
	return nil
}

func (ctx *Context) WriteNumber(n int) {
	ctx.WriteString(strconv.Itoa(n))
}

func (ctx *Context) WriteName(val ssa.Value) {
	switch v := val.(type) {
	case *ssa.Const:
		ctx.WriteByte('(')
		ctx.WriteTypeOf(v)
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
	case *ssa.Global:
		if pkg := v.Object().Pkg(); pkg != ctx.Pkg {
			ctx.Imports[pkg.Path()] = true
			ctx.WriteString(pkg.Name())
			ctx.WriteByte('.')
		}
		ctx.WriteString(v.Name())
	default:
		ctx.WriteString(v.Name())
	}
}

func (ctx *Context) WriteGoto(b *ssa.BasicBlock, succ int) {
	for _, instr := range b.Succs[succ].Instrs {
		if φ, ok := instr.(*ssa.Phi); ok {
			if len(*φ.Referrers()) == 0 {
				ctx.WriteByte('_')
			} else {
				ctx.WriteName(φ)
			}
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

func (ctx *Context) WriteCall(cc ssa.CallInstruction) (specialized bool) {
	c := cc.Common()

	concrete := func(v *ssa.Function) {
		if call := ctx.shouldRewrite(cc, v, make(map[ssa.CallInstruction]bool)); call != nil {
			ctx.Heap.Add(call)
			ctx.WriteString(call.Name())
			specialized = true
			return
		}

		if v.Parent() != nil {
			panic("specialize: TODO: WriteCall(inline)")
		} else if recv := v.Signature.Recv(); recv != nil {
			ctx.WriteByte('(')
			ctx.WriteType(recv.Type())
			ctx.WriteString(").")
		} else if pkg := v.Package().Object; pkg != ctx.Pkg {
			ctx.Imports[pkg.Path()] = true
			ctx.WriteString(pkg.Name())
			ctx.WriteByte('.')
		}

		ctx.WriteString(v.Name())
	}

	if c.IsInvoke() {
		v := ctx.SSA.LookupMethod(ctx.TypeOf(c.Value), ctx.Pkg, c.Method.Name())
		if v != nil {
			concrete(v)
		} else {
			ctx.WriteName(c.Value)
			ctx.WriteByte('.')
			ctx.WriteString(c.Method.Name())
		}
	} else {
		switch v := c.Value.(type) {
		case *ssa.Function:
			concrete(v)

		case *ssa.MakeClosure:
			panic("specialize: TODO: implement closures")

		case *ssa.Builtin:
			ctx.WriteString(v.Name())

		default:
			panic("specialize: TODO: WriteCall(value)")
		}
	}
	ctx.WriteByte('(')
	if c.IsInvoke() {
		ctx.WriteName(c.Value)
	}
	for i, p := range c.Args {
		if i != 0 || c.IsInvoke() {
			ctx.WriteByte(',')
		}
		ctx.WriteName(p)
	}
	if c.Signature().Variadic() {
		ctx.WriteString("...")
	}
	ctx.WriteByte(')')
	return
}

func (ctx *Context) TypeOf(v ssa.Value) types.Type {
	if t, ok := ctx.Type[v]; ok {
		return t
	}
	return v.Type()
}

func (ctx *Context) WriteTypeOf(v ssa.Value) {
	ctx.WriteType(ctx.TypeOf(v))
}

func Rewrite(ctx *Context) {
	ctx.Type = make(map[ssa.Value]types.Type)

	if !ctx.Call.Unmangled {
		for i, p := range ctx.Call.F.Params {
			ctx.Type[p] = ctx.Call.Call[i]
		}

		for _, b := range ctx.Call.F.Blocks {
			if v, ok := b.Instrs[len(b.Instrs)-1].(*ssa.Return); ok {
				for i, r := range v.Results {
					ctx.Type[r] = ctx.Call.Ret[i]
				}
			}
		}
	}

	for {
		changed := false

		for _, b := range ctx.Call.F.Blocks {
			for _, instr := range b.Instrs {
				if cc, ok := instr.(ssa.CallInstruction); ok {
					c := cc.Common()
					cf := c.StaticCallee()
					if c.IsInvoke() {
						cf = ctx.SSA.LookupMethod(ctx.TypeOf(c.Value), ctx.Pkg, c.Method.Name())
					}
					var fc *FuncCall
					if cf != nil {
						fc = ctx.shouldRewrite(cc, cf, make(map[ssa.CallInstruction]bool))
					}

					if fc != nil {
						call := fc.Call
						if c.IsInvoke() {
							if !types.Identical(ctx.TypeOf(c.Value), call[0]) {
								ctx.Type[c.Value] = call[0]
								changed = true
							}
							call = call[1:]
						}
						for i, p := range call {
							if !types.Identical(ctx.TypeOf(c.Args[i]), p) {
								ctx.Type[c.Args[i]] = p
								changed = true
							}
						}
					}
				}
			}
		}

		if !changed {
			break
		}
	}

	ctx.WriteString("func ")
	if ctx.Call.Unmangled {
		params := ctx.Call.F.Params
		if ctx.Call.F.Signature.Recv() != nil {
			ctx.WriteByte('(')
			ctx.WriteName(params[0])
			ctx.WriteByte(' ')
			ctx.WriteTypeOf(params[0])
			ctx.WriteByte(')')
			params = params[1:]
		}
		ctx.WriteString(ctx.Call.F.Name())
		ctx.WriteString("_Specialized(")
		for i, p := range params {
			if i != 0 {
				ctx.WriteByte(',')
			}
			ctx.WriteName(p)
			ctx.WriteByte(' ')
			ctx.WriteTypeOf(p)
		}
		ctx.WriteByte(')')
		if ctx.Call.F.Signature.Results().Len() != 0 {
			ctx.WriteType(ctx.Call.F.Signature.Results())
		}
	} else {
		ctx.WriteString(ctx.Call.Name())
		ctx.WriteByte('(')
		for i, p := range ctx.Call.F.Params {
			if i != 0 {
				ctx.WriteByte(',')
			}
			ctx.WriteName(p)
			ctx.WriteByte(' ')
			ctx.WriteTypeOf(p)
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
	}
	ctx.WriteString(" {\n")

	ctx.WriteString("var (")
	for _, b := range ctx.Call.F.Blocks {
		for _, instr := range b.Instrs {
			if v, ok := instr.(ssa.Value); ok {
				name := v.Name()
				if len(*v.Referrers()) == 0 {
					// never used
					continue
				}
				if t, ok := ctx.TypeOf(v).(*types.Tuple); ok {
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
					ctx.WriteTypeOf(v)
				}
			}
		}
	}
	ctx.WriteString("\n)\n")

	for _, b := range ctx.Call.F.Blocks {
		ctx.WriteByte('\n')
		if len(b.Preds) == 0 {
			ctx.WriteString("//")
		}
		ctx.WriteByte('b')
		ctx.WriteNumber(b.Index)
		ctx.WriteByte(':')
		if b.Comment != "" {
			ctx.WriteString(" // ")
			ctx.WriteString(b.Comment)
		}
		ctx.WriteByte('\n')
		for _, instr := range b.Instrs {
			if _, ok := instr.(*ssa.Phi); ok {
				// handled in WriteGoto
				continue
			}

			handledTypeChange := false

			if v, ok := instr.(ssa.Value); ok {
				name := v.Name()
				if r := v.Referrers(); r != nil && len(*r) == 0 {
					name = "_"
				}

				if t, ok := ctx.TypeOf(v).(*types.Tuple); ok {
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
							if name != "_" {
								ctx.WriteByte('_')
								ctx.WriteNumber(i)
							}
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
				ctx.WriteType(ctx.TypeOf(i).(*types.Pointer).Elem())
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
				handledTypeChange = ctx.WriteCall(i)

			case *ssa.ChangeInterface:
				ctx.WriteByte('(')
				ctx.WriteTypeOf(i)
				ctx.WriteString(")(")
				ctx.WriteName(i.X)
				ctx.WriteByte(')')

			case *ssa.ChangeType:
				ctx.WriteByte('(')
				ctx.WriteTypeOf(i)
				ctx.WriteString(")(")
				ctx.WriteName(i.X)
				ctx.WriteByte(')')

			case *ssa.Convert:
				ctx.WriteByte('(')
				ctx.WriteTypeOf(i)
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
				f := field(ctx.TypeOf(i.X), i.Field)
				ctx.WriteString(f.Name())

			case *ssa.FieldAddr:
				ctx.WriteByte('&')
				ctx.WriteName(i.X)
				ctx.WriteByte('.')
				f := field(ctx.TypeOf(i.X), i.Field)
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
				ctx.WriteTypeOf(i)
				ctx.WriteByte(',')
				ctx.WriteName(i.Size)
				ctx.WriteByte(')')

			case *ssa.MakeClosure:
				panic("specialize: TODO: implement closures")

			case *ssa.MakeInterface:
				ctx.WriteByte('(')
				ctx.WriteTypeOf(i)
				ctx.WriteString(")(")
				ctx.WriteName(i.X)
				ctx.WriteByte(')')
				handledTypeChange = true

			case *ssa.MakeMap:
				ctx.WriteString("make(")
				ctx.WriteTypeOf(i)
				if i.Reserve != nil {
					ctx.WriteByte(',')
					ctx.WriteName(i.Reserve)
				}
				ctx.WriteByte(')')

			case *ssa.MakeSlice:
				ctx.WriteString("make(")
				ctx.WriteTypeOf(i)
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
				if !types.Identical(ctx.TypeOf(i.X), ctx.TypeOf(i)) {
					ctx.WriteString(".(")
					ctx.WriteType(i.AssertedType)
					ctx.WriteByte(')')
				}
				handledTypeChange = true

			case *ssa.UnOp:
				if _, ok := i.X.(*ssa.Global); !ok || i.Op != token.MUL {
					ctx.WriteString(i.Op.String())
				}
				ctx.WriteName(i.X)

			default:
				panic("unreachable")
			}

			if v, ok := instr.(ssa.Value); ok && !handledTypeChange {
				if t := ctx.TypeOf(v); !types.Identical(t, v.Type()) {
					ctx.WriteString(".(")
					ctx.WriteType(t)
					ctx.WriteByte(')')
				}
			}
			ctx.WriteByte('\n')
		}
	}

	ctx.WriteString("}\n\n")
}

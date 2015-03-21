package main

import (
	"bytes"
	"container/heap"
	"fmt"
	"go/ast"
	"log"
	"os"
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
			if cc, ok := instr.(interface {
				Common() *ssa.CallCommon
			}); ok {
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

func Rewrite(h *FuncCallHeap, fc *FuncCall) {
	fc.F.WriteTo(os.Stdout)

	//ast.Print(iprog.Fset, f)
	//printer.Fprint(os.Stdout, iprog.Fset, f)
	fmt.Println()
}

func TypeExpr(pi *loader.PackageInfo, typ types.Type) ast.Expr {
	switch t := typ.(type) {
	case *types.Pointer:
		return &ast.StarExpr{
			X: TypeExpr(pi, t.Elem()),
		}
	case *types.Named:
		if t.Obj().Pkg() == pi.Pkg {
			return ast.NewIdent(t.Obj().Name())
		}
		if !t.Obj().Exported() {
			panic(fmt.Errorf("specialize: %v is not exported", t))
		}
		return &ast.SelectorExpr{
			X:   ast.NewIdent(t.Obj().Pkg().Name()),
			Sel: ast.NewIdent(t.Obj().Name()),
		}
	case *types.Basic:
		return ast.NewIdent(t.Name())
	case *types.Slice:
		return &ast.ArrayType{
			Elt: TypeExpr(pi, t.Elem()),
		}
	default:
		panic(fmt.Errorf("specialize: unhandled type: %T", t))
	}
}

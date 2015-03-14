package main

import (
	"log"

	"golang.org/x/tools/go/callgraph"
	"golang.org/x/tools/go/callgraph/rta"
	"golang.org/x/tools/go/loader"
	"golang.org/x/tools/go/ssa"
	"golang.org/x/tools/go/types"
)

func main() {
	log.SetFlags(0)
	log.SetPrefix("specialize: ")

	var conf loader.Config

	conf.ImportWithTests(".")

	iprog, err := conf.Load()
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
	found := make(map[*ssa.Function][]types.Type)
	for _, f := range functions {
		Analyze(f, result, found)
	}

	for f, call := range found {
		if call == nil {
			// incompatible calls to this function were found
			continue
		}

		log.Println(f)
		for i, v := range call {
			log.Printf("arg[%d] %v", i, v)
		}
	}
}

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

func Analyze(f *ssa.Function, result *rta.Result, found map[*ssa.Function][]types.Type) {
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
				call := make([]types.Type, len(c.Args))
				for i, v := range c.Args {
					call[i] = v.Type()
					if m, ok := v.(*ssa.MakeInterface); ok {
						anyInterface = true
						call[i] = m.X.Type()
					}
				}
				if anyInterface {
					if fc, ok := found[cf]; ok {
						if fc == nil {
							continue
						}

						for i, v := range fc {
							if !types.Identical(call[i], v) {
								call = nil
								break
							}
						}
					}
					found[cf] = call
				}
			}
		}
	}
}

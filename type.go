package main

import (
	"log"
	"strconv"

	"golang.org/x/tools/go/types"
)

func (ctx *Context) WriteType(typ types.Type) {
	ctx.writeType(typ, nil)
}

// modified from the function in https://github.com/golang/tools/blob/master/go/types/typestring.go#L46 to not use reflection.
func (ctx *Context) writeType(typ types.Type, visited []types.Type) {
	// Theoretically, this is a quadratic lookup algorithm, but in
	// practice deeply nested composite types with unnamed component
	// types are uncommon. This code is likely more efficient than
	// using a map.
	for _, t := range visited {
		if t == typ {
			return
		}
	}
	visited = append(visited, typ)

	switch t := typ.(type) {
	case nil:
		ctx.WriteString("<nil>")

	case *types.Basic:
		if t.Kind() == types.UnsafePointer {
			ctx.WriteString("unsafe.")
		}
		if types.GcCompatibilityMode {
			// forget the alias names
			switch t.Kind() {
			case types.Byte:
				t = types.Typ[types.Uint8]
			case types.Rune:
				t = types.Typ[types.Int32]
			}
		}
		ctx.WriteString(t.Name())

	case *types.Array:
		ctx.WriteByte('[')
		ctx.WriteString(strconv.FormatInt(t.Len(), 10))
		ctx.WriteByte(']')
		ctx.writeType(t.Elem(), visited)

	case *types.Slice:
		ctx.WriteString("[]")
		ctx.writeType(t.Elem(), visited)

	case *types.Struct:
		ctx.WriteString("struct{")
		for i, l := 0, t.NumFields(); i < l; i++ {
			f := t.Field(i)
			if i > 0 {
				ctx.WriteString("; ")
			}
			if f.Name() != "" {
				ctx.WriteString(f.Name())
				ctx.WriteByte(' ')
			}
			ctx.writeType(f.Type(), visited)
			if tag := t.Tag(i); tag != "" {
				ctx.WriteByte(' ')
				ctx.WriteString(strconv.Quote(tag))
			}
		}
		ctx.WriteByte('}')

	case *types.Pointer:
		ctx.WriteByte('*')
		ctx.writeType(t.Elem(), visited)

	case *types.Tuple:
		ctx.writeTuple(t, false, visited)

	case *types.Signature:
		ctx.WriteString("func")
		ctx.writeSignature(t, visited)

	case *types.Interface:
		// We write the source-level methods and embedded types rather
		// than the actual method set since resolved method signatures
		// may have non-printable cycles if parameters have anonymous
		// interface types that (directly or indirectly) embed the
		// current interface. For instance, consider the result type
		// of m:
		//
		//     type T interface{
		//         m() interface{ T }
		//     }
		//
		ctx.WriteString("interface{")
		if types.GcCompatibilityMode {
			// print flattened interface
			// (useful to compare against gc-generated interfaces)
			for i, l := 0, t.NumMethods(); i < l; i++ {
				m := t.Method(i)
				if i > 0 {
					ctx.WriteString("; ")
				}
				ctx.WriteString(m.Name())
				ctx.writeSignature(m.Type().(*types.Signature), visited)
			}
		} else {
			// print explicit interface methods and embedded types
			for i, l := 0, t.NumExplicitMethods(); i < l; i++ {
				m := t.ExplicitMethod(i)
				if i > 0 {
					ctx.WriteString("; ")
				}
				ctx.WriteString(m.Name())
				ctx.writeSignature(m.Type().(*types.Signature), visited)
			}
			for i, l := 0, t.NumEmbeddeds(); i < l; i++ {
				typ := t.Embedded(i)
				if i > 0 || t.NumExplicitMethods() > 0 {
					ctx.WriteString("; ")
				}
				ctx.writeType(typ, visited)
			}
		}
		ctx.WriteByte('}')

	case *types.Map:
		ctx.WriteString("map[")
		ctx.writeType(t.Key(), visited)
		ctx.WriteByte(']')
		ctx.writeType(t.Elem(), visited)

	case *types.Chan:
		var s string
		var parens bool
		switch t.Dir() {
		case types.SendRecv:
			s = "chan "
			// chan (<-chan T) requires parentheses
			if c, _ := t.Elem().(*types.Chan); c != nil && c.Dir() == types.RecvOnly {
				parens = true
			}
		case types.SendOnly:
			s = "chan<- "
		case types.RecvOnly:
			s = "<-chan "
		default:
			panic("unreachable")
		}
		ctx.WriteString(s)
		if parens {
			ctx.WriteByte('(')
		}
		ctx.writeType(t.Elem(), visited)
		if parens {
			ctx.WriteByte(')')
		}

	case *types.Named:
		s := "<Named w/o object>"
		if obj := t.Obj(); obj != nil {
			if pkg := obj.Pkg(); pkg != nil && pkg != ctx.Pkg {
				if ctx.Imports != nil {
					ctx.Imports[pkg.Path()] = true
				}
				ctx.WriteString(pkg.Name())
				ctx.WriteByte('.')
			}
			// TODO(gri): function-local named types should be displayed
			// differently from named types at package level to avoid
			// ambiguity.
			s = obj.Name()
		}
		ctx.WriteString(s)

	default:
		log.Panicf("%#v", t)
		panic("unreachable")
		//	// For externally defined implementations of Type.
		//	ctx.WriteString(t.String())
	}
}

func (ctx *Context) writeTuple(tup *types.Tuple, variadic bool, visited []types.Type) {
	ctx.WriteByte('(')
	if tup != nil {
		for i, l := 0, tup.Len(); i < l; i++ {
			v := tup.At(i)
			if i > 0 {
				ctx.WriteString(", ")
			}
			if v.Name() != "" {
				ctx.WriteString(v.Name())
				ctx.WriteByte(' ')
			}
			typ := v.Type()
			if variadic && i == tup.Len()-1 {
				if s, ok := typ.(*types.Slice); ok {
					ctx.WriteString("...")
					typ = s.Elem()
				} else {
					// special case:
					// append(s, "foo"...) leads to signature func([]byte, string...)
					if t, ok := typ.Underlying().(*types.Basic); !ok || t.Kind() != types.String {
						panic("internal error: string type expected")
					}
					ctx.writeType(typ, visited)
					ctx.WriteString("...")
					continue
				}
			}
			ctx.writeType(typ, visited)
		}
	}
	ctx.WriteByte(')')
}

func (ctx *Context) writeSignature(sig *types.Signature, visited []types.Type) {
	ctx.writeTuple(sig.Params(), sig.Variadic(), visited)

	n := sig.Results().Len()
	if n == 0 {
		// no result
		return
	}

	ctx.WriteByte(' ')
	if n == 1 && sig.Results().At(0).Name() == "" {
		// single unnamed result
		ctx.writeType(sig.Results().At(0).Type(), visited)
		return
	}

	// multiple or named result(s)
	ctx.writeTuple(sig.Results(), false, visited)
}

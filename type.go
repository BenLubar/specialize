package main

import (
	"bytes"
	"strconv"

	"golang.org/x/tools/go/types"
)

// modified from the function in https://github.com/golang/tools/blob/master/go/types/typestring.go#L46 to not use reflection.
func writeType(buf *bytes.Buffer, this *types.Package, typ types.Type, visited []types.Type) {
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
		buf.WriteString("<nil>")

	case *types.Basic:
		if t.Kind() == types.UnsafePointer {
			buf.WriteString("unsafe.")
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
		buf.WriteString(t.Name())

	case *types.Array:
		buf.WriteByte('[')
		buf.WriteString(strconv.FormatInt(t.Len(), 10))
		buf.WriteByte(']')
		writeType(buf, this, t.Elem(), visited)

	case *types.Slice:
		buf.WriteString("[]")
		writeType(buf, this, t.Elem(), visited)

	case *types.Struct:
		buf.WriteString("struct{")
		for i, l := 0, t.NumFields(); i < l; i++ {
			f := t.Field(i)
			if i > 0 {
				buf.WriteString("; ")
			}
			if f.Name() != "" {
				buf.WriteString(f.Name())
				buf.WriteByte(' ')
			}
			writeType(buf, this, f.Type(), visited)
			if tag := t.Tag(i); tag != "" {
				buf.WriteByte(' ')
				buf.WriteString(strconv.Quote(tag))
			}
		}
		buf.WriteByte('}')

	case *types.Pointer:
		buf.WriteByte('*')
		writeType(buf, this, t.Elem(), visited)

	case *types.Tuple:
		writeTuple(buf, this, t, false, visited)

	case *types.Signature:
		buf.WriteString("func")
		writeSignature(buf, this, t, visited)

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
		buf.WriteString("interface{")
		if types.GcCompatibilityMode {
			// print flattened interface
			// (useful to compare against gc-generated interfaces)
			for i, l := 0, t.NumMethods(); i < l; i++ {
				m := t.Method(i)
				if i > 0 {
					buf.WriteString("; ")
				}
				buf.WriteString(m.Name())
				writeSignature(buf, this, m.Type().(*types.Signature), visited)
			}
		} else {
			// print explicit interface methods and embedded types
			for i, l := 0, t.NumExplicitMethods(); i < l; i++ {
				m := t.ExplicitMethod(i)
				if i > 0 {
					buf.WriteString("; ")
				}
				buf.WriteString(m.Name())
				writeSignature(buf, this, m.Type().(*types.Signature), visited)
			}
			for i, l := 0, t.NumEmbeddeds(); i < l; i++ {
				typ := t.Embedded(i)
				if i > 0 || t.NumExplicitMethods() > 0 {
					buf.WriteString("; ")
				}
				writeType(buf, this, typ, visited)
			}
		}
		buf.WriteByte('}')

	case *types.Map:
		buf.WriteString("map[")
		writeType(buf, this, t.Key(), visited)
		buf.WriteByte(']')
		writeType(buf, this, t.Elem(), visited)

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
		buf.WriteString(s)
		if parens {
			buf.WriteByte('(')
		}
		writeType(buf, this, t.Elem(), visited)
		if parens {
			buf.WriteByte(')')
		}

	case *types.Named:
		s := "<Named w/o object>"
		if obj := t.Obj(); obj != nil {
			if pkg := obj.Pkg(); pkg != nil && pkg != this {
				//buf.WriteString(pkg.Path())
				buf.WriteString(pkg.Name())
				buf.WriteByte('.')
			}
			// TODO(gri): function-local named types should be displayed
			// differently from named types at package level to avoid
			// ambiguity.
			s = obj.Name()
		}
		buf.WriteString(s)

	default:
		panic("unreachable")
		//	// For externally defined implementations of Type.
		//	buf.WriteString(t.String())
	}
}

func writeTuple(buf *bytes.Buffer, this *types.Package, tup *types.Tuple, variadic bool, visited []types.Type) {
	buf.WriteByte('(')
	if tup != nil {
		for i, l := 0, tup.Len(); i < l; i++ {
			v := tup.At(i)
			if i > 0 {
				buf.WriteString(", ")
			}
			if v.Name() != "" {
				buf.WriteString(v.Name())
				buf.WriteByte(' ')
			}
			typ := v.Type()
			if variadic && i == tup.Len()-1 {
				if s, ok := typ.(*types.Slice); ok {
					buf.WriteString("...")
					typ = s.Elem()
				} else {
					// special case:
					// append(s, "foo"...) leads to signature func([]byte, string...)
					if t, ok := typ.Underlying().(*types.Basic); !ok || t.Kind() != types.String {
						panic("internal error: string type expected")
					}
					writeType(buf, this, typ, visited)
					buf.WriteString("...")
					continue
				}
			}
			writeType(buf, this, typ, visited)
		}
	}
	buf.WriteByte(')')
}

func writeSignature(buf *bytes.Buffer, this *types.Package, sig *types.Signature, visited []types.Type) {
	writeTuple(buf, this, sig.Params(), sig.Variadic(), visited)

	n := sig.Results().Len()
	if n == 0 {
		// no result
		return
	}

	buf.WriteByte(' ')
	if n == 1 && sig.Results().At(0).Name() == "" {
		// single unnamed result
		writeType(buf, this, sig.Results().At(0).Type(), visited)
		return
	}

	// multiple or named result(s)
	writeTuple(buf, this, sig.Results(), false, visited)
}

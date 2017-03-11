package pgtype

import (
	"fmt"
	"io"
	"reflect"
)

// ACLItem is used for PostgreSQL's aclitem data type. A sample aclitem
// might look like this:
//
//	postgres=arwdDxt/postgres
//
// Note, however, that because the user/role name part of an aclitem is
// an identifier, it follows all the usual formatting rules for SQL
// identifiers: if it contains spaces and other special characters,
// it should appear in double-quotes:
//
//	postgres=arwdDxt/"role with spaces"
//
type ACLItem struct {
	String string
	Status Status
}

func (dst *ACLItem) ConvertFrom(src interface{}) error {
	switch value := src.(type) {
	case ACLItem:
		*dst = value
	case string:
		*dst = ACLItem{String: value, Status: Present}
	case *string:
		if value == nil {
			*dst = ACLItem{Status: Null}
		} else {
			*dst = ACLItem{String: *value, Status: Present}
		}
	default:
		if originalSrc, ok := underlyingStringType(src); ok {
			return dst.ConvertFrom(originalSrc)
		}
		return fmt.Errorf("cannot convert %v to ACLItem", value)
	}

	return nil
}

func (src *ACLItem) AssignTo(dst interface{}) error {
	switch v := dst.(type) {
	case *string:
		if src.Status != Present {
			return fmt.Errorf("cannot assign %v to %T", src, dst)
		}
		*v = src.String
	default:
		if v := reflect.ValueOf(dst); v.Kind() == reflect.Ptr {
			el := v.Elem()
			switch el.Kind() {
			// if dst is a pointer to pointer, strip the pointer and try again
			case reflect.Ptr:
				if src.Status == Null {
					el.Set(reflect.Zero(el.Type()))
					return nil
				}
				if el.IsNil() {
					// allocate destination
					el.Set(reflect.New(el.Type().Elem()))
				}
				return src.AssignTo(el.Interface())
			case reflect.String:
				if src.Status != Present {
					return fmt.Errorf("cannot assign %v to %T", src, dst)
				}
				el.SetString(src.String)
				return nil
			}
		}
		return fmt.Errorf("cannot decode %v into %T", src, dst)
	}

	return nil
}

func (dst *ACLItem) DecodeText(src []byte) error {
	if src == nil {
		*dst = ACLItem{Status: Null}
		return nil
	}

	*dst = ACLItem{String: string(src), Status: Present}
	return nil
}

func (src ACLItem) EncodeText(w io.Writer) (bool, error) {
	switch src.Status {
	case Null:
		return true, nil
	case Undefined:
		return false, errUndefined
	}

	_, err := io.WriteString(w, src.String)
	return false, err
}
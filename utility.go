package mapper

import "reflect"

// structFields returns a slice of all of the StructFields within a given struct
func (s *StructToBSON) structFields() []reflect.StructField {
	t := s.value.Type()

	f := make([]reflect.StructField, 0)

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		// Can't access the value of unexported fields
		if field.PkgPath != "" {
			continue
		}

		// Ignoring omitted fields
		if tag := field.Tag.Get(s.TagName); tag == "-" {
			continue
		}

		f = append(f, field)
	}
	return f
}

// structVal checks if the argument is a struct or a pointer to a struct
// if so it returns the reflected value of the struct
//
// Panics if a struct || *struct is not passed to the function
func structVal(s interface{}) reflect.Value {
	v := reflect.ValueOf(s)

	for v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		panic("not struct")
	}

	return v
}

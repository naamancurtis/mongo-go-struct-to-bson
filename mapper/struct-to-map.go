package mapper

import (
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"reflect"
)

// Package built based off https://github.com/fatih/structs/

var (
	// By default, this package uses "bson" as the tag name
	// You can over-write this once you have wrapped your struct
	// in the mapping struct (StructToBSON) by chaining the
	// .SetTagName() call on the wrapped struct.
	DefaultTagName = "bson"
)

// The wrapper for your struct that enables this package to work
type StructToBSON struct {
	raw     interface{}
	value   reflect.Value
	TagName string
}

type MappingOpts struct {
	// Will just return a map with {_id: idVal} if it is present at the Top Level of the struct, else it returns all
	// other fields
	// Setting true on this flag over-writes means this result is prioritised over all other options that clash
	// with this Behavior.
	//
	//   Default: False
	UseIDifAvailable bool

	// Set to true if you require _id removed from your map
	// This will remove _id properties from nested data structures as well
	//
	// 	Default: False
	RemoveID bool

	// Set to true if you are generating a filter
	// If true, it will check all struct fields for zero type values and
	// omit any that are found regardless of any tag options
	//
	// This logic occurs after UseIDifAvailable & RemoveID
	//
	// 	Default: False
	GenerateFilter bool
}

// Returns the Input struct wrapped by the mapper struct
//
// Panics if the argument is not a struct or pointer to a struct
func NewBSONMapperStruct(s interface{}) *StructToBSON {
	return &StructToBSON{
		raw:     s,
		value:   structVal(s),
		TagName: DefaultTagName,
	}
}

// Sets the tag name to be parsed
//  // Default: `bson`
func (s *StructToBSON) SetTagName(tag string) {
	s.TagName = tag
}

// Wraps a struct and converts it to a BSON Map, factoring in any options passed
// as arguments
//
// TODO - Add documentation about "-" "omitempty", "omitnested", "flatten", "string
// It uses the tag name `bson` on the struct fields to generate the map
//
// The mapping is recursive for any data structures contained within the struct
//
// Example StructToBSON to be converted:
//
//   // type ExampleStruct struct {
//   //    Value1 string `bson:"myFirstValue"`
//   //	   Value2 []int `bson:"myIntSlice"`
//   // }
//
// The struct is first wrapped with the "StructToBSON" type to give
// access to the mapping functions
//
// The struct is then converted to a bson.M and returned
//
// Returns:
//
//   // bson.M {
//   //    { Key: "myFirstValue", Value: "Example String" },
//   //    { Key: "myIntSlice", Value: {1, 2, 3, 4, 5} },
//   // }
//
// The following tags are factored into the parsing:
//
// 	 // "omitempty" - Omit if the value is the zero value
// 	 // "omitnested" - Pass the value as opposed to recursively mapping the struct
// 	 // "flatten" - Pull out the data from the nested struct up one level
// 	 // "string" - Use the implementation of the Stringer interface for the value
// 	 // "-" - Do not map this field
//
func ConvertStructToBSONMap(s interface{}, opts *MappingOpts) bson.M {
	if reflect.ValueOf(s).Kind() != reflect.Struct && !(reflect.ValueOf(s).Kind() == reflect.Ptr && reflect.ValueOf(s).Elem().Kind() == reflect.Struct) {
		return nil
	}
	return NewBSONMapperStruct(s).ToBSONMap(opts)
}

func (s *StructToBSON) ToBSONMap(opts *MappingOpts) bson.M {
	out := bson.M{}

	fields := s.structFields()

	for _, field := range fields {
		name := field.Name
		val := s.value.FieldByName(name)
		isSubStruct := false
		var finalVal interface{}

		// Identify whether the struct field has tags or not
		tagName, tagOpts := parseTag(field.Tag.Get(s.TagName))
		if tagName != "" {
			name = tagName
		}

		if opts != nil && tagName == "_id" {
			if opts.UseIDifAvailable {
				return bson.M{"_id": val.Interface()}
			}
			if opts.RemoveID {
				continue
			}
		}

		// Decide whether to omit the field if it is empty or not
		if tagOpts.Has("omitempty") || (opts != nil && opts.GenerateFilter) {

			if val.IsZero() {
				continue
			}

			// Handling edge cases that reflect.value.IsZero doesn't catch
			switch val.Kind() {
			case reflect.Slice:
				if val.Len() == 0 {
					continue
				}
			case reflect.Map:
				if len(val.MapKeys()) == 0 {
					continue
				}
			}
		}

		// If nested data structures should not be omitted
		if !tagOpts.Has("omitnested") {
			finalVal = s.nestedData(val, opts)

			v := reflect.ValueOf(val.Interface())
			if v.Kind() == reflect.Ptr {
				v = v.Elem()
			}

			switch v.Kind() {
			case reflect.Map, reflect.Struct:
				isSubStruct = true
			}
		} else {
			finalVal = val.Interface()
		}

		// If the field should be a string, convert it to a string
		if tagOpts.Has("string") {
			s, ok := val.Interface().(fmt.Stringer)
			if ok {
				out[name] = s.String()
			}
			continue
		}

		// If the nested data objects should be flattened
		if isSubStruct && (tagOpts.Has("flatten")) {
			outMap := finalVal.(primitive.M)
			for k := range finalVal.(primitive.M) {
				out[k] = outMap[k]
			}
		} else {
			out[name] = finalVal
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// Identifies the nested data type and recursively iterates over it
// to return a BSON document for the nested data structure
func (s *StructToBSON) nestedData(val reflect.Value, opts *MappingOpts) interface{} {
	var finalVal interface{}
	v := reflect.ValueOf(val.Interface())

	// Converting a pointer to a value
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.Struct:
		n := NewBSONMapperStruct(val.Interface())
		n.TagName = s.TagName
		m := n.ToBSONMap(opts)

		if len(m) == 0 {
			finalVal = val.Interface()
		} else {
			finalVal = m
		}

	case reflect.Map:
		// Find the type of the value within the map
		mapElem := val.Type()
		switch mapElem.Kind() {
		case reflect.Ptr, reflect.Array, reflect.Map, reflect.Slice, reflect.Chan:
			mapElem = mapElem.Elem()
			if mapElem.Kind() == reflect.Ptr {
				mapElem = mapElem.Elem()
			}
		}

		// If we need to iterate over some form of struct in the map
		// ie. map[string]struct
		if mapElem.Kind() == reflect.Struct || (mapElem.Kind() == reflect.Slice && mapElem.Elem().Kind() == reflect.Struct) {
			m := bson.M{}
			for _, k := range val.MapKeys() {
				m[k.String()] = s.nestedData(val.MapIndex(k), opts)
			}
			finalVal = m
			break
		}
		finalVal = val.Interface()

	case reflect.Slice, reflect.Array:
		if val.Type().Kind() == reflect.Ptr {
			val = val.Elem()
		}

		// Ensuring there are no structs (which require further iteration) anywhere within the slice/array
		// As long as there are not, we just pass the value of the array/slice
		if val.Type().Elem().Kind() != reflect.Struct && !(val.Type().Elem().Kind() == reflect.Ptr && val.Type().Elem().Elem().Kind() == reflect.Struct) {
			finalVal = val.Interface()
			break
		}

		// If further iteration is needed, then iterate over the slice
		slices := make([]interface{}, val.Len())
		for x := 0; x < val.Len(); x++ {
			slices[x] = s.nestedData(val.Index(x), opts)
		}
		finalVal = slices

	default:
		finalVal = val.Interface()
	}

	return finalVal
}
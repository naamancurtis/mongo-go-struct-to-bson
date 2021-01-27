// Provides utility methods to support the converting of structs to bson maps for use in various MongoDB queries/patch updates.
//
// It is intended to be used alongside the Mongo-Go Driver
package mapper

import (
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"reflect"
)

// Package built based off https://github.com/fatih/structs/

var (
	// By default, this package uses `bson` as the tag name
	// You can over-write this once you have wrapped your struct
	// in the mapping struct (StructToBSON) by chaining the
	// .SetTagName() call on the wrapped struct.
	DefaultTagName = "bson"
)

// StructToBson is the wrapper for a struct that enables this package to work
type StructToBSON struct {
	raw     interface{}
	value   reflect.Value
	TagName string
}

// MappingOpts allows the setting of options which drive the behaviour behind how the struct is parsed
type MappingOpts struct {
	// Will just return bson.M { "_id": idVal } if the "_id" tag is present in that struct,
	// if it is not present or holds a zero value it will map the struct as you would expect.
	// Setting true on this flag gives it priority over all other functionality.
	// ie. If "_id" is present, all other fields will be ignored
	//
	// This option is included in recursive calls, so if a nested struct
	// has an "_id" tag (and the top level struct didn't) then the
	// nested struct field in the bson.M will only hold the { "_id": idVal } result.
	//
	//   // Default: False
	UseIDifAvailable bool

	// Will remove any "_id" fields from your bson.M
	// Note: this will remove "_id" fields from nested data structures as well
	//
	// 	// Default: False
	RemoveID bool

	// If true, it will check all struct fields for zero type values and
	// omit any that are found regardless of any tag options, effectively it enforces
	// the behaviour of the "omitempty" tag, regardless of whether the struct field
	// has it or not
	//
	// This logic occurs after UseIDifAvailable & RemoveID
	//
	// 	// Default: False
	GenerateFilterOrPatch bool
}

// NewBSONMapperStruct returns the input struct wrapped by the mapper struct
// along with the tag name which should be parsed in the mapping
//
// Panics if the argument is not a struct or pointer to a struct
func NewBSONMapperStruct(s interface{}) *StructToBSON {
	return &StructToBSON{
		raw:     s,
		value:   structVal(s),
		TagName: DefaultTagName,
	}
}

// SetTagName sets the tag name to be parsed
func (s *StructToBSON) SetTagName(tag string) {
	s.TagName = tag
}

// ConvertStructToBSONMap wraps a struct and converts it to a BSON Map, factoring in any options passed
// as arguments
// By default, it uses the tag name `bson` on the struct fields to generate the map
// The mapping is recursive for any data structures contained within the struct
//
// Example StructToBSON to be converted:
//
//   type ExampleStruct struct {
//      Value1 string `bson:"myFirstValue"`
//      Value2 []int `bson:"myIntSlice"`
//   }
//
// The struct is first wrapped with the "StructToBSON" type to give
// access to the mapping functions and is then converted to a bson.M
//
//   bson.M {
//      { Key: "myFirstValue", Value: "Example String" },
//      { Key: "myIntSlice", Value: {1, 2, 3, 4, 5} },
//   }
//
// The following tag options are factored into the parsing:
//
// 	 // "omitempty" - Omit if the value is the zero value
// 	 // "omitnested" - Pass the value of the struct directly as opposed to recursively mapping the struct
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

// ToBSONMap parses all struct fields and returns a bson.M { tagName: value }.
// If there are nested structs it calls recursively maps them as well
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
			if opts.UseIDifAvailable && val.Interface() != "" {
				return bson.M{"_id": val.Interface()}
			}
			if opts.RemoveID {
				continue
			}
		}

		// Decide whether to omit the field if it is empty or not
		if tagOpts.Has("omitempty") || (opts != nil && opts.GenerateFilterOrPatch) {

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

// nestedData identifies the nested data type and iterates over it
// to return a BSON map for the nested data structure
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

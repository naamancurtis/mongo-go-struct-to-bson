package main

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"go.mongodb.org/mongo-driver/bson"
	"reflect"
	"time"
)

var _ = Describe("Utility method", func() {
	It("NewBSONMapperStruct should return a new wrapped struct", func() {
		testStruct := struct {
			TestField1 string
			TestField2 bool
		}{
			TestField1: "Test String",
			TestField2: true,
		}

		result := NewBSONMapperStruct(testStruct)
		Expect(result.value.Interface()).To(Equal(reflect.ValueOf(testStruct).Interface()))
		Expect(result.raw).To(Equal(testStruct))
		Expect(result.TagName).To(Equal("bson"))
		Expect(reflect.ValueOf(result).Kind()).To(Equal(reflect.Ptr))
		Expect(reflect.ValueOf(result).Elem().Kind()).To(Equal(reflect.Struct))
	})

	It("SetTagName should set a new TagName", func() {
		testStruct := NewBSONMapperStruct(
			struct {
				TestField1 string
				TestField2 bool
			}{
				TestField1: "Test String",
				TestField2: true,
			})

		testStruct.SetTagName("TestTag")
		Expect(testStruct.TagName).To(Equal("TestTag"))
	})
})

var _ = Describe("The Mapping functions", func() {

	// Testing the functionality of the Mapping Options
	Context("should correctly utilise the Mapping Options struct", func() {
		type structWithID struct {
			ID         string `bson:"_id"`
			TestField1 int    `bson:"testField1"`
			TestField2 struct {
				ID         string `bson:"_id"`
				TestField3 bool   `bson:"testField3"`
			} `bson:"testField2"`
		}

		var testStruct structWithID
		BeforeEach(func() {
			testStruct = structWithID{
				ID:         "TEST ID 1",
				TestField1: 900,
				TestField2: struct {
					ID         string `bson:"_id"`
					TestField3 bool   `bson:"testField3"`
				}{
					ID:         "TEST ID 2",
					TestField3: true,
				},
			}
		})

		It("when UseID is set to true", func() {
			result := ConvertStructToBSONMap(testStruct, &MappingOpts{UseIDifAvailable: true})
			Expect(result).To(Equal(bson.M{"_id": "TEST ID 1"}))
		})

		It("when RemoveID is set to true", func() {
			result := ConvertStructToBSONMap(testStruct, &MappingOpts{RemoveID: true})

			expected := bson.M{
				"testField1": 900,
				"testField2": bson.M{
					"testField3": true,
				},
			}
			Expect(result).To(Equal(expected))
		})

		It("when UseID & RemoveID are set to true", func() {
			result := ConvertStructToBSONMap(testStruct, &MappingOpts{UseIDifAvailable: true, RemoveID: true})
			Expect(result).To(Equal(bson.M{"_id": "TEST ID 1"}))
		})
	})

	// Testing the functionality of the mapping a flat struct (no nested structs)
	Context("should", func() {
		type valueStruct struct {
			Str         string
			Num         int
			Bool        bool
			Float       float64
			Time        time.Time
			SliceInt    []int
			SliceString []string
			Map         map[string]int
		}

		type flatStruct struct {
			// Values
			Str         string         `bson:"str"`
			Num         int            `bson:"num"`
			Bool        bool           `bson:"bool"`
			Float       float64        `bson:"float"`
			Time        time.Time      `bson:"time"`
			SliceInt    []int          `bson:"sliceInt"`
			SliceString []string       `bson:"sliceString"`
			Map         map[string]int `bson:"map"`

			// Pointers
			StrPtr         *string         `bson:"strPtr"`
			NumPtr         *int            `bson:"numPtr"`
			BoolPtr        *bool           `bson:"boolPtr"`
			FloatPtr       *float64        `bson:"floatPtr"`
			TimePtr        *time.Time      `bson:"timePtr"`
			SliceIntPtr    *[]int          `bson:"sliceIntPtr"`
			SliceStringPtr *[]string       `bson:"sliceStringPtr"`
			MapPtr         *map[string]int `bson:"mapPtr"`
		}

		var validValues valueStruct

		BeforeEach(func() {
			validValues = valueStruct{
				Str:         "Test String",
				Num:         10,
				Bool:        true,
				Float:       10.1,
				Time:        time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
				SliceInt:    []int{1, 2, 3},
				SliceString: []string{"Test1", "Test 2"},
				Map:         map[string]int{"Test 1": 10, "Test 2!": 100},
			}
		})

		Context("map a flat struct", func() {
			var testStruct flatStruct
			var expected bson.M

			BeforeEach(func() {
				testStruct = flatStruct{
					Str:         validValues.Str,
					Num:         validValues.Num,
					Bool:        validValues.Bool,
					Float:       validValues.Float,
					Time:        validValues.Time,
					SliceInt:    validValues.SliceInt,
					SliceString: validValues.SliceString,
					Map:         validValues.Map,

					StrPtr:         &validValues.Str,
					NumPtr:         &validValues.Num,
					BoolPtr:        &validValues.Bool,
					FloatPtr:       &validValues.Float,
					TimePtr:        &validValues.Time,
					SliceIntPtr:    &validValues.SliceInt,
					SliceStringPtr: &validValues.SliceString,
					MapPtr:         &validValues.Map,
				}

				expected = bson.M{
					"str":         validValues.Str,
					"num":         validValues.Num,
					"bool":        validValues.Bool,
					"float":       validValues.Float,
					"time":        validValues.Time,
					"sliceInt":    validValues.SliceInt,
					"sliceString": validValues.SliceString,
					"map":         validValues.Map,

					"strPtr":         &validValues.Str,
					"numPtr":         &validValues.Num,
					"boolPtr":        &validValues.Bool,
					"floatPtr":       &validValues.Float,
					"timePtr":        &validValues.Time,
					"sliceIntPtr":    validValues.SliceInt,
					"sliceStringPtr": validValues.SliceString,
					"mapPtr":         &validValues.Map,
				}
			})
			It("should map a flat struct with nil options", func() {
				result := ConvertStructToBSONMap(testStruct, nil)
				Expect(len(result)).To(Equal(len(expected)))
				for k, _ := range result {
					Expect(result[k]).To(Equal(expected[k]))
				}
			})

			It("should map a flat struct with UseID set to true - with no _id provided", func() {
				result := ConvertStructToBSONMap(testStruct, &MappingOpts{UseIDifAvailable: true})
				Expect(len(result)).To(Equal(len(expected)))
				for k, _ := range result {
					Expect(result[k]).To(Equal(expected[k]))
				}
			})

			It("should map a flat struct with RemoveID set to true - no _id provided", func() {
				result := ConvertStructToBSONMap(testStruct, &MappingOpts{RemoveID: true})
				Expect(len(result)).To(Equal(len(expected)))
				for k, _ := range result {
					Expect(result[k]).To(Equal(expected[k]))
				}
			})

			It("should map a flat struct with both options set to true - no _id provided", func() {
				result := ConvertStructToBSONMap(testStruct, &MappingOpts{UseIDifAvailable: true, RemoveID: true})
				Expect(len(result)).To(Equal(len(expected)))
				for k, _ := range result {
					Expect(result[k]).To(Equal(expected[k]))
				}
			})
		})
	})

	// Testing the functionality of the Mapping zero values
	Context("should ignore zero values when flagged with omitempty", func() {
		DescribeTable("the type is", func(c interface{}) {
			result := ConvertStructToBSONMap(c, nil)
			Expect(result).To(BeNil())
		},
			Entry("a string", struct {
				Input string `bson:"Input,omitempty"`
			}{
				Input: "",
			}),
			Entry("an int", struct {
				Input int `bson:"Input,omitempty"`
			}{
				Input: 0,
			}),
			Entry("a bool", struct {
				Input bool `bson:"Input,omitempty"`
			}{
				Input: false,
			}),
			Entry("a slice", struct {
				Input []uint8 `bson:"Input,omitempty"`
			}{
				Input: []uint8{},
			}),
			Entry("a map", struct {
				Input map[string]struct{} `bson:"Input,omitempty"`
			}{
				Input: map[string]struct{}{},
			}),
			Entry("a nil value", struct {
				Input *struct{} `bson:"Input,omitempty"`
			}{
				Input: nil,
			}),
		)
	})
})

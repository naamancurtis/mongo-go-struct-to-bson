package mapper

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
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
			},
		)

		testStruct.SetTagName("TestTag")
		Expect(testStruct.TagName).To(Equal("TestTag"))
	})

	DescribeTable("CovertStructToBSONMap should return nil if", func(c interface{}) {
		result := ConvertStructToBSONMap(c, nil)
		Expect(result).To(BeNil())
	},
		Entry("a string is passed", "Test String"),
		Entry("an int is passed", 123),
		Entry("a bool is passed", true),
		Entry("a slice is passed", []int{1, 2, 3}),
		Entry("a map is passed", map[string]struct{}{"Test 1": struct{}{}, "Test 2": struct{}{}}),
		Entry("a function is passed", func() {}),
		Entry("a pointer to a string is passed", new(string)),
		Entry("a pointer to an int is passed", new(int)),
		Entry("a pointer to a bool is passed", new(bool)),
		Entry("a pointer to a slice is passed", new([]int)),
		Entry("a pointer to a map is passed", new(map[string]struct{})),
		Entry("a pointer to a function is passed", new(func())),
	)
})

var _ = Describe("The Mapping functions", func() {

	// Testing the functionality of the Mapping Options
	Context("should correctly utilise the Mapping Options struct", func() {
		type structWithID struct {
			ID         string `bson:"_id,omitempty"`
			TestField1 int    `bson:"testField1"`
			TestField2 struct {
				ID         string `bson:"_id,omitempty"`
				TestField3 bool   `bson:"testField3"`
			} `bson:"testField2"`
		}

		var testStruct structWithID
		BeforeEach(func() {
			testStruct = structWithID{
				ID:         "TEST ID 1",
				TestField1: 900,
				TestField2: struct {
					ID         string `bson:"_id,omitempty"`
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

		It("when UseID is set to true but no ID is provided", func() {
			testStruct.ID = ""
			testStruct.TestField2.ID = ""

			expected := bson.M{
				"testField1": 900,
				"testField2": bson.M{
					"testField3": true,
				},
			}

			result := ConvertStructToBSONMap(testStruct, &MappingOpts{UseIDifAvailable: true})
			Expect(result).To(Equal(expected))
		})

		It("when UseID is set to true but no ID is provided on the top level struct", func() {
			testStruct.ID = ""

			expected := bson.M{
				"testField1": 900,
				"testField2": bson.M{
					"_id": "TEST ID 2",
				},
			}

			result := ConvertStructToBSONMap(testStruct, &MappingOpts{UseIDifAvailable: true})
			Expect(result).To(Equal(expected))
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
			StrPtr         *string    `bson:"strPtr"`
			NumPtr         *int       `bson:"numPtr"`
			BoolPtr        *bool      `bson:"boolPtr"`
			FloatPtr       *float64   `bson:"floatPtr"`
			TimePtr        *time.Time `bson:"timePtr"`
			SliceIntPtr    *[]int     `bson:"sliceIntPtr"`
			SliceStringPtr *[]string  `bson:"sliceStringPtr"`
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
				Map:         map[string]int{"Test 1": 0, "Test 2!": 100, "Test 3": -100},
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

	// Testing the functionality of the mapping zero values
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
			Entry("a struct", struct {
				Input struct{} `bson:"Input,omitempty"`
			}{
				Input: struct{}{},
			}),
			Entry("a nil value", struct {
				Input *struct{} `bson:"Input,omitempty"`
			}{
				Input: nil,
			}),
		)
	})

	// Testing the functionality of the "string" tag
	Context("should convert", func() {
		It("a struct that implements Stringer interface to a string", func() {
			// Using time.Time as it implements the Stringer interface
			result := ConvertStructToBSONMap(
				struct {
					TestField1 time.Time `bson:"testField1,string"`
				}{
					TestField1: time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
				}, nil,
			)
			Expect(result).To(Equal(bson.M{"testField1": "2000-01-01 00:00:00 +0000 UTC"}))
		})
	})

	// Testing the functionality of nested structs
	Context("should correctly parse", func() {
		type valueStruct struct {
			String string         `bson:"str"`
			Slice  []string       `bson:"slice"`
			Map    map[string]int `bson:"map"`
			Time   time.Time      `bson:"time"`
		}

		var valuesStruct valueStruct
		BeforeEach(func() {
			valuesStruct = valueStruct{
				String: "Test String",
				Slice:  []string{"Test 1", "!@£$%^&*()", ""},
				Map:    map[string]int{"Test 1": 100, "Test 2!": 0, "Test 3": -100},
				Time:   time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
			}
		})

		It("a nested struct", func() {
			result := ConvertStructToBSONMap(
				struct {
					TestField1 string      `bson:"testField1"`
					TestField2 valueStruct `bson:"nestedStruct"`
				}{
					TestField1: valuesStruct.String,
					TestField2: valueStruct{
						String: valuesStruct.String,
						Slice:  valuesStruct.Slice,
						Map:    valuesStruct.Map,
						Time:   valuesStruct.Time,
					},
				}, nil,
			)

			expected := bson.M{
				"str":   valuesStruct.String,
				"slice": valuesStruct.Slice,
				"map":   valuesStruct.Map,
				"time":  valuesStruct.Time,
			}

			Expect(result["testField1"]).To(Equal(valuesStruct.String))
			for k, _ := range result["nestedStruct"].(bson.M) {
				Expect(result["nestedStruct"].(bson.M)[k]).To(Equal(expected[k]))
			}
		})

		It("a pointer to a nested struct", func() {
			result := ConvertStructToBSONMap(
				&struct {
					TestField1 string      `bson:"testField1"`
					TestField2 valueStruct `bson:"nestedStruct"`
				}{
					TestField1: valuesStruct.String,
					TestField2: valueStruct{
						String: valuesStruct.String,
						Slice:  valuesStruct.Slice,
						Map:    valuesStruct.Map,
						Time:   valuesStruct.Time,
					},
				}, nil,
			)

			expected := bson.M{
				"str":   valuesStruct.String,
				"slice": valuesStruct.Slice,
				"map":   valuesStruct.Map,
				"time":  valuesStruct.Time,
			}

			Expect(result["testField1"]).To(Equal(valuesStruct.String))
			for k, _ := range result["nestedStruct"].(bson.M) {
				Expect(result["nestedStruct"].(bson.M)[k]).To(Equal(expected[k]))
			}
		})

		It("a nested struct with the omitnested tag", func() {
			result := ConvertStructToBSONMap(
				struct {
					TestField1 string      `bson:"testField1"`
					TestField2 valueStruct `bson:"nestedStruct,omitnested"`
				}{
					TestField1: valuesStruct.String,
					TestField2: valueStruct{
						String: valuesStruct.String,
						Slice:  valuesStruct.Slice,
						Map:    valuesStruct.Map,
						Time:   valuesStruct.Time,
					},
				}, nil,
			)

			expected := bson.M{
				"testField1": valuesStruct.String,
				"nestedStruct": valueStruct{
					String: valuesStruct.String,
					Slice:  valuesStruct.Slice,
					Map:    valuesStruct.Map,
					Time:   valuesStruct.Time,
				},
			}

			Expect(result).To(Equal(expected))
		})

		It("a nested struct with the flatten tag", func() {
			result := ConvertStructToBSONMap(
				struct {
					TestField1 string      `bson:"testField1"`
					TestField2 valueStruct `bson:"nestedStruct,flatten"`
				}{
					TestField1: valuesStruct.String,
					TestField2: valueStruct{
						String: valuesStruct.String,
						Slice:  valuesStruct.Slice,
						Map:    valuesStruct.Map,
						Time:   valuesStruct.Time,
					},
				}, nil,
			)

			expected := bson.M{
				"testField1": valuesStruct.String,
				"str":        valuesStruct.String,
				"slice":      valuesStruct.Slice,
				"map":        valuesStruct.Map,
				"time":       valuesStruct.Time,
			}

			Expect(result).To(Equal(expected))
		})

		It("a nested struct with a slice of interfaces", func() {
			type interfaceStruct struct {
				TestField3 []interface{} `bson:"interfaces"`
			}

			t := []int{1, 2, 3, 4}
			interfaceSlice := make([]interface{}, len(t))
			for i, v := range t {
				interfaceSlice[i] = v
			}

			result := ConvertStructToBSONMap(
				struct {
					TestField1 string          `bson:"testField1"`
					TestField2 interfaceStruct `bson:"nestedStruct"`
				}{
					TestField1: valuesStruct.String,
					TestField2: interfaceStruct{
						TestField3: interfaceSlice,
					},
				}, nil,
			)

			expected := bson.M{
				"testField1": valuesStruct.String,
				"nestedStruct": bson.M{
					"interfaces": interfaceSlice,
				},
			}

			Expect(result).To(Equal(expected))
		})
	})

	// Testing the functionality of a slice of structs
	Context("should correctly parse", func() {
		type valueStruct struct {
			String string         `bson:"str"`
			Slice  []string       `bson:"slice"`
			Map    map[string]int `bson:"map"`
			Time   time.Time      `bson:"time"`
		}

		var valuesStruct valueStruct
		var expectedStruct bson.M

		BeforeEach(func() {
			valuesStruct = valueStruct{
				String: "Test String",
				Slice:  []string{"Test 1", "!@£$%^&*()", ""},
				Map:    map[string]int{"Test 1": 100, "Test 2!": 0, "Test 3": -100},
				Time:   time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
			}
			expectedStruct = bson.M{
				"str":   valuesStruct.String,
				"slice": valuesStruct.Slice,
				"map":   valuesStruct.Map,
				"time":  valuesStruct.Time,
			}
		})

		It("a slice of structs", func() {
			result := ConvertStructToBSONMap(
				struct {
					TestField1 []valueStruct `bson:"sliceStruct"`
				}{
					TestField1: []valueStruct{valuesStruct, valuesStruct},
				}, nil,
			)

			Expect(len(result["sliceStruct"].([]interface{}))).To(Equal(2))
			Expect(result["sliceStruct"].([]interface{})[0]).To(Equal(expectedStruct))
			Expect(result["sliceStruct"].([]interface{})[1]).To(Equal(expectedStruct))
		})

		It("a slice of pointers to structs", func() {
			result := ConvertStructToBSONMap(
				struct {
					TestField1 []*valueStruct `bson:"sliceStruct"`
				}{
					TestField1: []*valueStruct{&valuesStruct, &valuesStruct},
				}, nil,
			)

			Expect(len(result["sliceStruct"].([]interface{}))).To(Equal(2))
			Expect(result["sliceStruct"].([]interface{})[0]).To(Equal(expectedStruct))
			Expect(result["sliceStruct"].([]interface{})[1]).To(Equal(expectedStruct))
		})
	})

	// Testing the functionality of a map of structs
	Context("should correctly parse", func() {
		type valueStruct struct {
			String string         `bson:"str"`
			Slice  []string       `bson:"slice"`
			Map    map[string]int `bson:"map"`
			Time   time.Time      `bson:"time"`
		}

		var valuesStruct valueStruct
		var expectedStruct bson.M

		BeforeEach(func() {
			valuesStruct = valueStruct{
				String: "Test String",
				Slice:  []string{"Test 1", "!@£$%^&*()", ""},
				Map:    map[string]int{"Test 1": 100, "Test 2!": 0, "Test 3": -100},
				Time:   time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
			}
			expectedStruct = bson.M{
				"str":   valuesStruct.String,
				"slice": valuesStruct.Slice,
				"map":   valuesStruct.Map,
				"time":  valuesStruct.Time,
			}
		})

		It("a map of structs", func() {
			result := ConvertStructToBSONMap(
				struct {
					TestField1 map[string]valueStruct `bson:"mapStruct"`
				}{
					TestField1: map[string]valueStruct{"Test 1": valuesStruct, "Test 2": valuesStruct},
				}, nil,
			)

			Expect(len(result["mapStruct"].(bson.M))).To(Equal(2))
			Expect(result["mapStruct"].(bson.M)["Test 1"]).To(Equal(expectedStruct))
			Expect(result["mapStruct"].(bson.M)["Test 2"]).To(Equal(expectedStruct))
		})

		It("a map of pointers to structs", func() {
			result := ConvertStructToBSONMap(
				struct {
					TestField1 map[string]*valueStruct `bson:"mapStruct"`
				}{
					TestField1: map[string]*valueStruct{"Test 1": &valuesStruct, "Test 2": &valuesStruct},
				}, nil,
			)

			Expect(len(result["mapStruct"].(bson.M))).To(Equal(2))
			Expect(result["mapStruct"].(bson.M)["Test 1"]).To(Equal(expectedStruct))
			Expect(result["mapStruct"].(bson.M)["Test 2"]).To(Equal(expectedStruct))
		})
	})
})

var _ = Describe("The package should be able to map", func() {

	type Metadata struct {
		LastActive time.Time `bson:"lastActive"`
	}

	type Characteristics struct {
		LeftHanded bool `bson:"leftHanded"`
		Tall       bool `bson:"tall"`
	}

	type User struct {
		ID              primitive.ObjectID `bson:"_id"`
		FirstName       string             `bson:"firstName"`
		LastName        string             `bson:"lastName,omitempty"`
		DoB             time.Time          `bson:"dob,string"`
		Characteristics *Characteristics   `bson:"characteristics,flatten"`
		Metadata        Metadata           `bson:"metadata,omitnested"`
		Secret          string             `bson:"-"`
		favouriteColor  string
	}

	var user User

	objID, _ := primitive.ObjectIDFromHex("54759eb3c090d83494e2d804")

	BeforeEach(func() {
		user = User{
			ID:        objID,
			FirstName: "Jane",
			LastName:  "",
			DoB:       time.Date(1985, 6, 15, 0, 0, 0, 0, time.UTC),
			Characteristics: &Characteristics{
				LeftHanded: true,
				Tall:       false,
			},
			Metadata:       Metadata{LastActive: time.Date(2019, 7, 23, 14, 0, 0, 0, time.UTC)},
			Secret:         "secret",
			favouriteColor: "blue",
		}
	})

	It("an example user profile, with no options", func() {
		result := ConvertStructToBSONMap(user, nil)
		expected := bson.M{
			"_id":        objID,
			"firstName":  "Jane",
			"dob":        "1985-06-15 00:00:00 +0000 UTC",
			"leftHanded": true,
			"tall":       false,
			"metadata":   Metadata{LastActive: time.Date(2019, 7, 23, 14, 0, 0, 0, time.UTC)},
		}
		Expect(result).To(Equal(expected))
	})

	It("an example user profile, with UseIDifAvailable", func() {
		result := ConvertStructToBSONMap(user, &MappingOpts{UseIDifAvailable: true})
		expected := bson.M{
			"_id": objID,
		}
		Expect(result).To(Equal(expected))
	})

	It("an example user profile, with RemoveID ", func() {
		result := ConvertStructToBSONMap(user, &MappingOpts{RemoveID: true})
		expected := bson.M{
			"firstName":  "Jane",
			"dob":        "1985-06-15 00:00:00 +0000 UTC",
			"leftHanded": true,
			"tall":       false,
			"metadata":   Metadata{LastActive: time.Date(2019, 7, 23, 14, 0, 0, 0, time.UTC)},
		}
		Expect(result).To(Equal(expected))
	})

	It("an example user profile, with GenerateFilterOrPatch ", func() {
		user.Metadata = Metadata{}
		user.ID = primitive.ObjectID{}
		user.DoB = time.Date(1, 1, 1, 0, 0, 0, 0, time.UTC)

		result := ConvertStructToBSONMap(user, &MappingOpts{GenerateFilterOrPatch: true})
		expected := bson.M{
			"firstName":  "Jane",
			"leftHanded": true,
		}
		Expect(result).To(Equal(expected))
	})

	It("an example user profile, with RemoveID & GenerateFilterOrPatch ", func() {
		user.Metadata = Metadata{}
		user.DoB = time.Date(1, 1, 1, 0, 0, 0, 0, time.UTC)

		result := ConvertStructToBSONMap(user, &MappingOpts{RemoveID: true, GenerateFilterOrPatch: true})
		expected := bson.M{
			"firstName":  "Jane",
			"leftHanded": true,
		}
		Expect(result).To(Equal(expected))
	})
})

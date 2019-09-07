package main

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"reflect"
	"time"
)

var _ = Describe("structFields", func() {
	var testStruct *StructToBSON

	It("should correctly return the right size array of structFields", func() {
		testStruct = NewBSONMapperStruct(
			struct {
				TestField1 string    `bson:"testField1"`
				TestField2 int       `bson:"testField2"`
				TestField3 time.Time `bson:"testField3"`
				testField4 float64
				testField5 bool
			}{
				TestField1: "Test String",
				TestField2: 100,
				TestField3: time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
				testField4: 10.1,
				testField5: true,
			})

		result := testStruct.structFields()

		Expect(len(result)).To(Equal(3))
	})

	It("should correctly ignore fields that should be omitted", func() {
		testStruct = NewBSONMapperStruct(
			struct {
				TestField1 string    `bson:"testField1"`
				TestField2 int       `bson:"testField2"`
				TestField3 time.Time `bson:"-"`
				testField4 float64
				testField5 bool
			}{
				TestField1: "Test String",
				TestField2: 100,
				TestField3: time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
				testField4: 10.1,
				testField5: true,
			})

		result := testStruct.structFields()

		Expect(len(result)).To(Equal(2))
	})
})

var _ = Describe("structVal", func() {
	It("should correctly process a struct", func() {
		testStruct := struct {
			TestField1 string
			TestField2 bool
		}{
			TestField1: "Test String",
			TestField2: true,
		}

		result := structVal(testStruct)

		Expect(result.Interface()).To(Equal(reflect.ValueOf(testStruct).Interface()))
	})

	It("should correctly process a pointer to struct", func() {
		testStruct := &struct {
			TestField1 string
			TestField2 bool
		}{
			TestField1: "Test String",
			TestField2: true,
		}

		result := structVal(testStruct)

		Expect(result.Interface()).To(Equal(reflect.ValueOf(testStruct).Elem().Interface()))
	})

	Context("should panic", func() {
		type PanicTestCase struct {
			input interface{}
		}

		DescribeTable("if", func(c PanicTestCase) {
			caseDidPanic := false
			defer func() {
				if r := recover(); r != nil {
					caseDidPanic = true
				}
			}()
			_ = structVal(c.input)
			Expect(caseDidPanic).To(BeTrue())
		},
			Entry("a string is passed", PanicTestCase{input: "Test String"}),
			Entry("an int is passed", PanicTestCase{input: 123}),
			Entry("a bool is passed", PanicTestCase{input: true}),
			Entry("a slice is passed", PanicTestCase{input: []int{1, 2, 3}}),
			Entry("a map is passed", PanicTestCase{input: map[string]struct{}{"Test 1": struct{}{}, "Test 2": struct{}{}}}),
			Entry("a function is passed", PanicTestCase{input: func() {}}),
			Entry("a pointer to a string is passed", PanicTestCase{input: new(string)}),
			Entry("a pointer to an int is passed", PanicTestCase{input: new(int)}),
			Entry("a pointer to a bool is passed", PanicTestCase{input: new(bool)}),
			Entry("a pointer to a slice is passed", PanicTestCase{input: new([]int)}),
			Entry("a pointer to a map is passed", PanicTestCase{input: new(map[string]struct{})}),
			Entry("a pointer to a function is passed", PanicTestCase{input: new(func())}),
		)
	})
})

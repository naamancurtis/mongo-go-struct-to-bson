package mapper

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Tags should", func() {

	Context("use \"Has()\" to check if", func() {
		var tagOpts tagOptions

		BeforeEach(func() {
			tagOpts = tagOptions{}
			tagOpts["TEST_TAG"] = struct{}{}
			tagOpts["Tag with Space"] = struct{}{}
		})

		It("a tag exists", func() {
			result := tagOpts.Has("TEST_TAG")
			Expect(result).To(BeTrue())
		})

		It("a tag doesn't exist", func() {
			result := tagOpts.Has("FAIL")
			Expect(result).To(BeFalse())
		})

		It("a tag exists with spaces in", func() {
			result := tagOpts.Has("Tag with Space")
			Expect(result).To(BeTrue())
		})

		It("an empty string is passed as an argument", func() {
			result := tagOpts.Has("")
			Expect(result).To(BeFalse())
		})
	})

	Context("parse strings into Tags", func() {

		It("if a tag follows the expected format", func() {
			tagName, tagOpts := parseTag("test1,omitempty")
			Expect(tagName).To(Equal("test1"))
			Expect(tagOpts).To(Equal(tagOptions{"omitempty": struct{}{}}))
		})

		It("if a tag is empty", func() {
			tagName, tagOpts := parseTag("")
			Expect(tagName).To(Equal(""))
			Expect(tagOpts).To(Equal(tagOptions{}))
		})

		It("if a tag has no options", func() {
			tagName, tagOpts := parseTag("test1")
			Expect(tagName).To(Equal("test1"))
			Expect(tagOpts).To(Equal(tagOptions{}))
		})

		It("if a tag has multiple options", func() {
			tagName, tagOpts := parseTag("test1,opt1,opt2")
			Expect(tagName).To(Equal("test1"))
			Expect(tagOpts).To(Equal(tagOptions{"opt1": struct{}{}, "opt2": struct{}{}}))
		})
	})
})

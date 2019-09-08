package mapper

import "strings"

type tagOptions map[string]struct{}

// Has checks whether a string is present in the tag options
func (t tagOptions) Has(opt string) bool {
	if _, ok := t[opt]; ok {
		return true
	}
	return false
}

// parseTag parses the tag on a struct field
// it extracts both the name and the options
func parseTag(tag string) (string, tagOptions) {
	res := strings.Split(tag, ",")
	m := make(tagOptions)
	for i, opt := range res {
		if i == 0 {
			continue
		}
		m[opt] = struct{}{}
	}
	return res[0], m
}

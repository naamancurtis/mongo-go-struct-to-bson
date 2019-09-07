package mapper

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestMongoGoStructToBson(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "MongoGoStructToBson Suite")
}

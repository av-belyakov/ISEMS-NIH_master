package mytestpackages_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestMyTestPackages(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "MyTestPackages Suite")
}

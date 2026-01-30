package ui_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestTheme(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Theme Suite")
}

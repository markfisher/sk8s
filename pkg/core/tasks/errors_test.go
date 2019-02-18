package tasks_test

import (
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	"github.com/projectriff/riff/pkg/core/tasks"
)

var _ = Describe("The result handling utility", func() {

	var formatter func(tasks.CorrelatedResult) string

	BeforeEach(func() {
		formatter = func(result tasks.CorrelatedResult) string {
			err := result.Error
			if err == nil {
				return ""
			}
			return fmt.Sprintf("got %s, resulted in error: %s", result.Input, err.Error())
		}
	})

	It("formats the single error message", func() {
		results := []tasks.CorrelatedResult{{Input: "foo", Error: errors.New("nope")}}

		result := tasks.MergeResults(results, formatter)

		Expect(result).To(MatchError("got foo, resulted in error: nope"))
	})

	It("skips empty messages", func() {
		results := []tasks.CorrelatedResult{
			{Input: "foo", Error: nil},
			{Input: "bar", Error: errors.New("nope")},
			{Input: "baz", Error: nil},
		}

		result := tasks.MergeResults(results, formatter)

		Expect(result).To(MatchError("got bar, resulted in error: nope"))
	})

	It("merges multiple error messages", func() {
		results := []tasks.CorrelatedResult{
			{Input: "foo", Error: nil},
			{Input: "bar", Error: errors.New("nope")},
			{Input: "baz", Error: errors.New("still nope")},
		}

		result := tasks.MergeResults(results, formatter)

		Expect(result).To(MatchError(`got bar, resulted in error: nope
got baz, resulted in error: still nope`))
	})

	It("returns nil when there are only nil errors", func() {
		results := []tasks.CorrelatedResult{{Input: "foo", Error: nil}}

		result := tasks.MergeResults(results, formatter)

		Expect(result).To(BeNil())
	})
})

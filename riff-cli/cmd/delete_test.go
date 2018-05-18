package cmd

import (
	"fmt"
	"os"

	"github.com/juju/errgo/errors"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/projectriff/riff/riff-cli/pkg/kubectl"
	"github.com/spf13/cobra"
)

var _ = Describe("The delete command", func() {

	const canned_kubectl_get_response = `{
				"apiVersion": "projectriff.io/v1alpha1",
				"kind": "Link",
				"metadata": {},
				"spec": {
					"function": "%s",
					"input": "myInputTopic",
					"output": "myOutputTopic"
				}
			}`

	var (
		oldCWD        string
		realKubeCtl   *kubectl.MockKubeCtl
		dryRunKubeCtl *kubectl.MockKubeCtl
		deleteCmd     *cobra.Command
		args          []string
	)

	BeforeEach(func() {
		var err error

		oldCWD, err = os.Getwd()
		Expect(err).NotTo(HaveOccurred())

		realKubeCtl = new(kubectl.MockKubeCtl)
		dryRunKubeCtl = new(kubectl.MockKubeCtl)

		deleteCmd, _ = Delete(realKubeCtl, dryRunKubeCtl)
		args = []string{}
	})

	AfterEach(func() {
		realKubeCtl.AssertExpectations(GinkgoT())
		dryRunKubeCtl.AssertExpectations(GinkgoT())

		os.Chdir(oldCWD)
	})

	Context("with no --name provided", func() {
		BeforeEach(func() {
			os.Chdir("../test_data/command/echo")
		})

		It("should delete the function and link based on dirname", func() {
			link := fmt.Sprintf(canned_kubectl_get_response, "echo")
			realKubeCtl.On("Exec", []string{"get", "links.projectriff.io", "echo", "-o", "json"}).Return(link, nil)

			realKubeCtl.On("Exec", []string{"delete", "links.projectriff.io", "echo"}).Return("", nil)
			realKubeCtl.On("Exec", []string{"delete", "functions.projectriff.io", "echo"}).Return("", nil)

			err := deleteCmd.Execute()
			Expect(err).NotTo(HaveOccurred())
		})

		It("should delete the function, topic, and link when run with --all", func() {
			deleteCmd.SetArgs([]string{"--all"})

			link := fmt.Sprintf(canned_kubectl_get_response, "echo")
			realKubeCtl.On("Exec", []string{"get", "links.projectriff.io", "echo", "-o", "json"}).Return(link, nil)

			realKubeCtl.On("Exec", []string{"delete", "links.projectriff.io", "echo"}).Return("", nil)
			realKubeCtl.On("Exec", []string{"delete", "functions.projectriff.io", "echo"}).Return("", nil)
			realKubeCtl.On("Exec", []string{"delete", "topics.projectriff.io", "myInputTopic"}).Return("", nil)
			realKubeCtl.On("Exec", []string{"delete", "topics.projectriff.io", "myOutputTopic"}).Return("", nil)

			err := deleteCmd.Execute()
			Expect(err).NotTo(HaveOccurred())
		})

		It("should delete the function when run with --all and the topics do not exist", func() {
			deleteCmd.SetArgs([]string{"--all"})

			link := fmt.Sprintf(canned_kubectl_get_response, "echo")
			realKubeCtl.On("Exec", []string{"get", "links.projectriff.io", "echo", "-o", "json"}).Return(link, nil)

			realKubeCtl.On("Exec", []string{"delete", "links.projectriff.io", "echo"}).Return("", nil)
			realKubeCtl.On("Exec", []string{"delete", "functions.projectriff.io", "echo"}).Return("", nil)
			realKubeCtl.On("Exec", []string{"delete", "topics.projectriff.io", "myInputTopic"}).Return("", errors.New("Error from server (NotFound): topics.projectriff.io \"myInputTopic\" not found"))
			realKubeCtl.On("Exec", []string{"delete", "topics.projectriff.io", "myOutputTopic"}).Return("", errors.New("Error from server (NotFound): topics.projectriff.io \"myOutputTopic\" not found"))

			err := deleteCmd.Execute()
			Expect(err).NotTo(HaveOccurred())

		})
		It("should delete the function when run with --all and one topic do not exist", func() {
			deleteCmd.SetArgs([]string{"--all"})

			link := fmt.Sprintf(canned_kubectl_get_response, "echo")
			realKubeCtl.On("Exec", []string{"get", "links.projectriff.io", "echo", "-o", "json"}).Return(link, nil)

			realKubeCtl.On("Exec", []string{"delete", "links.projectriff.io", "echo"}).Return("", nil)
			realKubeCtl.On("Exec", []string{"delete", "functions.projectriff.io", "echo"}).Return("", nil)
			realKubeCtl.On("Exec", []string{"delete", "topics.projectriff.io", "myInputTopic"}).Return("", errors.New("Error from server (NotFound): topics.projectriff.io \"myInputTopic\" not found"))
			realKubeCtl.On("Exec", []string{"delete", "topics.projectriff.io", "myOutputTopic"}).Return("", nil)

			err := deleteCmd.Execute()
			Expect(err).NotTo(HaveOccurred())

		})

		Context("when --namespace is set", func() {
			BeforeEach(func() {
				args = append(args, "--namespace", "my-ns")
			})
			It("should delete the function and link based on dirname", func() {
				deleteCmd.SetArgs(args)

				link := fmt.Sprintf(canned_kubectl_get_response, "echo")
				realKubeCtl.On("Exec", []string{"get", "--namespace", "my-ns", "links.projectriff.io", "echo", "-o", "json"}).Return(link, nil)

				realKubeCtl.On("Exec", []string{"delete", "links.projectriff.io", "echo", "--namespace", "my-ns"}).Return("", nil)
				realKubeCtl.On("Exec", []string{"delete", "functions.projectriff.io", "echo", "--namespace", "my-ns"}).Return("", nil)

				err := deleteCmd.Execute()
				Expect(err).NotTo(HaveOccurred())
			})

			It("should delete the function, topic, and link when run with --all", func() {

				args = append(args, "--all")
				deleteCmd.SetArgs(args)

				link := fmt.Sprintf(canned_kubectl_get_response, "echo")
				realKubeCtl.On("Exec", []string{"get", "--namespace", "my-ns", "links.projectriff.io", "echo", "-o", "json"}).Return(link, nil)

				realKubeCtl.On("Exec", []string{"delete", "links.projectriff.io", "echo", "--namespace", "my-ns"}).Return("", nil)
				realKubeCtl.On("Exec", []string{"delete", "functions.projectriff.io", "echo", "--namespace", "my-ns"}).Return("", nil)
				realKubeCtl.On("Exec", []string{"delete", "topics.projectriff.io", "myInputTopic", "--namespace", "my-ns"}).Return("", nil)
				realKubeCtl.On("Exec", []string{"delete", "topics.projectriff.io", "myOutputTopic", "--namespace", "my-ns"}).Return("", nil)

				err := deleteCmd.Execute()
				Expect(err).NotTo(HaveOccurred())
			})

		})
	})

	Context("when --name is provided", func() {
		BeforeEach(func() {
			args = append(args, "--name", "my-function")
		})

		It("should delete the function and link based on name", func() {
			deleteCmd.SetArgs(args)

			link := fmt.Sprintf(canned_kubectl_get_response, "my-function")
			realKubeCtl.On("Exec", []string{"get", "links.projectriff.io", "my-function", "-o", "json"}).Return(link, nil)

			realKubeCtl.On("Exec", []string{"delete", "links.projectriff.io", "my-function"}).Return("", nil)
			realKubeCtl.On("Exec", []string{"delete", "functions.projectriff.io", "my-function"}).Return("", nil)

			err := deleteCmd.Execute()
			Expect(err).NotTo(HaveOccurred())
		})

		It("should delete the function, topic, and link when run with --all", func() {
			args = append(args, "--all")
			deleteCmd.SetArgs(args)

			link := fmt.Sprintf(canned_kubectl_get_response, "my-function")
			realKubeCtl.On("Exec", []string{"get", "links.projectriff.io", "my-function", "-o", "json"}).Return(link, nil)

			realKubeCtl.On("Exec", []string{"delete", "links.projectriff.io", "my-function"}).Return("", nil)
			realKubeCtl.On("Exec", []string{"delete", "functions.projectriff.io", "my-function"}).Return("", nil)
			realKubeCtl.On("Exec", []string{"delete", "topics.projectriff.io", "myInputTopic"}).Return("", nil)
			realKubeCtl.On("Exec", []string{"delete", "topics.projectriff.io", "myOutputTopic"}).Return("", nil)

			err := deleteCmd.Execute()
			Expect(err).NotTo(HaveOccurred())
		})

		Context("when --namespace is set", func() {
			BeforeEach(func() {
				args = append(args, "--namespace", "my-ns")
			})

			It("should delete the function and link based on name", func() {
				deleteCmd.SetArgs(args)

				link := fmt.Sprintf(canned_kubectl_get_response, "my-function")
				realKubeCtl.On("Exec", []string{"get", "--namespace", "my-ns", "links.projectriff.io", "my-function", "-o", "json"}).Return(link, nil)

				realKubeCtl.On("Exec", []string{"delete", "links.projectriff.io", "my-function", "--namespace", "my-ns"}).Return("", nil)
				realKubeCtl.On("Exec", []string{"delete", "functions.projectriff.io", "my-function", "--namespace", "my-ns"}).Return("", nil)

				err := deleteCmd.Execute()
				Expect(err).NotTo(HaveOccurred())
			})

			It("should delete the function, topic, and link when run with --all", func() {
				args = append(args, "--all")
				deleteCmd.SetArgs(args)

				link := fmt.Sprintf(canned_kubectl_get_response, "my-function")
				realKubeCtl.On("Exec", []string{"get", "--namespace", "my-ns", "links.projectriff.io", "my-function", "-o", "json"}).Return(link, nil)

				realKubeCtl.On("Exec", []string{"delete", "links.projectriff.io", "my-function", "--namespace", "my-ns"}).Return("", nil)
				realKubeCtl.On("Exec", []string{"delete", "functions.projectriff.io", "my-function", "--namespace", "my-ns"}).Return("", nil)
				realKubeCtl.On("Exec", []string{"delete", "topics.projectriff.io", "myInputTopic", "--namespace", "my-ns"}).Return("", nil)
				realKubeCtl.On("Exec", []string{"delete", "topics.projectriff.io", "myOutputTopic", "--namespace", "my-ns"}).Return("", nil)

				err := deleteCmd.Execute()
				Expect(err).NotTo(HaveOccurred())
			})

		})
	})

	It("should report kubectl errors", func() {
		deleteCmd.SetArgs([]string{"--name", "whatever"})

		link := fmt.Sprintf(canned_kubectl_get_response, "whatever")
		realKubeCtl.On("Exec", []string{"get", "links.projectriff.io", "whatever", "-o", "json"}).Return(link, nil)

		realKubeCtl.On("Exec", []string{"delete", "links.projectriff.io", "whatever"}).Return("", fmt.Errorf("Whoops"))

		err := deleteCmd.Execute()
		Expect(err).To(MatchError("Whoops"))
	})

	It("should not use the real kubectl client when using --dry-run", func() {
		deleteCmd.SetArgs([]string{"--all", "--name", "whatever", "--dry-run"})

		link := fmt.Sprintf(canned_kubectl_get_response, "whatever")
		realKubeCtl.On("Exec", []string{"get", "links.projectriff.io", "whatever", "-o", "json"}).Return(link, nil)

		dryRunKubeCtl.On("Exec", []string{"delete", "links.projectriff.io", "whatever"}).Return("", nil)
		dryRunKubeCtl.On("Exec", []string{"delete", "functions.projectriff.io", "whatever"}).Return("", nil)
		dryRunKubeCtl.On("Exec", []string{"delete", "topics.projectriff.io", "myInputTopic"}).Return("", nil)
		dryRunKubeCtl.On("Exec", []string{"delete", "topics.projectriff.io", "myOutputTopic"}).Return("", nil)

		err := deleteCmd.Execute()
		Expect(err).NotTo(HaveOccurred())
	})

})

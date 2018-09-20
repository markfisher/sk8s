package commands_test

import (
	"fmt"
	"github.com/knative/eventing/pkg/apis/channels/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	"github.com/projectriff/riff/cmd/commands"
	"github.com/projectriff/riff/pkg/core"
	"github.com/projectriff/riff/pkg/core/mocks"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/mock"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"
)

var _ = Describe("The riff subscription create command", func() {

	var (
		client        core.Client
		clientMock    *mocks.Client
		createCommand *cobra.Command
	)

	BeforeEach(func() {
		client = new(mocks.Client)
		clientMock = client.(*mocks.Client)
		createCommand = commands.SubscriptionCreate(&client)
	})

	AfterEach(func() {
		clientMock.AssertExpectations(GinkgoT())
	})

	It("should be documented", func() {
		Expect(createCommand.Name()).To(Equal("create"))
		Expect(createCommand.Short).NotTo(BeEmpty(), "missing short description")
		Expect(createCommand.Long).NotTo(BeEmpty(), "missing long description")
		Expect(createCommand.Example).NotTo(BeEmpty(), "missing example")
	})

	It("should define flags", func() {
		Expect(createCommand.Flag("subscriber")).NotTo(BeNil())
		Expect(createCommand.Flag("channel")).NotTo(BeNil())
		Expect(createCommand.Flag("reply-to")).NotTo(BeNil())
		Expect(createCommand.Flag("namespace")).NotTo(BeNil())
	})

	Context("when given wrong args or flags", func() {

		It("should fail with missing required flags", func() {
			createCommand.SetArgs([]string{})

			err := createCommand.Execute()

			Expect(err).To(MatchError(`required flag(s) "channel", "subscriber" not set`))
		})

		It("should fail with too many args", func() {
			createCommand.SetArgs([]string{
				"too", "much", "--subscriber", "service", "--channel", "input"})

			err := createCommand.Execute()

			Expect(err).To(MatchError(`accepts at most 1 arg(s), received 2`))
		})

		It("should fail with an invalid subscription name", func() {
			createCommand.SetArgs([]string{
				"@@invalid@@", "--subscriber", "service", "--channel", "input"})

			err := createCommand.Execute()

			Expect(err).To(MatchError(HavePrefix("a DNS-1123 subdomain must consist")))
		})
	})

	Context("when given valid args and flags", func() {
		It("should create the subscription with the provided name", func() {
			stdout := &strings.Builder{}
			createCommand.SetOutput(stdout)
			createCommand.SetArgs([]string{
				"subscription-name", "--channel", "coco-chanel", "--subscriber", "my-service"})
			clientMock.On("CreateSubscription", core.CreateSubscriptionOptions{
				Name:       "subscription-name",
				Subscriber: "my-service",
				Channel:    "coco-chanel",
			}).Return(nil, nil)

			err := createCommand.Execute()

			Expect(err).NotTo(HaveOccurred())
			Expect(stdout.String()).To(Equal("create completed successfully\n"))
		})

		It("should create the subscription with the service name by default", func() {
			stdout := &strings.Builder{}
			createCommand.SetOutput(stdout)
			createCommand.SetArgs([]string{
				"--channel", "coco-chanel", "--subscriber", "my-service"})
			clientMock.On("CreateSubscription", core.CreateSubscriptionOptions{
				Name:       "my-service",
				Subscriber: "my-service",
				Channel:    "coco-chanel",
			}).Return(nil, nil)

			err := createCommand.Execute()

			Expect(err).NotTo(HaveOccurred())
			Expect(stdout.String()).To(Equal("create completed successfully\n"))
		})

		It("should create the subscription with the output channel binding", func() {
			stdout := &strings.Builder{}
			createCommand.SetOutput(stdout)
			createCommand.SetArgs([]string{
				"subscription-name", "--channel", "coco-chanel", "--subscriber", "my-service",
				"--reply-to", "chanel-number-five"})
			clientMock.On("CreateSubscription", core.CreateSubscriptionOptions{
				Name:       "subscription-name",
				Subscriber: "my-service",
				Channel:    "coco-chanel",
				ReplyTo:    "chanel-number-five",
			}).Return(nil, nil)

			err := createCommand.Execute()

			Expect(err).NotTo(HaveOccurred())
			Expect(stdout.String()).To(Equal("create completed successfully\n"))
		})

		It("should create the subscription in the provided namespace", func() {
			stdout := &strings.Builder{}
			createCommand.SetOutput(stdout)
			createCommand.SetArgs([]string{
				"subscription-name",
				"--channel", "coco-chanel",
				"--subscriber", "my-service",
				"--reply-to", "chanel-number-five",
				"--namespace", "myspace"})
			expectedOptions := core.CreateSubscriptionOptions{
				Name:       "subscription-name",
				Subscriber: "my-service",
				Channel:    "coco-chanel",
				ReplyTo:    "chanel-number-five",
			}
			expectedOptions.Namespace = "myspace"
			clientMock.On("CreateSubscription", expectedOptions).Return(nil, nil)

			err := createCommand.Execute()

			Expect(err).NotTo(HaveOccurred())
			Expect(stdout.String()).To(Equal("create completed successfully\n"))
		})

		It("should propagate client error", func() {
			stdout := &strings.Builder{}
			createCommand.SetOutput(stdout)
			createCommand.SetArgs([]string{
				"--channel", "coco-chanel", "--subscriber", "my-service"})
			expectedError := errors.New("client failure")
			clientMock.On("CreateSubscription", core.CreateSubscriptionOptions{
				Name:       "my-service",
				Subscriber: "my-service",
				Channel:    "coco-chanel",
			}).Return(nil, expectedError)

			err := createCommand.Execute()

			Expect(err).To(MatchError(expectedError))
		})
	})

})


var _ = Describe("The riff subscription delete command", func() {

	var (
		client        core.Client
		clientMock    *mocks.Client
		deleteCommand *cobra.Command
	)

	BeforeEach(func() {
		client = new(mocks.Client)
		clientMock = client.(*mocks.Client)
		deleteCommand = commands.SubscriptionDelete(&client)
	})

	AfterEach(func() {
		clientMock.AssertExpectations(GinkgoT())
	})

	It("should be documented", func() {
		Expect(deleteCommand.Name()).To(Equal("delete"))
		Expect(deleteCommand.Short).NotTo(BeEmpty())
		Expect(deleteCommand.Example).NotTo(BeEmpty())
	})

	Context("when given wrong args or flags", func() {
		It("should fail if the number of arguments is incorrect", func() {
			err := deleteCommand.Execute()

			Expect(err).To(MatchError(Equal("accepts 1 arg(s), received 0")))
		})

		It("should fail if the required argument is invalid", func() {
			deleteCommand.SetArgs([]string{"@@invalid@@"})
			err := deleteCommand.Execute()

			Expect(err).To(MatchError(ContainSubstring("must start and end with an alphanumeric character")))
		})
	})

	Context("when given valid args and flags", func() {
		It("should unsubscribe based on the subscription name", func() {
			stdout := &strings.Builder{}
			deleteCommand.SetOutput(stdout)
			deleteCommand.SetArgs([]string{"subscription-name"})
			clientMock.On("DeleteSubscription", core.DeleteSubscriptionOptions{
				Name: "subscription-name",
			}).Return(nil)

			err := deleteCommand.Execute()

			Expect(err).NotTo(HaveOccurred())
			Expect(stdout.String()).To(Equal("delete completed successfully\n"))
		})

		It("should unsubscribe based on the subscription name and namespace", func() {
			stdout := &strings.Builder{}
			deleteCommand.SetOutput(stdout)
			deleteCommand.SetArgs([]string{"subscription-name", "--namespace", "ns"})
			options := core.DeleteSubscriptionOptions{Name: "subscription-name",}
			options.Namespace = "ns"
			clientMock.On("DeleteSubscription", options).Return(nil)

			err := deleteCommand.Execute()

			Expect(err).NotTo(HaveOccurred())
			Expect(stdout.String()).To(Equal("delete completed successfully\n"))
		})

		It("should propagate the client error", func() {
			deleteCommand.SetArgs([]string{"subscription-name"})
			clientMock.On("DeleteSubscription", mock.Anything).Return(fmt.Errorf("client error"))

			err := deleteCommand.Execute()

			Expect(err).To(MatchError(Equal("client error")))
		})
	})

})

const (
	listOutput = `NAME            CHANNEL SUBSCRIBER REPLY-TO 
my-subscription channel subscriber reply-to 

list completed successfully
`
	emptyListOutput = `NAME CHANNEL SUBSCRIBER REPLY-TO 

list completed successfully
`
	)

var _ = Describe("The riff subscription list command", func() {
	var (
		client      core.Client
		clientMock  *mocks.Client
		listCommand *cobra.Command
	)

	BeforeEach(func() {
		client = new(mocks.Client)
		clientMock = client.(*mocks.Client)
		listCommand = commands.SubscriptionList(&client)
	})

	AfterEach(func() {
		clientMock.AssertExpectations(GinkgoT())
	})

	It("should be documented", func() {
		Expect(listCommand.Name()).To(Equal("list"))
		Expect(listCommand.Short).NotTo(BeEmpty())
		Expect(listCommand.Example).NotTo(BeEmpty())
	})

	Context("when given wrong args or flags", func() {
		It("should fail if extra arguments are passed", func() {
			listCommand.SetArgs([]string{"extra-arg"})
			err := listCommand.Execute()

			Expect(err).To(MatchError(Equal("accepts 0 arg(s), received 1")))
		})
	})

	Context("when given valid args and flags", func() {

		var (
			out *strings.Builder
		)

		BeforeEach(func() {
			out = &strings.Builder{}
			listCommand.SetOutput(out)
		})

		It("should print the list of subscriptions", func() {
			subscriptions := listOf(v1alpha1.Subscription{
				ObjectMeta: metav1.ObjectMeta{
					Name: "my-subscription",
				},
				Spec: v1alpha1.SubscriptionSpec{
					Channel:    "channel",
					Subscriber: "subscriber",
					ReplyTo:    "reply-to",
				},
			})
			clientMock.On("ListSubscriptions", core.ListSubscriptionsOptions{}).Return(subscriptions, nil)

			err := listCommand.Execute()

			Expect(err).To(BeNil())
			Expect(out.String()).To(Equal(listOutput))
		})

		It("should print the list of subscription in the provided namespace", func() {
			namespace := "ns"
			listCommand.SetArgs([]string{"--namespace", namespace})
			subscriptions := listOf(v1alpha1.Subscription{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "my-subscription",
					Namespace: namespace,
				},
				Spec: v1alpha1.SubscriptionSpec{
					Channel:    "channel",
					Subscriber: "subscriber",
					ReplyTo:    "reply-to",
				},
			})
			options := core.ListSubscriptionsOptions{}
			options.Namespace = namespace
			clientMock.On("ListSubscriptions", options).Return(subscriptions, nil)

			err := listCommand.Execute()

			Expect(err).To(BeNil())
			Expect(out.String()).To(Equal(listOutput))
		})

		It("should space out headers when the list is empty", func() {
			clientMock.On("ListSubscriptions", core.ListSubscriptionsOptions{}).Return(
				&v1alpha1.SubscriptionList{}, nil)

			err := listCommand.Execute()

			Expect(err).To(BeNil())
			s := out.String()
			Expect(s).To(Equal(emptyListOutput))
		})

		It("should propagate the client error", func() {
			clientMock.On("ListSubscriptions", mock.Anything).Return(
				&v1alpha1.SubscriptionList{},
				fmt.Errorf("client error"))

			err := listCommand.Execute()

			Expect(err).To(MatchError(Equal("client error")))
		})
	})
})

func listOf(subscriptions ...v1alpha1.Subscription) *v1alpha1.SubscriptionList {
	clientResult := v1alpha1.SubscriptionList{}
	clientResult.Items = subscriptions
	return &clientResult
}

package kubernetes

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
)

var (
	cfg       *rest.Config
	clientset kubernetes.Interface
	testEnv   *envtest.Environment
)

func TestInformerIntegration(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "Informer Integration Suite")
}

var _ = ginkgo.BeforeSuite(func() {
	ginkgo.By("bootstrapping test environment")
	
	// Use default envtest behavior which should work with installed binaries
	testEnv = &envtest.Environment{
		CRDDirectoryPaths:     []string{},
		ErrorIfCRDPathMissing: false,
		// Use KUBEBUILDER_ASSETS env var if set, otherwise use default discovery
		UseExistingCluster: func() *bool {
			// Only use existing cluster if explicitly set
			if os.Getenv("USE_EXISTING_CLUSTER") == "true" {
				val := true
				return &val
			}
			return nil
		}(),
	}

	var err error
	cfg, err = testEnv.Start()
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	gomega.Expect(cfg).NotTo(gomega.BeNil())

	clientset, err = kubernetes.NewForConfig(cfg)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
})

var _ = ginkgo.AfterSuite(func() {
	ginkgo.By("tearing down the test environment")
	err := testEnv.Stop()
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
})

var _ = ginkgo.Describe("DeploymentInformer", func() {
	var (
		informer   *DeploymentInformer
		namespace  string
		deployment *appsv1.Deployment
	)

	ginkgo.BeforeEach(func() {
		namespace = "default"
		informer = NewDeploymentInformer(clientset, namespace, 1*time.Second)
		
		// Create a test deployment
		deployment = &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-deployment",
				Namespace: namespace,
			},
			Spec: appsv1.DeploymentSpec{
				Replicas: int32Ptr(1),
				Selector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"app": "test",
					},
				},
				Template: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							"app": "test",
						},
					},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name:  "test-container",
								Image: "nginx:latest",
							},
						},
					},
				},
			},
		}
	})

	ginkgo.AfterEach(func() {
		if informer.IsStarted() {
			informer.Stop()
		}
		
		// Clean up deployment
		err := clientset.AppsV1().Deployments(namespace).Delete(
			context.TODO(),
			deployment.Name,
			metav1.DeleteOptions{},
		)
		if err != nil {
			ginkgo.GinkgoLogr.Info("Failed to delete test deployment", "error", err)
		}
	})

	ginkgo.Context("when creating and starting informer", func() {
		ginkgo.It("should create informer with correct configuration", func() {
			gomega.Expect(informer.namespace).To(gomega.Equal(namespace))
			gomega.Expect(informer.resyncPeriod).To(gomega.Equal(1 * time.Second))
			gomega.Expect(informer.IsStarted()).To(gomega.BeFalse())
		})

		ginkgo.It("should start and stop informer successfully", func() {
			err := informer.Start()
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(informer.IsStarted()).To(gomega.BeTrue())

			informer.Stop()
			gomega.Expect(informer.IsStarted()).To(gomega.BeFalse())
		})

		ginkgo.It("should return error when starting already started informer", func() {
			err := informer.Start()
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			err = informer.Start()
			gomega.Expect(err).To(gomega.HaveOccurred())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("informer is already started"))
		})
	})

	ginkgo.Context("when working with deployments", func() {
		ginkgo.BeforeEach(func() {
			err := informer.Start()
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
		})

		ginkgo.It("should detect deployment creation", func() {
			eventHandler := &TestEventHandler{}
			informer.AddEventHandler(eventHandler)

			// Create deployment
			_, err := clientset.AppsV1().Deployments(namespace).Create(
				context.TODO(),
				deployment,
				metav1.CreateOptions{},
			)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			// Wait for event
			gomega.Eventually(func() bool {
				return eventHandler.GetOnAddCalled()
			}, 10*time.Second, 100*time.Millisecond).Should(gomega.BeTrue())

			lastDep := eventHandler.GetLastDeployment()
			gomega.Expect(lastDep).NotTo(gomega.BeNil())
			gomega.Expect(lastDep.Name).To(gomega.Equal("test-deployment"))
		})

		ginkgo.It("should list deployments from cache", func() {
			// Create deployment
			_, err := clientset.AppsV1().Deployments(namespace).Create(
				context.TODO(),
				deployment,
				metav1.CreateOptions{},
			)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			// Wait for cache sync
			gomega.Eventually(func() ([]*appsv1.Deployment, error) {
				return informer.ListDeployments()
			}, 10*time.Second, 100*time.Millisecond).Should(gomega.HaveLen(1))

			deployments, err := informer.ListDeployments()
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(deployments).To(gomega.HaveLen(1))
			gomega.Expect(deployments[0].Name).To(gomega.Equal("test-deployment"))
		})

		ginkgo.It("should get deployment by name from cache", func() {
			// Create deployment
			_, err := clientset.AppsV1().Deployments(namespace).Create(
				context.TODO(),
				deployment,
				metav1.CreateOptions{},
			)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			// Wait for cache sync
			gomega.Eventually(func() error {
				_, err := informer.GetDeployment(namespace, "test-deployment")
				return err
			}, 10*time.Second, 100*time.Millisecond).Should(gomega.Succeed())

			dep, err := informer.GetDeployment(namespace, "test-deployment")
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(dep.Name).To(gomega.Equal("test-deployment"))
		})

		ginkgo.It("should detect deployment deletion", func() {
			eventHandler := &TestEventHandler{}
			informer.AddEventHandler(eventHandler)

			// Create deployment
			_, err := clientset.AppsV1().Deployments(namespace).Create(
				context.TODO(),
				deployment,
				metav1.CreateOptions{},
			)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			// Wait for add event
			gomega.Eventually(func() bool {
				return eventHandler.GetOnAddCalled()
			}, 10*time.Second, 100*time.Millisecond).Should(gomega.BeTrue())

			// Delete deployment
			err = clientset.AppsV1().Deployments(namespace).Delete(
				context.TODO(),
				deployment.Name,
				metav1.DeleteOptions{},
			)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			// Wait for delete event
			gomega.Eventually(func() bool {
				return eventHandler.GetOnDeleteCalled()
			}, 10*time.Second, 100*time.Millisecond).Should(gomega.BeTrue())
		})

		ginkgo.It("should detect deployment updates", func() {
			eventHandler := &TestEventHandler{}
			informer.AddEventHandler(eventHandler)

			// Create deployment
			createdDep, err := clientset.AppsV1().Deployments(namespace).Create(
				context.TODO(),
				deployment,
				metav1.CreateOptions{},
			)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			// Wait for add event
			gomega.Eventually(func() bool {
				return eventHandler.GetOnAddCalled()
			}, 10*time.Second, 100*time.Millisecond).Should(gomega.BeTrue())

			// Update deployment
			createdDep.Spec.Replicas = int32Ptr(2)
			_, err = clientset.AppsV1().Deployments(namespace).Update(
				context.TODO(),
				createdDep,
				metav1.UpdateOptions{},
			)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			// Wait for update event
			gomega.Eventually(func() bool {
				return eventHandler.GetOnUpdateCalled()
			}, 10*time.Second, 100*time.Millisecond).Should(gomega.BeTrue())

			lastDep := eventHandler.GetLastDeployment()
			gomega.Expect(lastDep).NotTo(gomega.BeNil())
			gomega.Expect(*lastDep.Spec.Replicas).To(gomega.Equal(int32(2)))
		})
	})

	ginkgo.Context("when informer is not started", func() {
		ginkgo.It("should return error when trying to list deployments", func() {
			deployments, err := informer.ListDeployments()
			gomega.Expect(err).To(gomega.HaveOccurred())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("informer is not started"))
			gomega.Expect(deployments).To(gomega.BeNil())
		})

		ginkgo.It("should return error when trying to get deployment", func() {
			dep, err := informer.GetDeployment(namespace, "test-deployment")
			gomega.Expect(err).To(gomega.HaveOccurred())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("informer is not started"))
			gomega.Expect(dep).To(gomega.BeNil())
		})
	})
})

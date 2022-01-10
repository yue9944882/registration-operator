package integration

import (
	"context"
	"fmt"
	"time"

	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	v1 "open-cluster-management.io/api/operator/v1"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/util/cert"

	"github.com/openshift/library-go/pkg/controller/controllercmd"

	"open-cluster-management.io/registration-operator/pkg/helpers"
	"open-cluster-management.io/registration-operator/pkg/operators/clustermanager"
	certrotation "open-cluster-management.io/registration-operator/pkg/operators/clustermanager/controllers/certrotationcontroller"
	"open-cluster-management.io/registration-operator/test/integration/util"
)

func startHubOperator(ctx context.Context, mode v1.InstallMode) {
	certrotation.SigningCertValidity = time.Second * 30
	certrotation.TargetCertValidity = time.Second * 10
	certrotation.ResyncInterval = time.Second * 1

	var config *rest.Config
	switch mode {
	case v1.InstallModeDefault:
		config = restConfig
	case v1.InstallModeDetached:
		config = detachedRestConfig
	}

	o := &clustermanager.Options{}
	err := o.RunClusterManagerOperator(ctx, &controllercmd.ControllerContext{
		KubeConfig:    config,
		EventRecorder: util.NewIntegrationTestEventRecorder("integration"),
	})
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
}

var _ = ginkgo.Describe("ClusterManager Default Mode", func() {
	var cancel context.CancelFunc
	var hubRegistrationDeployment = fmt.Sprintf("%s-registration-controller", clusterManagerName)

	ginkgo.BeforeEach(func() {
		var ctx context.Context
		ctx, cancel = context.WithCancel(context.Background())
		go startHubOperator(ctx, v1.InstallModeDefault)
	})

	ginkgo.AfterEach(func() {
		if cancel != nil {
			cancel()
		}
	})

	ginkgo.Context("Deploy and clean hub component", func() {
		ginkgo.It("should have expected resource created successfully", func() {
			// Check namespace
			gomega.Eventually(func() error {
				if _, err := kubeClient.CoreV1().Namespaces().Get(context.Background(), hubNamespace, metav1.GetOptions{}); err != nil {
					return err
				}
				return nil
			}, eventuallyTimeout, eventuallyInterval).Should(gomega.BeNil())

			// Check clusterrole/clusterrolebinding
			hubRegistrationClusterRole := fmt.Sprintf("open-cluster-management:%s-registration:controller", clusterManagerName)
			hubRegistrationWebhookClusterRole := fmt.Sprintf("open-cluster-management:%s-registration:webhook", clusterManagerName)
			hubWorkWebhookClusterRole := fmt.Sprintf("open-cluster-management:%s-registration:webhook", clusterManagerName)
			gomega.Eventually(func() error {
				if _, err := kubeClient.RbacV1().ClusterRoles().Get(context.Background(), hubRegistrationClusterRole, metav1.GetOptions{}); err != nil {
					return err
				}
				return nil
			}, eventuallyTimeout, eventuallyInterval).Should(gomega.BeNil())
			gomega.Eventually(func() error {
				if _, err := kubeClient.RbacV1().ClusterRoles().Get(context.Background(), hubRegistrationWebhookClusterRole, metav1.GetOptions{}); err != nil {
					return err
				}
				return nil
			}, eventuallyTimeout, eventuallyInterval).Should(gomega.BeNil())
			gomega.Eventually(func() error {
				if _, err := kubeClient.RbacV1().ClusterRoles().Get(context.Background(), hubWorkWebhookClusterRole, metav1.GetOptions{}); err != nil {
					return err
				}
				return nil
			}, eventuallyTimeout, eventuallyInterval).Should(gomega.BeNil())
			gomega.Eventually(func() error {
				if _, err := kubeClient.RbacV1().ClusterRoleBindings().Get(context.Background(), hubRegistrationClusterRole, metav1.GetOptions{}); err != nil {
					return err
				}
				return nil
			}, eventuallyTimeout, eventuallyInterval).Should(gomega.BeNil())
			gomega.Eventually(func() error {
				if _, err := kubeClient.RbacV1().ClusterRoleBindings().Get(context.Background(), hubRegistrationWebhookClusterRole, metav1.GetOptions{}); err != nil {
					return err
				}
				return nil
			}, eventuallyTimeout, eventuallyInterval).Should(gomega.BeNil())
			gomega.Eventually(func() error {
				if _, err := kubeClient.RbacV1().ClusterRoleBindings().Get(context.Background(), hubWorkWebhookClusterRole, metav1.GetOptions{}); err != nil {
					return err
				}
				return nil
			}, eventuallyTimeout, eventuallyInterval).Should(gomega.BeNil())

			// Check service account
			hubRegistrationSA := fmt.Sprintf("%s-registration-controller-sa", clusterManagerName)
			hubRegistrationWebhookSA := fmt.Sprintf("%s-registration-webhook-sa", clusterManagerName)
			hubWorkWebhookSA := fmt.Sprintf("%s-work-webhook-sa", clusterManagerName)
			gomega.Eventually(func() error {
				if _, err := kubeClient.CoreV1().ServiceAccounts(hubNamespace).Get(context.Background(), hubRegistrationSA, metav1.GetOptions{}); err != nil {
					return err
				}
				return nil
			}, eventuallyTimeout, eventuallyInterval).Should(gomega.BeNil())
			gomega.Eventually(func() error {
				if _, err := kubeClient.CoreV1().ServiceAccounts(hubNamespace).Get(context.Background(), hubRegistrationWebhookSA, metav1.GetOptions{}); err != nil {
					return err
				}
				return nil
			}, eventuallyTimeout, eventuallyInterval).Should(gomega.BeNil())
			gomega.Eventually(func() error {
				if _, err := kubeClient.CoreV1().ServiceAccounts(hubNamespace).Get(context.Background(), hubWorkWebhookSA, metav1.GetOptions{}); err != nil {
					return err
				}
				return nil
			}, eventuallyTimeout, eventuallyInterval).Should(gomega.BeNil())

			// Check deployment
			gomega.Eventually(func() error {
				if _, err := kubeClient.AppsV1().Deployments(hubNamespace).Get(context.Background(), hubRegistrationDeployment, metav1.GetOptions{}); err != nil {
					return err
				}
				return nil
			}, eventuallyTimeout, eventuallyInterval).Should(gomega.BeNil())

			hubRegistrationWebhookDeployment := fmt.Sprintf("%s-registration-webhook", clusterManagerName)
			gomega.Eventually(func() error {
				if _, err := kubeClient.AppsV1().Deployments(hubNamespace).Get(context.Background(), hubRegistrationWebhookDeployment, metav1.GetOptions{}); err != nil {
					return err
				}
				return nil
			}, eventuallyTimeout, eventuallyInterval).Should(gomega.BeNil())

			hubWorkWebhookDeployment := fmt.Sprintf("%s-work-webhook", clusterManagerName)
			gomega.Eventually(func() error {
				if _, err := kubeClient.AppsV1().Deployments(hubNamespace).Get(context.Background(), hubWorkWebhookDeployment, metav1.GetOptions{}); err != nil {
					return err
				}
				return nil
			}, eventuallyTimeout, eventuallyInterval).Should(gomega.BeNil())

			// Check service
			gomega.Eventually(func() error {
				if _, err := kubeClient.CoreV1().Services(hubNamespace).Get(context.Background(), "cluster-manager-registration-webhook", metav1.GetOptions{}); err != nil {
					return err
				}
				return nil
			}, eventuallyTimeout, eventuallyInterval).Should(gomega.BeNil())

			gomega.Eventually(func() error {
				if _, err := kubeClient.CoreV1().Services(hubNamespace).Get(context.Background(), "cluster-manager-work-webhook", metav1.GetOptions{}); err != nil {
					return err
				}
				return nil
			}, eventuallyTimeout, eventuallyInterval).Should(gomega.BeNil())

			// Check webhook secret
			registrationWebhookSecret := "registration-webhook-serving-cert"
			gomega.Eventually(func() error {
				s, err := kubeClient.CoreV1().Secrets(hubNamespace).Get(context.Background(), registrationWebhookSecret, metav1.GetOptions{})
				if err != nil {
					return err
				}
				if s.Data == nil {
					return fmt.Errorf("s.Data is nil")
				} else if s.Data["tls.crt"] == nil {
					return fmt.Errorf("s.Data doesn't contain key 'tls.crt'")
				} else if s.Data["tls.key"] == nil {
					return fmt.Errorf("s.Data doesn't contain key 'tls.key'")
				}
				return nil
			}, eventuallyTimeout, eventuallyInterval).Should(gomega.BeNil())

			workWebhookSecret := "work-webhook-serving-cert"
			gomega.Eventually(func() error {
				s, err := kubeClient.CoreV1().Secrets(hubNamespace).Get(context.Background(), workWebhookSecret, metav1.GetOptions{})
				if err != nil {
					return err
				}
				if s.Data == nil {
					return fmt.Errorf("s.Data is nil")
				} else if s.Data["tls.crt"] == nil {
					return fmt.Errorf("s.Data doesn't contain key 'tls.crt'")
				} else if s.Data["tls.key"] == nil {
					return fmt.Errorf("s.Data doesn't contain key 'tls.key'")
				}
				return nil
			}, eventuallyTimeout, eventuallyInterval).Should(gomega.BeNil())

			// Check validating webhook
			registrationValidtingWebhook := "managedclustervalidators.admission.cluster.open-cluster-management.io"
			gomega.Eventually(func() error {
				if _, err := kubeClient.AdmissionregistrationV1().ValidatingWebhookConfigurations().Get(context.Background(), registrationValidtingWebhook, metav1.GetOptions{}); err != nil {
					return err
				}
				return nil
			}, eventuallyTimeout, eventuallyInterval).Should(gomega.BeNil())

			workValidtingWebhook := "manifestworkvalidators.admission.work.open-cluster-management.io"
			gomega.Eventually(func() error {
				if _, err := kubeClient.AdmissionregistrationV1().ValidatingWebhookConfigurations().Get(context.Background(), workValidtingWebhook, metav1.GetOptions{}); err != nil {
					return err
				}
				return nil
			}, eventuallyTimeout, eventuallyInterval).Should(gomega.BeNil())

			util.AssertClusterManagerCondition(clusterManagerName, operatorClient, "Applied", "ClusterManagerApplied", metav1.ConditionTrue)
		})

		ginkgo.It("Deployment should be updated when clustermanager is changed", func() {
			gomega.Eventually(func() error {
				if _, err := kubeClient.AppsV1().Deployments(hubNamespace).Get(context.Background(), hubRegistrationDeployment, metav1.GetOptions{}); err != nil {
					return err
				}
				return nil
			}, eventuallyTimeout, eventuallyInterval).Should(gomega.BeNil())

			// Check if generations are correct
			gomega.Eventually(func() error {
				actual, err := operatorClient.OperatorV1().ClusterManagers().Get(context.Background(), clusterManagerName, metav1.GetOptions{})
				if err != nil {
					return err
				}

				if actual.Generation != actual.Status.ObservedGeneration {
					return fmt.Errorf("except generation to be %d, but got %d", actual.Status.ObservedGeneration, actual.Generation)
				}

				return nil
			}, eventuallyTimeout, eventuallyInterval).Should(gomega.BeNil())

			clusterManager, err := operatorClient.OperatorV1().ClusterManagers().Get(context.Background(), clusterManagerName, metav1.GetOptions{})
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			clusterManager.Spec.RegistrationImagePullSpec = "testimage:latest"
			_, err = operatorClient.OperatorV1().ClusterManagers().Update(context.Background(), clusterManager, metav1.UpdateOptions{})
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			gomega.Eventually(func() error {
				actual, err := kubeClient.AppsV1().Deployments(hubNamespace).Get(context.Background(), hubRegistrationDeployment, metav1.GetOptions{})
				if err != nil {
					return err
				}
				gomega.Expect(len(actual.Spec.Template.Spec.Containers)).Should(gomega.Equal(1))
				if actual.Spec.Template.Spec.Containers[0].Image != "testimage:latest" {
					return fmt.Errorf("expected image to be testimage:latest but get %s", actual.Spec.Template.Spec.Containers[0].Image)
				}
				return nil
			}, eventuallyTimeout, eventuallyInterval).Should(gomega.BeNil())

			// Check if generations are correct
			gomega.Eventually(func() error {
				actual, err := operatorClient.OperatorV1().ClusterManagers().Get(context.Background(), clusterManagerName, metav1.GetOptions{})
				if err != nil {
					return err
				}

				if actual.Generation != actual.Status.ObservedGeneration {
					return fmt.Errorf("except generation to be %d, but got %d", actual.Status.ObservedGeneration, actual.Generation)
				}

				return nil
			}, eventuallyTimeout, eventuallyInterval).Should(gomega.BeNil())

			// Check if relatedResources are correct
			gomega.Eventually(func() error {
				actual, err := operatorClient.OperatorV1().ClusterManagers().Get(context.Background(), clusterManagerName, metav1.GetOptions{})
				if err != nil {
					return err
				}
				if len(actual.Status.RelatedResources) != 33 {
					return fmt.Errorf("should get 33 relatedResources, actual got %v", len(actual.Status.RelatedResources))
				}
				return nil
			}, eventuallyTimeout, eventuallyInterval).ShouldNot(gomega.HaveOccurred())

		})

		ginkgo.It("Deployment should be added nodeSelector and toleration when add nodePlacement into clustermanager", func() {
			clusterManager, err := operatorClient.OperatorV1().ClusterManagers().Get(context.Background(), clusterManagerName, metav1.GetOptions{})
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			clusterManager.Spec.NodePlacement = v1.NodePlacement{
				NodeSelector: map[string]string{"node-role.kubernetes.io/infra": ""},
				Tolerations: []corev1.Toleration{
					{
						Key:      "node-role.kubernetes.io/infra",
						Operator: corev1.TolerationOpExists,
						Effect:   corev1.TaintEffectNoSchedule,
					},
				},
			}
			_, err = operatorClient.OperatorV1().ClusterManagers().Update(context.Background(), clusterManager, metav1.UpdateOptions{})
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			gomega.Eventually(func() error {
				actual, err := kubeClient.AppsV1().Deployments(hubNamespace).Get(context.Background(), hubRegistrationDeployment, metav1.GetOptions{})
				if err != nil {
					return err
				}
				gomega.Expect(len(actual.Spec.Template.Spec.Containers)).Should(gomega.Equal(1))
				if len(actual.Spec.Template.Spec.NodeSelector) == 0 {
					return fmt.Errorf("length of node selector should not equals to 0")
				}
				if _, ok := actual.Spec.Template.Spec.NodeSelector["node-role.kubernetes.io/infra"]; !ok {
					return fmt.Errorf("node-role.kubernetes.io/infra not exist")
				}
				if len(actual.Spec.Template.Spec.Tolerations) == 0 {
					return fmt.Errorf("length of node selecor should not equals to 0")
				}
				for _, toleration := range actual.Spec.Template.Spec.Tolerations {
					if toleration.Key == "node-role.kubernetes.io/infra" {
						return nil
					}
				}

				return fmt.Errorf("no key equals to node-role.kubernetes.io/infra")
			}, eventuallyTimeout, eventuallyInterval).Should(gomega.BeNil())
		})
		ginkgo.It("Deployment should be reconciled when manually updated", func() {
			gomega.Eventually(func() error {
				if _, err := kubeClient.AppsV1().Deployments(hubNamespace).Get(context.Background(), hubRegistrationDeployment, metav1.GetOptions{}); err != nil {
					return err
				}
				return nil
			}, eventuallyTimeout, eventuallyInterval).Should(gomega.BeNil())
			registrationoDeployment, err := kubeClient.AppsV1().Deployments(hubNamespace).Get(context.Background(), hubRegistrationDeployment, metav1.GetOptions{})
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			registrationoDeployment.Spec.Template.Spec.Containers[0].Image = "testimage2:latest"
			_, err = kubeClient.AppsV1().Deployments(hubNamespace).Update(context.Background(), registrationoDeployment, metav1.UpdateOptions{})
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Eventually(func() error {
				registrationoDeployment, err := kubeClient.AppsV1().Deployments(hubNamespace).Get(context.Background(), hubRegistrationDeployment, metav1.GetOptions{})
				if err != nil {
					return appsv1.ErrInvalidLengthGenerated
				}
				if registrationoDeployment.Spec.Template.Spec.Containers[0].Image != "testimage:latest" {
					return fmt.Errorf("image should be testimage:latest, but get %s", registrationoDeployment.Spec.Template.Spec.Containers[0].Image)
				}
				return nil
			}, eventuallyTimeout, eventuallyInterval).Should(gomega.BeNil())

			// Check if generations are correct
			gomega.Eventually(func() error {
				actual, err := operatorClient.OperatorV1().ClusterManagers().Get(context.Background(), clusterManagerName, metav1.GetOptions{})
				if err != nil {
					return err
				}

				registrationDeployment, err := kubeClient.AppsV1().Deployments(hubNamespace).Get(context.Background(), hubRegistrationDeployment, metav1.GetOptions{})
				if err != nil {
					return err
				}

				deploymentGeneration := helpers.NewGenerationStatus(appsv1.SchemeGroupVersion.WithResource("deployments"), registrationDeployment)
				actualGeneration := helpers.FindGenerationStatus(actual.Status.Generations, deploymentGeneration)
				if deploymentGeneration.LastGeneration != actualGeneration.LastGeneration {
					return fmt.Errorf("expected LastGeneration shoud be %d, but get %d", actualGeneration.LastGeneration, deploymentGeneration.LastGeneration)
				}
				return nil
			}, eventuallyTimeout, eventuallyInterval).Should(gomega.BeNil())
		})
	})

	ginkgo.Context("Cluster manager statuses", func() {
		ginkgo.It("should have correct degraded conditions", func() {
			gomega.Eventually(func() error {
				if _, err := kubeClient.AppsV1().Deployments(hubNamespace).Get(context.Background(), hubRegistrationDeployment, metav1.GetOptions{}); err != nil {
					return err
				}
				return nil
			}, eventuallyTimeout, eventuallyInterval).Should(gomega.BeNil())

			// The cluster manager should be unavailable at first
			util.AssertClusterManagerCondition(clusterManagerName, operatorClient, "HubRegistrationDegraded", "UnavailableRegistrationPod", metav1.ConditionTrue)

			// Update replica of deployment
			registrationDeployment, err := kubeClient.AppsV1().Deployments(hubNamespace).Get(context.Background(), hubRegistrationDeployment, metav1.GetOptions{})
			gomega.Expect(err).ToNot(gomega.HaveOccurred())
			registrationDeployment.Status.AvailableReplicas = 3
			registrationDeployment.Status.Replicas = 3
			registrationDeployment.Status.ReadyReplicas = 3
			_, err = kubeClient.AppsV1().Deployments(hubNamespace).UpdateStatus(context.Background(), registrationDeployment, metav1.UpdateOptions{})
			gomega.Expect(err).ToNot(gomega.HaveOccurred())

			// The cluster manager should be functional at last
			util.AssertClusterManagerCondition(clusterManagerName, operatorClient, "HubRegistrationDegraded", "RegistrationFunctional", metav1.ConditionFalse)
		})
	})

	ginkgo.Context("Serving cert rotation", func() {
		ginkgo.It("should rotate both serving cert and signing cert before they become expired", func() {
			secretNames := []string{"signer-secret", "registration-webhook-serving-cert", "work-webhook-serving-cert"}
			// wait until all secrets and configmap are in place
			gomega.Eventually(func() error {
				for _, name := range secretNames {
					if _, err := kubeClient.CoreV1().Secrets(hubNamespace).Get(context.Background(), name, metav1.GetOptions{}); err != nil {
						return err
					}
				}
				if _, err := kubeClient.CoreV1().ConfigMaps(hubNamespace).Get(context.Background(), "ca-bundle-configmap", metav1.GetOptions{}); err != nil {
					return err
				}
				return nil
			}, eventuallyTimeout, eventuallyInterval).Should(gomega.BeNil())

			// both serving cert and signing cert should aways be valid
			gomega.Consistently(func() error {
				configmap, err := kubeClient.CoreV1().ConfigMaps(hubNamespace).Get(context.Background(), "ca-bundle-configmap", metav1.GetOptions{})
				if err != nil {
					return err
				}
				for _, name := range []string{"signer-secret", "registration-webhook-serving-cert", "work-webhook-serving-cert"} {
					secret, err := kubeClient.CoreV1().Secrets(hubNamespace).Get(context.Background(), name, metav1.GetOptions{})
					if err != nil {
						return err
					}

					certificates, err := cert.ParseCertsPEM(secret.Data["tls.crt"])
					if err != nil {
						return err
					}
					if len(certificates) == 0 {
						return fmt.Errorf("certificates length equals to 0")
					}

					now := time.Now()
					certificate := certificates[0]
					if now.After(certificate.NotAfter) {
						return fmt.Errorf("certificate after NotAfter")
					}
					if now.Before(certificate.NotBefore) {
						return fmt.Errorf("certificate before NotBefore")
					}

					if name == "signer-secret" {
						continue
					}

					// ensure signing cert of serving certs in the ca bundle configmap
					caCerts, err := cert.ParseCertsPEM([]byte(configmap.Data["ca-bundle.crt"]))
					if err != nil {
						return err
					}

					found := false
					for _, caCert := range caCerts {
						if certificate.Issuer.CommonName != caCert.Subject.CommonName {
							continue
						}
						if now.After(caCert.NotAfter) {
							return fmt.Errorf("certificate after NotAfter")
						}
						if now.Before(caCert.NotBefore) {
							return fmt.Errorf("certificate before NotBefore")
						}
						found = true
						break
					}
					if !found {
						return fmt.Errorf("not found")
					}
				}
				return nil
			}, eventuallyTimeout*3, eventuallyInterval*3).Should(gomega.BeNil())
		})
	})
})

var _ = ginkgo.Describe("ClusterManager Detached Mode", func() {
	var cancel context.CancelFunc
	var hubRegistrationDeployment = fmt.Sprintf("%s-registration-controller", clusterManagerName)

	ginkgo.BeforeEach(func() {
		var ctx context.Context
		ctx, cancel = context.WithCancel(context.Background())
		go startHubOperator(ctx, v1.InstallModeDetached)
	})

	ginkgo.AfterEach(func() {
		if cancel != nil {
			cancel()
		}
	})

	ginkgo.Context("Deploy and clean hub component", func() {
		ginkgo.It("should have expected resource created successfully", func() {
			// Check namespace
			gomega.Eventually(func() error {
				if _, err := detachedKubeClient.CoreV1().Namespaces().Get(context.Background(), hubNamespaceDetached, metav1.GetOptions{}); err != nil {
					return err
				}
				return nil
			}, eventuallyTimeout, eventuallyInterval).Should(gomega.BeNil())

			// Check clusterrole/clusterrolebinding
			hubRegistrationClusterRole := fmt.Sprintf("open-cluster-management:%s-registration:controller", clusterManagerName)
			hubRegistrationWebhookClusterRole := fmt.Sprintf("open-cluster-management:%s-registration:webhook", clusterManagerName)
			hubWorkWebhookClusterRole := fmt.Sprintf("open-cluster-management:%s-registration:webhook", clusterManagerName)
			gomega.Eventually(func() error {
				if _, err := detachedKubeClient.RbacV1().ClusterRoles().Get(context.Background(), hubRegistrationClusterRole, metav1.GetOptions{}); err != nil {
					return err
				}
				return nil
			}, eventuallyTimeout, eventuallyInterval).Should(gomega.BeNil())
			gomega.Eventually(func() error {
				if _, err := detachedKubeClient.RbacV1().ClusterRoles().Get(context.Background(), hubRegistrationWebhookClusterRole, metav1.GetOptions{}); err != nil {
					return err
				}
				return nil
			}, eventuallyTimeout, eventuallyInterval).Should(gomega.BeNil())
			gomega.Eventually(func() error {
				if _, err := detachedKubeClient.RbacV1().ClusterRoles().Get(context.Background(), hubWorkWebhookClusterRole, metav1.GetOptions{}); err != nil {
					return err
				}
				return nil
			}, eventuallyTimeout, eventuallyInterval).Should(gomega.BeNil())
			gomega.Eventually(func() error {
				if _, err := detachedKubeClient.RbacV1().ClusterRoleBindings().Get(context.Background(), hubRegistrationClusterRole, metav1.GetOptions{}); err != nil {
					return err
				}
				return nil
			}, eventuallyTimeout, eventuallyInterval).Should(gomega.BeNil())
			gomega.Eventually(func() error {
				if _, err := detachedKubeClient.RbacV1().ClusterRoleBindings().Get(context.Background(), hubRegistrationWebhookClusterRole, metav1.GetOptions{}); err != nil {
					return err
				}
				return nil
			}, eventuallyTimeout, eventuallyInterval).Should(gomega.BeNil())
			gomega.Eventually(func() error {
				if _, err := detachedKubeClient.RbacV1().ClusterRoleBindings().Get(context.Background(), hubWorkWebhookClusterRole, metav1.GetOptions{}); err != nil {
					return err
				}
				return nil
			}, eventuallyTimeout, eventuallyInterval).Should(gomega.BeNil())

			// Check service account
			hubRegistrationSA := fmt.Sprintf("%s-registration-controller-sa", clusterManagerName)
			hubRegistrationWebhookSA := fmt.Sprintf("%s-registration-webhook-sa", clusterManagerName)
			hubWorkWebhookSA := fmt.Sprintf("%s-work-webhook-sa", clusterManagerName)
			gomega.Eventually(func() error {
				if _, err := detachedKubeClient.CoreV1().ServiceAccounts(hubNamespaceDetached).Get(context.Background(), hubRegistrationSA, metav1.GetOptions{}); err != nil {
					return err
				}
				return nil
			}, eventuallyTimeout, eventuallyInterval).Should(gomega.BeNil())
			gomega.Eventually(func() error {
				if _, err := detachedKubeClient.CoreV1().ServiceAccounts(hubNamespaceDetached).Get(context.Background(), hubRegistrationWebhookSA, metav1.GetOptions{}); err != nil {
					return err
				}
				return nil
			}, eventuallyTimeout, eventuallyInterval).Should(gomega.BeNil())
			gomega.Eventually(func() error {
				if _, err := detachedKubeClient.CoreV1().ServiceAccounts(hubNamespaceDetached).Get(context.Background(), hubWorkWebhookSA, metav1.GetOptions{}); err != nil {
					return err
				}
				return nil
			}, eventuallyTimeout, eventuallyInterval).Should(gomega.BeNil())

			// Check deployment
			gomega.Eventually(func() error {
				if _, err := detachedKubeClient.AppsV1().Deployments(hubNamespaceDetached).Get(context.Background(), hubRegistrationDeployment, metav1.GetOptions{}); err != nil {
					return err
				}
				return nil
			}, eventuallyTimeout, eventuallyInterval).Should(gomega.BeNil())

			hubRegistrationWebhookDeployment := fmt.Sprintf("%s-registration-webhook", clusterManagerName)
			gomega.Eventually(func() error {
				if _, err := detachedKubeClient.AppsV1().Deployments(hubNamespaceDetached).Get(context.Background(), hubRegistrationWebhookDeployment, metav1.GetOptions{}); err != nil {
					return err
				}
				return nil
			}, eventuallyTimeout, eventuallyInterval).Should(gomega.BeNil())

			hubWorkWebhookDeployment := fmt.Sprintf("%s-work-webhook", clusterManagerName)
			gomega.Eventually(func() error {
				if _, err := detachedKubeClient.AppsV1().Deployments(hubNamespaceDetached).Get(context.Background(), hubWorkWebhookDeployment, metav1.GetOptions{}); err != nil {
					return err
				}
				return nil
			}, eventuallyTimeout, eventuallyInterval).Should(gomega.BeNil())

			// Check service
			gomega.Eventually(func() error {
				if _, err := detachedKubeClient.CoreV1().Services(hubNamespaceDetached).Get(context.Background(), "cluster-manager-registration-webhook", metav1.GetOptions{}); err != nil {
					return err
				}
				return nil
			}, eventuallyTimeout, eventuallyInterval).Should(gomega.BeNil())

			gomega.Eventually(func() error {
				if _, err := detachedKubeClient.CoreV1().Services(hubNamespaceDetached).Get(context.Background(), "cluster-manager-work-webhook", metav1.GetOptions{}); err != nil {
					return err
				}
				return nil
			}, eventuallyTimeout, eventuallyInterval).Should(gomega.BeNil())

			// Check webhook secret
			registrationWebhookSecret := "registration-webhook-serving-cert"
			gomega.Eventually(func() error {
				s, err := detachedKubeClient.CoreV1().Secrets(hubNamespaceDetached).Get(context.Background(), registrationWebhookSecret, metav1.GetOptions{})
				if err != nil {
					return err
				}
				if s.Data == nil {
					return fmt.Errorf("s.Data is nil")
				} else if s.Data["tls.crt"] == nil {
					return fmt.Errorf("s.Data doesn't contain key 'tls.crt'")
				} else if s.Data["tls.key"] == nil {
					return fmt.Errorf("s.Data doesn't contain key 'tls.key'")
				}
				return nil
			}, eventuallyTimeout, eventuallyInterval).Should(gomega.BeNil())

			workWebhookSecret := "work-webhook-serving-cert"
			gomega.Eventually(func() error {
				s, err := detachedKubeClient.CoreV1().Secrets(hubNamespaceDetached).Get(context.Background(), workWebhookSecret, metav1.GetOptions{})
				if err != nil {
					return err
				}
				if s.Data == nil {
					return fmt.Errorf("s.Data is nil")
				} else if s.Data["tls.crt"] == nil {
					return fmt.Errorf("s.Data doesn't contain key 'tls.crt'")
				} else if s.Data["tls.key"] == nil {
					return fmt.Errorf("s.Data doesn't contain key 'tls.key'")
				}
				return nil
			}, eventuallyTimeout, eventuallyInterval).Should(gomega.BeNil())

			// Check validating webhook
			registrationValidtingWebhook := "managedclustervalidators.admission.cluster.open-cluster-management.io"
			gomega.Eventually(func() error {
				if _, err := detachedKubeClient.AdmissionregistrationV1().ValidatingWebhookConfigurations().Get(context.Background(), registrationValidtingWebhook, metav1.GetOptions{}); err != nil {
					return err
				}
				return nil
			}, eventuallyTimeout, eventuallyInterval).Should(gomega.BeNil())

			workValidtingWebhook := "manifestworkvalidators.admission.work.open-cluster-management.io"
			gomega.Eventually(func() error {
				if _, err := detachedKubeClient.AdmissionregistrationV1().ValidatingWebhookConfigurations().Get(context.Background(), workValidtingWebhook, metav1.GetOptions{}); err != nil {
					return err
				}
				return nil
			}, eventuallyTimeout, eventuallyInterval).Should(gomega.BeNil())

			util.AssertClusterManagerCondition(clusterManagerName, detachedOperatorClient, "Applied", "ClusterManagerApplied", metav1.ConditionTrue)
		})

		ginkgo.It("Deployment should be updated when clustermanager is changed", func() {
			gomega.Eventually(func() error {
				if _, err := detachedKubeClient.AppsV1().Deployments(hubNamespaceDetached).Get(context.Background(), hubRegistrationDeployment, metav1.GetOptions{}); err != nil {
					return err
				}
				return nil
			}, eventuallyTimeout, eventuallyInterval).Should(gomega.BeNil())

			// Check if generations are correct
			gomega.Eventually(func() error {
				actual, err := detachedOperatorClient.OperatorV1().ClusterManagers().Get(context.Background(), clusterManagerName, metav1.GetOptions{})
				if err != nil {
					return err
				}

				if actual.Generation != actual.Status.ObservedGeneration {
					return fmt.Errorf("except generation to be %d, but got %d", actual.Status.ObservedGeneration, actual.Generation)
				}

				return nil
			}, eventuallyTimeout, eventuallyInterval).Should(gomega.BeNil())

			clusterManager, err := detachedOperatorClient.OperatorV1().ClusterManagers().Get(context.Background(), clusterManagerName, metav1.GetOptions{})
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			clusterManager.Spec.RegistrationImagePullSpec = "testimage:latest"
			_, err = detachedOperatorClient.OperatorV1().ClusterManagers().Update(context.Background(), clusterManager, metav1.UpdateOptions{})
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			gomega.Eventually(func() error {
				actual, err := detachedKubeClient.AppsV1().Deployments(hubNamespaceDetached).Get(context.Background(), hubRegistrationDeployment, metav1.GetOptions{})
				if err != nil {
					return err
				}
				gomega.Expect(len(actual.Spec.Template.Spec.Containers)).Should(gomega.Equal(1))
				if actual.Spec.Template.Spec.Containers[0].Image != "testimage:latest" {
					return fmt.Errorf("expected image to be testimage:latest but get %s", actual.Spec.Template.Spec.Containers[0].Image)
				}
				return nil
			}, eventuallyTimeout, eventuallyInterval).Should(gomega.BeNil())

			// Check if generations are correct
			gomega.Eventually(func() error {
				actual, err := detachedOperatorClient.OperatorV1().ClusterManagers().Get(context.Background(), clusterManagerName, metav1.GetOptions{})
				if err != nil {
					return err
				}

				if actual.Generation != actual.Status.ObservedGeneration {
					return fmt.Errorf("except generation to be %d, but got %d", actual.Status.ObservedGeneration, actual.Generation)
				}

				return nil
			}, eventuallyTimeout, eventuallyInterval).Should(gomega.BeNil())

			// Check if relatedResources are correct
			gomega.Eventually(func() error {
				actual, err := detachedOperatorClient.OperatorV1().ClusterManagers().Get(context.Background(), clusterManagerName, metav1.GetOptions{})
				if err != nil {
					return err
				}
				if len(actual.Status.RelatedResources) != 33 {
					return fmt.Errorf("should get 33 relatedResources, actual got %v", len(actual.Status.RelatedResources))
				}
				return nil
			}, eventuallyTimeout, eventuallyInterval).ShouldNot(gomega.HaveOccurred())

		})

		ginkgo.It("Deployment should be added nodeSelector and toleration when add nodePlacement into clustermanager", func() {
			clusterManager, err := detachedOperatorClient.OperatorV1().ClusterManagers().Get(context.Background(), clusterManagerName, metav1.GetOptions{})
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			clusterManager.Spec.NodePlacement = v1.NodePlacement{
				NodeSelector: map[string]string{"node-role.kubernetes.io/infra": ""},
				Tolerations: []corev1.Toleration{
					{
						Key:      "node-role.kubernetes.io/infra",
						Operator: corev1.TolerationOpExists,
						Effect:   corev1.TaintEffectNoSchedule,
					},
				},
			}
			_, err = detachedOperatorClient.OperatorV1().ClusterManagers().Update(context.Background(), clusterManager, metav1.UpdateOptions{})
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			gomega.Eventually(func() error {
				actual, err := detachedKubeClient.AppsV1().Deployments(hubNamespaceDetached).Get(context.Background(), hubRegistrationDeployment, metav1.GetOptions{})
				if err != nil {
					return err
				}
				gomega.Expect(len(actual.Spec.Template.Spec.Containers)).Should(gomega.Equal(1))
				if len(actual.Spec.Template.Spec.NodeSelector) == 0 {
					return fmt.Errorf("length of node selector should not equals to 0")
				}
				if _, ok := actual.Spec.Template.Spec.NodeSelector["node-role.kubernetes.io/infra"]; !ok {
					return fmt.Errorf("node-role.kubernetes.io/infra not exist")
				}
				if len(actual.Spec.Template.Spec.Tolerations) == 0 {
					return fmt.Errorf("length of node selecor should not equals to 0")
				}
				for _, toleration := range actual.Spec.Template.Spec.Tolerations {
					if toleration.Key == "node-role.kubernetes.io/infra" {
						return nil
					}
				}

				return fmt.Errorf("no key equals to node-role.kubernetes.io/infra")
			}, eventuallyTimeout, eventuallyInterval).Should(gomega.BeNil())
		})
		ginkgo.It("Deployment should be reconciled when manually updated", func() {
			gomega.Eventually(func() error {
				if _, err := detachedKubeClient.AppsV1().Deployments(hubNamespaceDetached).Get(context.Background(), hubRegistrationDeployment, metav1.GetOptions{}); err != nil {
					return err
				}
				return nil
			}, eventuallyTimeout, eventuallyInterval).Should(gomega.BeNil())
			registrationoDeployment, err := detachedKubeClient.AppsV1().Deployments(hubNamespaceDetached).Get(context.Background(), hubRegistrationDeployment, metav1.GetOptions{})
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			registrationoDeployment.Spec.Template.Spec.Containers[0].Image = "testimage2:latest"
			_, err = detachedKubeClient.AppsV1().Deployments(hubNamespaceDetached).Update(context.Background(), registrationoDeployment, metav1.UpdateOptions{})
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Eventually(func() error {
				registrationoDeployment, err := detachedKubeClient.AppsV1().Deployments(hubNamespaceDetached).Get(context.Background(), hubRegistrationDeployment, metav1.GetOptions{})
				if err != nil {
					return err
				}
				if registrationoDeployment.Spec.Template.Spec.Containers[0].Image != "testimage:latest" {
					return fmt.Errorf("image should be testimage:latest, but get %s", registrationoDeployment.Spec.Template.Spec.Containers[0].Image)
				}
				return nil
			}, eventuallyTimeout, eventuallyInterval).Should(gomega.BeNil())

			// Check if generations are correct
			gomega.Eventually(func() error {
				actual, err := detachedOperatorClient.OperatorV1().ClusterManagers().Get(context.Background(), clusterManagerName, metav1.GetOptions{})
				if err != nil {
					return err
				}

				registrationDeployment, err := detachedKubeClient.AppsV1().Deployments(hubNamespaceDetached).Get(context.Background(), hubRegistrationDeployment, metav1.GetOptions{})
				if err != nil {
					return err
				}

				deploymentGeneration := helpers.NewGenerationStatus(appsv1.SchemeGroupVersion.WithResource("deployments"), registrationDeployment)
				actualGeneration := helpers.FindGenerationStatus(actual.Status.Generations, deploymentGeneration)
				if deploymentGeneration.LastGeneration != actualGeneration.LastGeneration {
					return fmt.Errorf("expected LastGeneration shoud be %d, but get %d", actualGeneration.LastGeneration, deploymentGeneration.LastGeneration)
				}
				return nil
			}, eventuallyTimeout, eventuallyInterval).Should(gomega.BeNil())
		})
	})

	ginkgo.Context("Cluster manager statuses", func() {
		ginkgo.It("should have correct degraded conditions", func() {
			gomega.Eventually(func() error {
				if _, err := detachedKubeClient.AppsV1().Deployments(hubNamespaceDetached).Get(context.Background(), hubRegistrationDeployment, metav1.GetOptions{}); err != nil {
					return err
				}
				return nil
			}, eventuallyTimeout, eventuallyInterval).Should(gomega.BeNil())

			// The cluster manager should be unavailable at first
			util.AssertClusterManagerCondition(clusterManagerName, detachedOperatorClient, "HubRegistrationDegraded", "UnavailableRegistrationPod", metav1.ConditionTrue)

			// Update replica of deployment
			registrationDeployment, err := detachedKubeClient.AppsV1().Deployments(hubNamespaceDetached).Get(context.Background(), hubRegistrationDeployment, metav1.GetOptions{})
			gomega.Expect(err).ToNot(gomega.HaveOccurred())
			registrationDeployment.Status.AvailableReplicas = 3
			registrationDeployment.Status.Replicas = 3
			registrationDeployment.Status.ReadyReplicas = 3
			_, err = detachedKubeClient.AppsV1().Deployments(hubNamespaceDetached).UpdateStatus(context.Background(), registrationDeployment, metav1.UpdateOptions{})
			gomega.Expect(err).ToNot(gomega.HaveOccurred())

			// The cluster manager should be functional at last
			util.AssertClusterManagerCondition(clusterManagerName, detachedOperatorClient, "HubRegistrationDegraded", "RegistrationFunctional", metav1.ConditionFalse)
		})
	})

	ginkgo.Context("Serving cert rotation", func() {
		ginkgo.It("should rotate both serving cert and signing cert before they become expired", func() {
			secretNames := []string{"signer-secret", "registration-webhook-serving-cert", "work-webhook-serving-cert"}
			// wait until all secrets and configmap are in place
			gomega.Eventually(func() error {
				for _, name := range secretNames {
					if _, err := detachedKubeClient.CoreV1().Secrets(hubNamespaceDetached).Get(context.Background(), name, metav1.GetOptions{}); err != nil {
						return err
					}
				}
				if _, err := detachedKubeClient.CoreV1().ConfigMaps(hubNamespaceDetached).Get(context.Background(), "ca-bundle-configmap", metav1.GetOptions{}); err != nil {
					return err
				}
				return nil
			}, eventuallyTimeout, eventuallyInterval).Should(gomega.BeNil())

			// both serving cert and signing cert should aways be valid
			gomega.Consistently(func() error {
				configmap, err := detachedKubeClient.CoreV1().ConfigMaps(hubNamespaceDetached).Get(context.Background(), "ca-bundle-configmap", metav1.GetOptions{})
				if err != nil {
					return err
				}
				for _, name := range []string{"signer-secret", "registration-webhook-serving-cert", "work-webhook-serving-cert"} {
					secret, err := detachedKubeClient.CoreV1().Secrets(hubNamespaceDetached).Get(context.Background(), name, metav1.GetOptions{})
					if err != nil {
						return err
					}

					certificates, err := cert.ParseCertsPEM(secret.Data["tls.crt"])
					if err != nil {
						return err
					}
					if len(certificates) == 0 {
						return fmt.Errorf("certificates length equals to 0")
					}

					now := time.Now()
					certificate := certificates[0]
					if now.After(certificate.NotAfter) {
						return fmt.Errorf("certificate after NotAfter")
					}
					if now.Before(certificate.NotBefore) {
						return fmt.Errorf("certificate before NotBefore")
					}

					if name == "signer-secret" {
						continue
					}

					// ensure signing cert of serving certs in the ca bundle configmap
					caCerts, err := cert.ParseCertsPEM([]byte(configmap.Data["ca-bundle.crt"]))
					if err != nil {
						return err
					}

					found := false
					for _, caCert := range caCerts {
						if certificate.Issuer.CommonName != caCert.Subject.CommonName {
							continue
						}
						if now.After(caCert.NotAfter) {
							return fmt.Errorf("certificate after NotAfter")
						}
						if now.Before(caCert.NotBefore) {
							return fmt.Errorf("certificate before NotBefore")
						}
						found = true
						break
					}
					if !found {
						return fmt.Errorf("not found")
					}
				}
				return nil
			}, eventuallyTimeout*3, eventuallyInterval*3).Should(gomega.BeNil())
		})
	})
})

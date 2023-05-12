/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	seleniumv1 "quay.io/molnar_liviusz/selenium-test-operator/api/v1"
	//+kubebuilder:scaffold:imports
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var cfg *rest.Config
var k8sClient client.Client
var testEnv *envtest.Environment

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t, "Controller Suite")
}

var _ = BeforeSuite(func() {
	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))

	By("bootstrapping test environment")
	testEnv = &envtest.Environment{
		CRDDirectoryPaths:     []string{filepath.Join("..", "config", "crd", "bases")},
		ErrorIfCRDPathMissing: true,
	}

	var err error
	// cfg is defined in this file globally.
	cfg, err = testEnv.Start()
	Expect(err).NotTo(HaveOccurred())
	Expect(cfg).NotTo(BeNil())

	err = seleniumv1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	//+kubebuilder:scaffold:scheme

	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
	Expect(err).NotTo(HaveOccurred())
	Expect(k8sClient).NotTo(BeNil())
})

var _ = AfterSuite(func() {
	By("tearing down the test environment")
	err := testEnv.Stop()
	Expect(err).NotTo(HaveOccurred())
})

const namespace = "operator-system"

var _ = Describe("e2e testing of automating a test", Ordered, func() {
	BeforeAll(func() {
		// Moon is installed in this test
		By("Creating moon namespace")
		cmd := exec.Command("kubectl", "create", "ns", "moon")
		_, _ = Run(cmd)

		By("Installing Moon")
		Expect(InstallMoon()).To(Succeed())
	})

	Context("SeleniumTest Operator", func() {
		It("should run successfully", func() {
			var controllerPodName string
			var err error
			projectDir, _ := GetProjectDir()

			// operatorImage stores the name of the image used in the example
			var operatorImage = "quay.io/molnar_liviusz/selenium-test-operator:v0.0.24"

			By("deploying the SeleniumTest operator' controller-manager")
			var cmd = exec.Command("make", "deploy", fmt.Sprintf("IMG=%s", operatorImage))
			_, err = Run(cmd)
			ExpectWithOffset(1, err).NotTo(HaveOccurred())

			By("validating that the controller-manager pod is running as expected")
			verifyControllerUp := func() error {
				// Get pod name
				cmd = exec.Command("kubectl", "get",
					"pods", "-l", "control-plane=controller-manager",
					"-o", "go-template={{ range .items }}{{ if not .metadata.deletionTimestamp }}{{ .metadata.name }}"+
						"{{ \"\\n\" }}{{ end }}{{ end }}",
					"-n", namespace,
				)
				podOutput, err := Run(cmd)
				ExpectWithOffset(2, err).NotTo(HaveOccurred())
				podNames := GetNonEmptyLines(string(podOutput))
				if len(podNames) != 1 {
					return fmt.Errorf("expect 1 controller pods running, but got %d", len(podNames))
				}
				controllerPodName = podNames[0]
				ExpectWithOffset(2, controllerPodName).Should(ContainSubstring("controller-manager"))

				// Validate pod status
				cmd = exec.Command("kubectl", "get",
					"pods", controllerPodName, "-o", "jsonpath={.status.phase}",
					"-n", namespace,
				)
				status, err := Run(cmd)
				ExpectWithOffset(2, err).NotTo(HaveOccurred())
				if string(status) != "Running" {
					return fmt.Errorf("controller pod in %s status", status)
				}
				return nil
			}
			EventuallyWithOffset(1, verifyControllerUp, time.Minute, time.Second).Should(Succeed())

			By("creating an configmap with a selenium .side test")
			EventuallyWithOffset(1, func() error {
				cmd = exec.Command("kubectl", "apply", "-f", filepath.Join(projectDir,
					"config/samples/sample-configmap.yaml"), "-n", namespace)
				_, err = Run(cmd)
				return err
			}, time.Minute, time.Second).Should(Succeed())

			By("creating an instance of the SeleniumTest(CR)")
			EventuallyWithOffset(1, func() error {
				cmd = exec.Command("kubectl", "apply", "-f", filepath.Join(projectDir,
					"config/samples/selenium_v1_seleniumtest.yaml"), "-n", namespace)
				_, err = Run(cmd)
				return err
			}, time.Minute, time.Second).Should(Succeed())

			By("validating that the SeleniumTestResult custom resource is created or updated")
			EventuallyWithOffset(1, func() error {
				cmd = exec.Command("kubectl", "get", "seleniumtestresult",
					"seleniumtest-sample", "-o", "jsonpath={.metadata.creationTimestamp}",
					"-n", namespace,
				)
				time, err := Run(cmd)
				fmt.Println(string(time))
				ExpectWithOffset(2, err).NotTo(HaveOccurred())
				if !strings.Contains(string(time), ":") {
					return fmt.Errorf("SeleniumTestResult was not created")
				}
				return nil
			}, 5*time.Minute, time.Second).Should(Succeed())
		})
	})

	Context("When creating SeleniumTest", func() {
		It("Should should look up configMap and error for not found", func() {
			projectDir, _ := GetProjectDir()
			var err error

			By("Creating a new SeleniumTest")
			EventuallyWithOffset(1, func() error {
				var cmd = exec.Command("kubectl", "apply", "-f", filepath.Join(projectDir,
					"config/testing/seleniumtest_testfile.yaml"), "-n", "default")
				_, err = Run(cmd)
				return err
			}, time.Minute, time.Second).Should(Succeed())

			By("By checking the SeleniumTest did not created ServiceAccount")
			var cmd = exec.Command("kubectl", "get",
				"serviceaccounts", "test-seleniumtest", "-o", "json",
				"-n", "default",
			)
			_, err = Run(cmd)
			ExpectWithOffset(2, err).To(HaveOccurred())
		})
	})

	AfterAll(func() {
		By("Uninstalling Moon")
		UninstallMoon()

		By("Removing moon namespace")
		cmd := exec.Command("kubectl", "delete", "ns", "moon")
		_, _ = Run(cmd)
	})
})

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
	"context"

	"github.com/prometheus/client_golang/prometheus"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"

	seleniumv1 "quay.io/molnar_liviusz/selenium-test-operator/api/v1"
)

// SeleniumTestResultReconciler reconciles a SeleniumTestResult object
type SeleniumTestResultReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

var log2 = ctrllog.Log.WithName("controller_seleniumtestresult")

//+kubebuilder:rbac:groups=selenium.mliviusz.com,resources=seleniumtestresults,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=selenium.mliviusz.com,resources=seleniumtestresults/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=selenium.mliviusz.com,resources=seleniumtestresults/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the SeleniumTestResult object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
func (r *SeleniumTestResultReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log2 := ctrllog.FromContext(ctx)

	log2.Info("Reconciling SeleniumTestResult", "Request.Namespace", req.Namespace, "Request.Name", req.Name)

	instance := &seleniumv1.SeleniumTestResult{}
	err := r.Client.Get(context.Background(), req.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			log2.Info("SeleniumTestResult deleted")
			return ctrl.Result{}, nil
		}
		log2.Error(err, "Failed to get SeleniumTestResult")
		return ctrl.Result{}, err
	}

	labels := prometheus.Labels{"test_name": instance.Name, "namespace": instance.Namespace}
	log2.Info("Updating metric selenium_test_results with labels", "test_name", instance.Name, "namespace", instance.Namespace)
	if instance.Spec.Success {
		test_results.With(labels).Set(1)
	} else {
		test_results.With(labels).Set(0)
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *SeleniumTestResultReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&seleniumv1.SeleniumTestResult{}).
		Complete(r)
}

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
	/*
		"context"

		"k8s.io/apimachinery/pkg/runtime"
		"k8s.io/apimachinery/pkg/api/errors"
		"k8s.io/apimachinery/pkg/types"
		"k8s.io/apimachinery/pkg/util/intstr"
		"k8s.io/apimachinery/pkg/util/runtime"
		"k8s.io/apimachinery/pkg/util/wait"
		"k8s.io/client-go/kubernetes"
		"k8s.io/client-go/rest"
		"k8s.io/client-go/tools/clientcmd"
		"k8s.io/client-go/tools/record"
	*/
	"github.com/go-logr/logr"
	/*	"github.com/operator-framework/operator-sdk/pkg/k8sutil"
		"github.com/operator-framework/operator-sdk/pkg/predicate"
		"github.com/operator-framework/operator-sdk/pkg/sdk"
		"github.com/operator-framework/operator-sdk/pkg/util"
		"github.com/operator-framework/operator-sdk/pkg/util/retry"
		"github.com/operator-framework/operator-sdk/pkg/util/validation"
		"github.com/prometheus/common/log"
		"github.com/rs/xid"

		ctrl "sigs.k8s.io/controller-runtime"

		corev1 "k8s.io/api/core/v1"
		batchv1beta1 "k8s.io/api/batch/v1beta1"
		"k8s.io/apimachinery/pkg/api/resource"
		metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
		"sigs.k8s.io/controller-runtime/pkg/client"
		"sigs.k8s.io/controller-runtime/pkg/controller"
		"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
		"sigs.k8s.io/controller-runtime/pkg/event"
		"sigs.k8s.io/controller-runtime/pkg/handler"
		"sigs.k8s.io/controller-runtime/pkg/manager"
		"sigs.k8s.io/controller-runtime/pkg/predicate"
		"sigs.k8s.io/controller-runtime/pkg/reconcile"
		"sigs.k8s.io/controller-runtime/pkg/source"

		seleniumv1 "github.com/Mliviusz/selenium-test-operator/api/v1"
	*/
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"

	seleniumv1 "quay.io/molnar_liviusz/selenium-test-operator/api/v1"

	"github.com/prometheus/client_golang/prometheus"
    "sigs.k8s.io/controller-runtime/pkg/metrics"
)

var log = ctrllog.Log.WithName("controller_seleniumtest")

// SeleniumTestReconciler reconciles a SeleniumTest object
type SeleniumTestReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

const (
	finalizerName     = "seleniumtest.selenium/finalizer"
	configMapKey      = "config"
	cronJobAPIGroup   = "batch"
	cronJobAPIVersion = "v1beta1"
)

//+kubebuilder:rbac:groups=selenium.selenium,resources=seleniumtests,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=selenium.selenium,resources=seleniumtests/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=selenium.selenium,resources=seleniumtests/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the SeleniumTest object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
func (r *SeleniumTestReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrllog.FromContext(ctx)

	log.Info("Reconciling SeleniumTest", "Request.Namespace", req.Namespace, "Request.Name", req.Name)

	// Add metrics information
	initmetrics()
	// TODO
	goobers.Inc()

	instance := &seleniumv1.SeleniumTest{}
	err := r.client.Get(context.Background(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			r.log.Info("SeleniumTest deleted")
			return reconcile.Result{}, nil
		}
		r.log.Error(err, "Failed to get SeleniumTest")
		return reconcile.Result{}, err
	}

	// Add finalizer for this CR
	if instance.ObjectMeta.DeletionTimestamp.IsZero() {
		if !util.Contains(instance.ObjectMeta.Finalizers, finalizerName) {
			instance.ObjectMeta.Finalizers = append(instance.ObjectMeta.Finalizers, finalizerName)
			if err := r.client.Update(context.Background(), instance); err != nil {
				r.log.Error(err, "Failed to update SeleniumTest with finalizer")
				return reconcile.Result{}, err
			}
		}
	} else {
		// Handle deletion reconciliation loop
		if util.Contains(instance.ObjectMeta.Finalizers, finalizerName) {
			// Run finalization logic for finalizerName. If the
			// finalization logic fails, don't remove the finalizer so
			// that we can retry during the next reconciliation.
			if err := r.finalizeSeleniumTest(instance); err != nil {
				r.log.Error(err, "Failed to finalize SeleniumTest")
				return reconcile.Result{}, err
			}

			// Remove finalizer. Once all finalizers have been
			// removed, the object will be deleted.
			instance.ObjectMeta.Finalizers = util.Filter(instance.ObjectMeta.Finalizers, finalizerName)
			if err := r.client.Update(context.Background(), instance); err != nil {
				r.log.Error(err, "Failed to update SeleniumTest with finalizer")
				return reconcile.Result{}, err
			}
		}

		// Stop reconciliation as the item is being deleted
		return reconcile.Result{}, nil
	}

	// Ensure the ConfigMap is present
	//err = r.ensureConfigMap(instance)
	//if err != nil {
	//	r.log.Error(err, "Failed to ensure ConfigMap is present")
	//	return reconcile.Result{}, err
	//}

	// Ensure the CronJob is present
	err = r.ensureCronJob(instance)
	if err != nil {
		r.log.Error(err, "Failed to ensure CronJob is present")
		return reconcile.Result{}, err
	}

	// Update status
	// TODO add job states to status
	err = retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		instance.Status.LastScheduleTime = instance.Spec.Schedule
		return r.client.Status().Update(context.Background(), instance)
	})
	if err != nil {
		r.log.Error(err, "Failed to update SeleniumTest status")
		return reconcile.Result{}, err
	}

	r.log.Info("Reconcile loop completed successfully")
	return reconcile.Result{}, nil
}

func (r *SeleniumTestReconciler) ensureCronJob(instance *seleniumv1.SeleniumTest) error {
	cronJob := &batchv1beta1.CronJob{}
	err := r.client.Get(context.Background(), types.NamespacedName{Namespace: instance.Namespace, Name: instance.Name}, cronJob)
	if err != nil && errors.IsNotFound(err) {
		// Create the CronJob
		newCronJob := r.newCronJobForSeleniumTest(instance)
		r.log.Info("Creating a new CronJob", "CronJob.Namespace", newCronJob.Namespace, "CronJob.Name", newCronJob.Name)
		err = r.client.Create(context.Background(), newCronJob)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	return nil
}

func (r *SeleniumTestReconciler) newCronJobForSeleniumTest(instance *seleniumv1.SeleniumTest) *batchv1beta1.CronJob {
	labels := map[string]string{
		"app": "selenium-test",
	}
	trueVar := true

	return &batchv1beta1.CronJob{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.Name,
			Namespace: instance.Namespace,
			Labels:    labels,
		},
		Spec: batchv1beta1.CronJobSpec{
			Schedule:          instance.Spec.Schedule,
			JobTemplate:       batchv1beta1.JobTemplateSpec{Spec: batchv1.JobSpec{Template: corev1.PodTemplateSpec{}}},
			ConcurrencyPolicy: batchv1beta1.ForbidConcurrent,
			JobBackoffLimit:   instance.Spec.JobBackoffLimit,
		},
	}

	// Configure the container template
	// TODO
	container := corev1.Container{
		Name:            "selenium-test",
		Image:           instance.Spec.Image,
		ImagePullPolicy: corev1.PullIfNotPresent,
		Command:         []string{"/bin/sh", "-c", "echo Hello from the SeleniumTest CronJob"},
	}

	// Create a volume and volume mount for the ConfigMap
	// TODO
	volumeName := "config-volume"
	volume := corev1.Volume{
		Name: volumeName,
		VolumeSource: corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: instance.Name,
				},
			},
		},
	}

	volumeMount := corev1.VolumeMount{
		Name:      volumeName,
		MountPath: "/mnt/config",
	}

	// Add the container to the pod template
	cronJob.Spec.JobTemplate.Spec.Template.Spec.Containers = []corev1.Container{container}
	cronJob.Spec.JobTemplate.Spec.Template.Spec.Volumes = []corev1.Volume{volume}
	cronJob.Spec.JobTemplate.Spec.Template.Spec.Containers[0].VolumeMounts = []corev1.VolumeMount{volumeMount}

	// Configure the JobBackoffLimit to 0 so that failed Jobs are not retried
	cronJob.Spec.JobBackoffLimit = 0

	// Set the owner reference so that the CronJob gets deleted when the SeleniumTest is deleted
	controllerutil.SetControllerReference(instance, cronJob, r.scheme)

	return cronJob
}

// SetupWithManager sets up the controller with the Manager.
func (r *SeleniumTestReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.scheme = mgr.GetScheme()
	r.client = mgr.GetClient()
	r.log = logf.Log.WithName("controller_seleniumtest")

	if err := r.watchConfigMap(); err != nil {
		return err
	}

	if err := r.watchCronJob(); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&seleniumv1.SeleniumTest{}).
		Complete(r)
}

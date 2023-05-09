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

	//"github.com/go-logr/logr"

	"k8s.io/apimachinery/pkg/api/errors"
	//"k8s.io/apimachinery/pkg/util/intstr"
	//"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/types"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	//"github.com/operator-framework/operator-sdk/pkg/util"

	batchv1 "k8s.io/api/batch/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	//batchv1beta1 "k8s.io/api/batch/v1beta1"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"

	//"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	seleniumv1 "quay.io/molnar_liviusz/selenium-test-operator/api/v1"
	//"github.com/prometheus/client_golang/prometheus"
	//"sigs.k8s.io/controller-runtime/pkg/metrics"
)

// SeleniumTestReconciler reconciles a SeleniumTest object
type SeleniumTestReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

var log = ctrllog.Log.WithName("controller_seleniumtest")

//+kubebuilder:rbac:groups=selenium.mliviusz.com,resources=seleniumtests,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=selenium.mliviusz.com,resources=seleniumtests/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=selenium.mliviusz.com,resources=seleniumtests/finalizers,verbs=update
//+kubebuilder:rbac:groups=*,resources=cronjobs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=*,resources=configmaps,verbs=get;list;watch;create;update;patch
//+kubebuilder:rbac:groups=*,resources=serviceaccounts,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=*,resources=rolebindings,verbs=get;list;watch;create;update;patch;delete

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

	instance := &seleniumv1.SeleniumTest{}
	err := r.Client.Get(context.Background(), req.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("SeleniumTest deleted")
			return ctrl.Result{}, nil
		}
		log.Error(err, "Failed to get SeleniumTest")
		return ctrl.Result{}, err
	}

	// Ensure the ConfigMap is present
	err = r.ensureConfigMap(instance)
	if err != nil {
		log.Error(err, "Failed to ensure ConfigMap is present")
		return ctrl.Result{}, err
	}

	// Ensure the ServiceAccount is present, if not, creates it
	err = r.ensureServiceAccount(instance)
	if err != nil {
		log.Error(err, "Failed to ensure ServiceAccount for CronJob is present")
		return ctrl.Result{}, err
	}

	// Ensure the RoleBinding is present, if not, creates it
	err = r.ensureRoleBinding(instance)
	if err != nil {
		log.Error(err, "Failed to ensure RoleBinding for ServiceAccount is present")
		return ctrl.Result{}, err
	}

	// Ensure the CronJob is present, if not, creates it
	err = r.ensureCronJob(instance)
	if err != nil {
		log.Error(err, "Failed to ensure CronJob is present")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *SeleniumTestReconciler) ensureConfigMap(instance *seleniumv1.SeleniumTest) error {
	configMap := &corev1.ConfigMap{}
	err := r.Client.Get(context.Background(), types.NamespacedName{Namespace: instance.Namespace, Name: instance.Spec.ConfigMapName}, configMap)
	if err != nil && errors.IsNotFound(err) {
		log.Info("ConfigMap with the name ", instance.Spec.ConfigMapName, " in namespace ", instance.Namespace, " was not found")
		return err
	} else if err != nil {
		return err
	}

	return nil
}

func (r *SeleniumTestReconciler) ensureServiceAccount(instance *seleniumv1.SeleniumTest) error {
	serviceAccount := &corev1.ServiceAccount{}
	err := r.Client.Get(context.Background(), types.NamespacedName{Namespace: instance.Namespace, Name: instance.Name}, serviceAccount)
	if err != nil && errors.IsNotFound(err) {
		// Create the ServiceAccount
		newServiceAccount := r.newServiceAccountForSeleniumTest(instance)
		log.Info("Creating a new ServiceAccount", "ServiceAccount.Namespace", newServiceAccount.Namespace, "ServiceAccount.Name", newServiceAccount.Name)
		err = r.Client.Create(context.Background(), newServiceAccount)
		if err != nil {
			return err
		} else {
			log.Info("ServiceAccount ", newServiceAccount.Name, " created")
		}
	} else if err != nil {
		return err
	}

	return nil
}

func (r *SeleniumTestReconciler) ensureRoleBinding(instance *seleniumv1.SeleniumTest) error {
	roleBinding := &rbacv1.RoleBinding{}
	err := r.Client.Get(context.Background(), types.NamespacedName{Namespace: instance.Namespace, Name: instance.Name}, roleBinding)
	if err != nil && errors.IsNotFound(err) {
		// Create the RoleBinding
		newRoleBinding := r.newRoleBindingForSeleniumTest(instance)
		log.Info("Creating a new RoleBinding", "RoleBinding.Namespace", newRoleBinding.Namespace, "RoleBinding.Name", newRoleBinding.Name)
		err = r.Client.Create(context.Background(), newRoleBinding)
		if err != nil {
			return err
		} else {
			log.Info("RoleBinding ", newRoleBinding.Name, " created")
		}
	} else if err != nil {
		return err
	}

	return nil
}

func (r *SeleniumTestReconciler) ensureCronJob(instance *seleniumv1.SeleniumTest) error {
	cronJob := &batchv1.CronJob{}
	err := r.Client.Get(context.Background(), types.NamespacedName{Namespace: instance.Namespace, Name: instance.Name}, cronJob)
	if err != nil && errors.IsNotFound(err) {
		// Create the CronJob
		newCronJob := r.newCronJobForSeleniumTest(instance)
		log.Info("Creating a new CronJob", "CronJob.Namespace", newCronJob.Namespace, "CronJob.Name", newCronJob.Name)
		err = r.Client.Create(context.Background(), newCronJob)
		if err != nil {
			return err
		} else {
			log.Info("CronJob ", newCronJob.Name, " created")
		}
	} else if err != nil {
		return err
	}

	return nil
}

func (r *SeleniumTestReconciler) newServiceAccountForSeleniumTest(instance *seleniumv1.SeleniumTest) *corev1.ServiceAccount {
	labels := map[string]string{
		"app": "selenium-test",
	}

	var serviceAccount = &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.Name,
			Namespace: instance.Namespace,
			Labels:    labels,
		},
	}

	// Set the owner reference so that the RoleBinding gets deleted when the SeleniumTest is deleted
	controllerutil.SetControllerReference(instance, serviceAccount, r.Scheme)

	return serviceAccount
}

func (r *SeleniumTestReconciler) newRoleBindingForSeleniumTest(instance *seleniumv1.SeleniumTest) *rbacv1.RoleBinding {
	labels := map[string]string{
		"app": "selenium-test",
	}

	var roleBinding = &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.Name,
			Namespace: instance.Namespace,
			Labels:    labels,
		},
		Subjects: []rbacv1.Subject{
			rbacv1.Subject{Kind: "ServiceAccount", Name: instance.Name, Namespace: instance.Namespace},
		},
		RoleRef: rbacv1.RoleRef{APIGroup: "rbac.authorization.k8s.io", Kind: "ClusterRole", Name: "operator-seleniumtestresult-editor-role"},
	}

	// Set the owner reference so that the RoleBinding gets deleted when the SeleniumTest is deleted
	controllerutil.SetControllerReference(instance, roleBinding, r.Scheme)

	return roleBinding
}

func (r *SeleniumTestReconciler) newCronJobForSeleniumTest(instance *seleniumv1.SeleniumTest) *batchv1.CronJob {
	labels := map[string]string{
		"app": "selenium-test",
	}

	var cronJob = &batchv1.CronJob{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.Name,
			Namespace: instance.Namespace,
			Labels:    labels,
		},
		Spec: batchv1.CronJobSpec{
			Schedule:          instance.Spec.Schedule,
			JobTemplate:       batchv1.JobTemplateSpec{Spec: batchv1.JobSpec{Template: corev1.PodTemplateSpec{}}},
			ConcurrencyPolicy: batchv1.ForbidConcurrent,
		},
	}

	// Configure the container template
	// TODO
	container := corev1.Container{
		Name:            "selenium-test",
		Image:           instance.Spec.Repository + "/" + instance.Spec.Image + ":" + instance.Spec.Tag,
		ImagePullPolicy: corev1.PullIfNotPresent,
		Env: []corev1.EnvVar{
			corev1.EnvVar{Name: "POD_NAME", Value: instance.Name},
			corev1.EnvVar{Name: "POD_NAMESPACE", ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{FieldPath: "metadata.namespace"}}},
			corev1.EnvVar{Name: "RETRIES", Value: instance.Spec.Retries},
			corev1.EnvVar{Name: "SELENIUM_GRID", Value: instance.Spec.SeleniumGrid},
		},
	}

	// Create a volume and volume mount for the ConfigMap
	volumeName := "config-volume"
	volume := corev1.Volume{
		Name: volumeName,
		VolumeSource: corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: instance.Spec.ConfigMapName,
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
	cronJob.Spec.JobTemplate.Spec.Template.Spec.RestartPolicy = "OnFailure"
	cronJob.Spec.JobTemplate.Spec.Template.Spec.Volumes = []corev1.Volume{volume}
	cronJob.Spec.JobTemplate.Spec.Template.Spec.Containers[0].VolumeMounts = []corev1.VolumeMount{volumeMount}
	cronJob.Spec.JobTemplate.Spec.Template.Spec.ServiceAccountName = instance.Name

	// Configure the JobBackoffLimit so that failed Jobs are retried
	//    cronJob.Spec.JobBackoffLimit = instance.Spec.JobBackOffLimit

	// Set the owner reference so that the CronJob gets deleted when the SeleniumTest is deleted
	controllerutil.SetControllerReference(instance, cronJob, r.Scheme)

	return cronJob
}

// SetupWithManager sets up the controller with the Manager.
func (r *SeleniumTestReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&seleniumv1.SeleniumTest{}).
		Complete(r)
}

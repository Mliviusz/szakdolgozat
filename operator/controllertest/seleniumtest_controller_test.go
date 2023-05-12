package contro11ers

import (
	"context"
    "testing"

    "k8s.io/apimachinery/pkg/runtime"
    "sigs.k8s.io/controller-runtime/pkg/client"
    "sigs.k8s.io/controller-runtime/pkg/client/fake"
	seleniumv1 "quay.io/molnar_liviusz/selenium-test-operator/api/v1"
)
func TestSeleniumTestController(t *testing.T) {
    seleniumtest := &seleniumv1.SeleniumTest{
        ObjectMeta: metav1.ObjectMeta{
            Name:      "test-seleniumtest",
            Namespace: "default",
            Labels: map[string]string{
				"app": "selenium-test",
			},
        },
		Spec: seleniumv1.SeleniumTestSpec{
			schedule: "*/2 * * * *"
			repository: "quay.io"
			image: "molnar_liviusz/selenium-test-runner"
			tag: "v0.0.10"
			configMapName: "code"
			retries: "3"
			seleniumGrid: "http://moon.moon.svc:4444/wd/hub"
		},
    }

    // Objects to track in the fake client.
    objs := []runtime.Object{seleniumtest}

    // Create a fake client to mock API calls.
    cl := fake.NewFakeClient(objs...)

    s := runtime.NewScheme()
	assert.NoError(t, scheme.AddToScheme(s))
	cl := fake.NewFakeClientWithScheme(s)

	// Test when the instance is not found
	r := &SeleniumTestReconciler{
		Client: cl,
		Scheme: s,
	}
	req := ctrl.Request{
		NamespacedName: client.ObjectKey{
			Namespace: testNamespace,
			Name:      testName,
		},
	}
	res, err := r.Reconcile(context.Background(), req)
	assert.NoError(t, err)
	assert.Equal(t, ctrl.Result{}, res)

	// Test when the instance is found
	assert.NoError(t, cl.Create(context.Background(), testInstance))
	res, err = r.Reconcile(context.Background(), req)
	assert.NoError(t, err)
	assert.Equal(t, ctrl.Result{}, res)

	// Test when the ConfigMap already exists
	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: testNamespace,
			Name:      testName + "-configmap",
		},
	}
	assert.NoError(t, cl.Create(context.Background(), configMap))
	res, err = r.ensureConfigMap(testInstance)
	assert.NoError(t, err)
	assert.Equal(t, nil, res)

	// Test when the ConfigMap does not exist
	configMapName := testName + "-configmap"
	assert.NoError(t, cl.Delete(context.Background(), configMap))
	res, err = r.ensureConfigMap(testInstance)
	assert.NoError(t, err)
	assert.Equal(t, true, errors.IsNotFound(err))
}

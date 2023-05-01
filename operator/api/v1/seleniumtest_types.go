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

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// SeleniumTestSpec defines the desired state of SeleniumTest
type SeleniumTestSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Repository    string `json:"repository"`
	Image         string `json:"image"`
	Tag           string `json:"tag"`
	Schedule      string `json:"schedule"`
	ConfigMapName string `json:"configMapName"`
	Retries       string `json:"retries"`
	SeleniumGrid  string `json:"seleniumGrid"`
}

// SeleniumTestStatus defines the observed state of SeleniumTest
type SeleniumTestStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	CronJobName string `json:"cronJobName"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// SeleniumTest is the Schema for the seleniumtests API
type SeleniumTest struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SeleniumTestSpec   `json:"spec,omitempty"`
	Status SeleniumTestStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// SeleniumTestList contains a list of SeleniumTest
type SeleniumTestList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SeleniumTest `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SeleniumTest{}, &SeleniumTestList{})
}

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

// SeleniumTestResultSpec defines the desired state of SeleniumTestResult
type SeleniumTestResultSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of SeleniumTestResult. Edit seleniumtestresult_types.go to remove/update
	Success bool `json:"success"`
	EndTime int  `json:"endTime"`
}

// SeleniumTestResultStatus defines the observed state of SeleniumTestResult
type SeleniumTestResultStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// SeleniumTestResult is the Schema for the seleniumtestresults API
type SeleniumTestResult struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SeleniumTestResultSpec   `json:"spec,omitempty"`
	Status SeleniumTestResultStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// SeleniumTestResultList contains a list of SeleniumTestResult
type SeleniumTestResultList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SeleniumTestResult `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SeleniumTestResult{}, &SeleniumTestResultList{})
}

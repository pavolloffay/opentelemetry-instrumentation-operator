/*
Copyright 2021.

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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// OpenTelemetryInstrumentationSpec defines the desired state of OpenTelemetryInstrumentation
type OpenTelemetryInstrumentationSpec struct {
	// Foo is an example field of OpenTelemetryInstrumentation. Edit opentelemetryinstrumentation_types.go to remove/update
	OTLPEndpoint     string `json:"OTLPEndpoint,omitempty"`
	JavaagentImage   string `json:"javaagentImage,omitempty"`
	TracesSampler    string `json:"tracesSampler,omitempty"`
	TracesSamplerArg string `json:"tracesSamplerArg,omitempty"`
}

// OpenTelemetryInstrumentationStatus defines the observed state of OpenTelemetryInstrumentation
type OpenTelemetryInstrumentationStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// OpenTelemetryInstrumentation is the Schema for the opentelemetryinstrumentations API
type OpenTelemetryInstrumentation struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   OpenTelemetryInstrumentationSpec   `json:"spec,omitempty"`
	Status OpenTelemetryInstrumentationStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// OpenTelemetryInstrumentationList contains a list of OpenTelemetryInstrumentation
type OpenTelemetryInstrumentationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []OpenTelemetryInstrumentation `json:"items"`
}

func init() {
	SchemeBuilder.Register(&OpenTelemetryInstrumentation{}, &OpenTelemetryInstrumentationList{})
}

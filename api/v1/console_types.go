/*


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

// ConsoleSpec defines the desired state of Console
type ConsoleSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Configuration Configuration `json:"configuration" yaml:"configuration" toml:"configuration" export:"true"`
}

// +k8s:deepcopy-gen=true
// HTTPConfiguration contains all the HTTP configuration parameters.
type Configuration struct {
	Routers map[string]*Router `json:"routers,omitempty" toml:"routers,omitempty" yaml:"routers,omitempty" export:"true"`
}

// +k8s:deepcopy-gen=true

// Router holds the router configuration.
type Router struct {
	// Backend Server Name
	Server string `json:"server,omitempty" toml:"server,omitempty" yaml:"server,omitempty" export:"true"`
	// Reverse proxy rule
	Rule string `json:"rule,omitempty" toml:"rule,omitempty" yaml:"rule,omitempty" export:"true"`
	// If want to remove url path when proxy server request backend
	Path string `json:"path,omitempty" yaml:"path,omitempty" toml:"path,omitempty" export:"true"`
}

// ConsoleStatus defines the observed state of Console
type ConsoleStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Show the number of backend proxy
	Number int `json:"number,omitempty" toml:"number,omitempoty" yaml:"routers,omitempoty" export:"true"`
	// Config information
	Routers string `json:"routers,omitempty" toml:"routers,omitempty" yaml:"routers,omitempty" export:"true"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:JSONPath=".status.number",name="READY",type="integer"
// +kubebuilder:printcolumn:JSONPath=".status.routers",name="ROUTERS",type="string"

// Console is the Schema for the consoles API
type Console struct {
	metav1.TypeMeta   `json:",inline" yaml:",inline" toml:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty" yaml:"metadata,omitempty" toml:"metadata,omitempty"`

	Spec   ConsoleSpec   `json:"spec,omitempty" yaml:"spec,omitempty" toml:"spec,omitempty"`
	Status ConsoleStatus `json:"status,omitempty" yaml:"status,omitempty" toml:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ConsoleList contains a list of Console
type ConsoleList struct {
	metav1.TypeMeta `json:",inline" yaml:",inline" toml:",inline"`
	metav1.ListMeta `json:"metadata,omitempty" yaml:"metadata,omitempty" toml:"metadata,omitempty"`
	Items           []Console `json:"items" yaml:"items" toml:"items"`
}

func init() {
	SchemeBuilder.Register(&Console{}, &ConsoleList{})
}

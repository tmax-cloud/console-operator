package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ConsoleSpec defines the desired state of Console
type ConsoleSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
	ConsoleApp ConsoleApp `json:"consoleApp,omitempty"`
}

type ConsoleApp struct {
	Repository      string             `json:"repository,omitempty"`
	Tag             string             `json:"tag,omitempty"`
	ImagePullPolicy corev1.PullPolicy  `json:"imagePullPolicy,omitempty"`
	Replicas        int32              `json:"replicas,omitempty"`
	Port            int32              `json:"port,omitempty"`
	TargetPort      int                `json:"targetPort,omitempty"`
	ServiceType     corev1.ServiceType `json:"serviceType,omitempty"`
}

// ConsoleStatus defines the observed state of Console
type ConsoleStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Console is the Schema for the consoles API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=consoles,scope=Namespaced
type Console struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ConsoleSpec   `json:"spec,omitempty"`
	Status ConsoleStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ConsoleList contains a list of Console
type ConsoleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Console `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Console{}, &ConsoleList{})
}

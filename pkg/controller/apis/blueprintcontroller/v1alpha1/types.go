package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Database is a specification for a Database resource
type Database struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DatabaseSpec   `json:"spec"`
	Status DatabaseStatus `json:"status"`
}

// DatabaseSpec is the spec for a Database resource
type DatabaseSpec struct {
	Host HostField `json:"host"`
	App  AppField  `json:"app"`
}

type HostField struct {
	Ip       string `json:"ip"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type AppField struct {
	Script ScriptField `json:"script"`
}

type ScriptField struct {
	Install   string `json:"install"`
	Start     string `json:"start"`
	Stop      string `json:"stop"`
	Restart   string `json:"restart"`
	Uninstall string `json:"uninstall"`
}

// DatabaseStatus is the status for a Database resource
type DatabaseStatus struct {
	Install string `json:"install"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// DatabaseList is a list of Database resources
type DatabaseList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []Database `json:"items"`
}

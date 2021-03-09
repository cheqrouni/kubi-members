package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ResourceQuotaClaim defines a request modify a ResourcesQuota
type ProjectMembers struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Status  ProjectMembersStatus `json:"status,omitempty"`
	Members []ProjectMember      `json:"members,omitempty"`
}

type ProjectMember struct {
	Dn          string `json:"dn,omitempty"`
	Cn          string `json:"cn,omitempty"`
	DisplayName string `json:"displayName,omitempty"`
	Mail        string `json:"mail,omitempty"`
}

const (
	PhaseHealthy = "HEALTHY"
	PhaseError   = "ERROR"
)

// ProjectMembersStatus defines the observed state of ResourceQuotaClaim
type ProjectMembersStatus struct {
	Phase   string `json:"phase,omitempty"`
	Details string `json:"details,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ProjectMembersList contains a list of ResourceQuotaClaim
type ProjectMembersList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ProjectMembers `json:"items"`
}

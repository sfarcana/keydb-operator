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

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// KeyDBSpec defines the desired state of KeyDB
type KeyDBSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of KeyDB. Edit KeyDB_types.go to remove/update
	Replicas    int32  `json:"replicas"`
	ImageTag    string `json:"imagetag"`
	Mode        string `json:"mode"`
	ServiceType string `json:"servicetype"`

	ImagePullPolicy    corev1.PullPolicy           `json:"imagePullPolicy,omitempty"`
	Resources          corev1.ResourceRequirements `json:"resources,omitempty"`
	Storage            RedisStorage                `json:"storage,omitempty"`
	Exporter           RedisExporter               `json:"exporter,omitempty"`
	Affinity           *corev1.Affinity            `json:"affinity,omitempty"`
	Tolerations        []corev1.Toleration         `json:"tolerations,omitempty"`
	NodeSelector       map[string]string           `json:"nodeSelector,omitempty"`
	ServiceAnnotations map[string]string           `json:"serviceAnnotations,omitempty"`
}

// RedisExporter defines the specification for the redis exporter
type RedisExporter struct {
	Enabled         bool              `json:"enabled,omitempty"`
	Image           string            `json:"image,omitempty"`
	ImagePullPolicy corev1.PullPolicy `json:"imagePullPolicy,omitempty"`
}

// RedisStorage defines the structure used to store the Redis Data
type RedisStorage struct {
	KeepAfterDeletion     bool                          `json:"keepAfterDeletion,omitempty"`
	EmptyDir              *corev1.EmptyDirVolumeSource  `json:"emptyDir,omitempty"`
	PersistentVolumeClaim *corev1.PersistentVolumeClaim `json:"persistentVolumeClaim,omitempty"`
}

// KeyDBStatus defines the observed state of KeyDB
type KeyDBStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Nodes []string `json:"nodes"`
	//Services []string `json:"service"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// KeyDB is the Schema for the keydbs API
type KeyDB struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KeyDBSpec   `json:"spec,omitempty"`
	Status KeyDBStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// KeyDBList contains a list of KeyDB
type KeyDBList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []KeyDB `json:"items"`
}

func init() {
	SchemeBuilder.Register(&KeyDB{}, &KeyDBList{})
}

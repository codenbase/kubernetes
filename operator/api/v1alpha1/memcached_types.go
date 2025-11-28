/*
Copyright 2025.

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

// 重要提醒：这个文件修改后，需要执行 `make` 命令重新生成代码。

// MemcachedSpec 定义了 Memcached 的期望状态
type MemcachedSpec struct {
	// Size 定义了 Memcached 集群的期望副本数量
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=5
	Size int32 `json:"size,omitempty"`
}

// MemcachedStatus 定义了 Memcached 的观察状态（实际运行状态）
type MemcachedStatus struct {
	// Nodes 记录了当前运行的 Pod 名称列表
	// +optional
	Nodes []string `json:"nodes,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Memcached is the Schema for the memcacheds API
type Memcached struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitzero"`

	// spec defines the desired state of Memcached
	// +required
	Spec MemcachedSpec `json:"spec"`

	// status defines the observed state of Memcached
	// +optional
	Status MemcachedStatus `json:"status,omitzero"`
}

// +kubebuilder:object:root=true

// MemcachedList contains a list of Memcached
type MemcachedList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitzero"`
	Items           []Memcached `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Memcached{}, &MemcachedList{})
}

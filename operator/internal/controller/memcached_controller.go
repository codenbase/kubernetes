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

package controller

import (
	"context"
	"reflect"
	"sort"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	// 注意：这里的导入路径需要根据你实际的 go.mod 模块名进行修改
	cachev1alpha1 "example.com/operator/api/v1alpha1"
)

// MemcachedReconciler 负责调谐 Memcached 对象
// 它继承了 client.Client，因此可以直接使用 r.Get, r.List 等方法
type MemcachedReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// 下面的 //+kubebuilder 注释非常重要！
// controller-gen 会读取这些注释，自动生成 RBAC 规则（config/rbac/role.yaml）

// 1. 允许操作 Memcached 核心资源
// +kubebuilder:rbac:groups=cache.example.com,resources=memcacheds,verbs=get;list;watch;create;update;patch;delete
// 2. 允许操作 Memcached 的 Status 子资源
// +kubebuilder:rbac:groups=cache.example.com,resources=memcacheds/status,verbs=get;update;patch
// 3. 允许操作 Memcached 的 Finalizers（用于删除逻辑）
// +kubebuilder:rbac:groups=cache.example.com,resources=memcacheds/finalizers,verbs=update
// 4. 【新增】允许操作 Deployment（因为我们要创建和管理它）
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// 5. 【新增】允许操作 Pod（因为我们要读取 Pod 列表来更新 Status）
// +kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch

// Reconcile 是 Kubernetes 调谐循环的核心。
// 它的目标是让集群的“当前状态”不断逼近用户的“期望状态”。
// 每当 Memcached CR 发生变化，或者它拥有的 Deployment 发生变化时，这个函数都会被触发。
func (r *MemcachedReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	// 获取上下文中的 Logger
	logger := log.FromContext(ctx)

	// ------------------------------------------------------------------
	// 步骤 1: 获取 Memcached 实例 (CR)
	// ------------------------------------------------------------------
	memcached := &cachev1alpha1.Memcached{}
	if err := r.Get(ctx, req.NamespacedName, memcached); err != nil {
		// 如果错误是“未找到”，说明该 CR 可能已经被删除了。
		// 在这种情况下，我们不需要报错，直接返回空结果即可。
		// 关联的 Deployment 会因为 OwnerReference 被 K8s 垃圾回收机制自动删除。
		//
		// client.IgnoreNotFound(err) 会在资源不存在时返回 nil。
		// 为什么不返回 error？因为资源被删除了是“终态”。
		// 如果我们返回 error，控制器框架会认为处理失败，会进行指数退避重试（Requeue），
		// 但资源已经没了，重试除了浪费 CPU 没有任何意义。
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// ------------------------------------------------------------------
	// 步骤 2: 检查并管理关联的 Deployment
	// ------------------------------------------------------------------
	// 定义一个 Deployment 变量，用于接收查询结果
	found := &appsv1.Deployment{}
	err := r.Get(ctx, types.NamespacedName{Name: memcached.Name, Namespace: memcached.Namespace}, found)

	// 情况 A: Deployment 不存在 -> 创建它
	if err != nil && errors.IsNotFound(err) {
		// 调用辅助函数构建期望的 Deployment 对象
		dep, err := r.deploymentForMemcached(memcached)
		if err != nil {
			logger.Error(err, "构建 Deployment 对象失败")
			return ctrl.Result{}, err
		}

		logger.Info("正在创建新的 Deployment", "Namespace", dep.Namespace, "Name", dep.Name)

		if err = r.Create(ctx, dep); err != nil {
			logger.Error(err, "创建 Deployment 失败")
			return ctrl.Result{}, err
		}

		// 创建成功后，Deployment 的变更会再次触发调谐循环。
		return ctrl.Result{}, nil

	} else if err != nil {
		// 情况 B: 查询出错 -> 返回错误
		logger.Error(err, "获取 Deployment 失败")
		return ctrl.Result{}, err
	}

	// ------------------------------------------------------------------
	// 步骤 3: 调谐 (更新 Deployment 规格)
	// ------------------------------------------------------------------
	// 此时 Deployment 已经存在，我们需要检查它的配置是否符合预期。
	// 这里我们只检查 Replicas（副本数），实际项目中可能还需要检查 Image、Env 等。
	size := memcached.Spec.Size
	if found.Spec.Replicas == nil || *found.Spec.Replicas != size {
		// 将 Deployment 的副本数量设置为期望的副本数
		found.Spec.Replicas = &size

		logger.Info("检测到副本数不一致，正在更新 Deployment", "当前值", *found.Spec.Replicas, "期望值", size)
		if err = r.Update(ctx, found); err != nil {
			logger.Error(err, "更新 Deployment 失败")
			return ctrl.Result{}, err
		}

		// Deployment 的更新，会再次触发调谐循环，执行新一轮的检查。
		// 这里不需要设置 Requeue: true。
		// 因为只要 Update 执行成功，API Server 就会产生一个 Update 事件。
		// 我们的 SetupWithManager 里的 Owns() 监听到了这个事件，
		// 会自动触发下一次 Reconcile，进行再次确认。
		return ctrl.Result{}, nil
	}

	// ------------------------------------------------------------------
	// 步骤 4: 更新 Memcached 的 Status
	// ------------------------------------------------------------------
	// 我们希望在 Memcached CR 的 Status 中看到当前实际运行的 Pod 名称列表。

	// 列出所有属于该 Memcached 的 Pod
	podList := &corev1.PodList{}
	listOpts := []client.ListOption{
		client.InNamespace(memcached.Namespace),
		client.MatchingLabels(labelsForMemcached(memcached.Name)),
	}
	if err = r.List(ctx, podList, listOpts...); err != nil {
		logger.Error(err, "列出 Pod 失败")
		return ctrl.Result{}, err
	}

	// 获取 Pod 名称列表
	podNames := getPodNames(podList.Items)

	// 【优化】必须先排序，防止因为列表顺序不同导致 DeepEqual 误判
	sort.Strings(podNames)
	sort.Strings(memcached.Status.Nodes)

	// 如果 Status 中的状态与实际查询到的不一致，则更新 Status
	if !reflect.DeepEqual(podNames, memcached.Status.Nodes) {
		memcached.Status.Nodes = podNames
		logger.Info("正在更新 Memcached Status", "Pod列表", podNames)

		if err := r.Status().Update(ctx, memcached); err != nil {
			logger.Error(err, "更新 Status 失败")
			return ctrl.Result{}, err
		}
	}

	// 一切正常，直接返回
	return ctrl.Result{}, nil
}

// SetupWithManager 将 Controller 注册到 Manager 中
func (r *MemcachedReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		// For: 监听 Memcached 资源本身的增删改事件
		For(&cachev1alpha1.Memcached{}).

		// Owns: 监听“属于” Memcached 的 Deployment 变更。
		// 【关键】这里内置了过滤逻辑：只有当 Deployment 的 metadata.ownerReferences
		// 指向了某个 Memcached CR 时，该 Deployment 的事件才会触发 Reconcile。
		// 这意味着集群里其他无关的 Deployment 变化不会打扰到我们。
		Owns(&appsv1.Deployment{}).
		Complete(r)
}

// ------------------------------------------------------------------
// 辅助函数部分
// ------------------------------------------------------------------

// deploymentForMemcached 根据 Memcached CR 生成对应的 Deployment 对象
func (r *MemcachedReconciler) deploymentForMemcached(m *cachev1alpha1.Memcached) (*appsv1.Deployment, error) {
	// 定义标签
	ls := labelsForMemcached(m.Name)
	// 定义副本数
	replicas := m.Spec.Size

	// 构造 Deployment 对象
	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      m.Name,
			Namespace: m.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: ls,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: ls,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Image:   "memcached:1.6.39-alpine", // 示例中使用固定镜像，实际可从 Spec 中获取
						Name:    "memcached",
						Command: []string{"memcached", "--memory-limit=64", "-o", "modern", "-v"},
						Ports: []corev1.ContainerPort{{
							ContainerPort: 11211,
							Name:          "memcached",
						}},
					}},
				},
			},
		},
	}

	// 【重要】设置 OwnerReference
	// 这步操作将 Deployment 设置为 Memcached 的“子资源”。
	// SetControllerReference 有两个核心作用：
	// 1. 垃圾回收（GC）：当 Memcached CR 被删除时，K8s 会级联删除这个 Deployment。
	// 2. 事件过滤：它在 Deployment 中注入了 OwnerReference，让 Controller 的 Owns() 方法
	//    能够识别出这个 Deployment 是属于我们的，从而建立 Watch 关系。
	if err := ctrl.SetControllerReference(m, dep, r.Scheme); err != nil {
		return nil, err
	}

	return dep, nil
}

// labelsForMemcached 返回选择 Pod 的通用标签
func labelsForMemcached(name string) map[string]string {
	return map[string]string{"app": "memcached", "memcached_cr": name}
}

// getPodNames 提取 Pod 列表中的名称
func getPodNames(pods []corev1.Pod) []string {
	var podNames []string
	for _, pod := range pods {
		podNames = append(podNames, pod.Name)
	}
	return podNames
}

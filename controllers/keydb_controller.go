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

package controllers

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/go-logr/logr"
	"github.com/pingcap/errors"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"

	cachev1alpha1 "github.com/sfarcana/keydb-operator/api/v1alpha1"
)

//const SerectUtils = "\n" +
//	"#!/bin/bash" +
//	"\nset -euxo pipefail" +
//	"\nhost=\"$(hostname)\"" +
//	"\nport=\"6379\"" +
//	"\nreplicas=()" +
//	"\nfor node in {0..1}; do" +
//	"\n  if [ \"$host\" != \"keydb-sample-${node}\" ]; then" +
//	"\n      replicas+=(\"--replicaof keydb-sample-${node}.keydb-sample ${port}\")" +
//	"\n  fi" +
//	"\ndone" +
//	"\nexec keydb-server /etc/keydb/redis.conf \\" +
//	"\n    --active-replica yes \\" +
//	"\n    --multi-master yes \\" +
//	"\n    --appendonly no \\" +
//	"\n    --bind 0.0.0.0 \\" +
//	"\n    --port \"$port\" \\" +
//	"\n    --protected-mode no \\" +
//	"\n    --server-threads 2 \\" +
//	"\n    \"${replicas[@]}\""

const templateSecretsUtils = "\n" +
	"#!/bin/bash" +
	"\nset -euxo pipefail" +
	"\nhost=\"$(hostname)\"" +
	"\nport=\"6379\"" +
	"\nreplicas=()" +
	"\nfor node in {0..var-number-nodes}; do" +
	"\n  if [ \"$host\" != \"var-keydb-sample-${node}\" ]; then" +
	"\n      replicas+=(\"--replicaof var-keydb-sample-${node}.var-keydb-sample ${port}\")" +
	"\n  fi" +
	"\ndone" +
	"\nexec keydb-server /etc/keydb/redis.conf \\" +
	"\n    --active-replica yes \\" +
	"\n    --multi-master yes \\" +
	"\n    --appendonly no \\" +
	"\n    --bind 0.0.0.0 \\" +
	"\n    --port \"$port\" \\" +
	"\n    --protected-mode no \\" +
	"\n    --server-threads 2 \\" +
	"\n    \"${replicas[@]}\""

// KeyDBReconciler reconciles a KeyDB object
type KeyDBReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=cache.tiki.vn,resources=keydbs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cache.tiki.vn,resources=keydbs/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;

func (r *KeyDBReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("keydb", req.NamespacedName)
	keydb := &cachev1alpha1.KeyDB{}
	err := r.Get(ctx, req.NamespacedName, keydb)
	// your logic here

	// Fetch the KeyDB instance

	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			log.Info("KeyDB resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		log.Error(err, "Failed to get KeyDB")
		return ctrl.Result{}, err
	}

	// Check if the serect already exists, if not create a new one
	foundSerect := &corev1.Secret{}
	err = r.Get(ctx, types.NamespacedName{Name: keydb.Name + "-utils", Namespace: keydb.Namespace}, foundSerect)
	if err != nil && errors.IsNotFound(err) {
		// Define a new Serect
		dep := r.secretForKeydb(keydb)

		log.Info("Creating a new Serect", "Serect.Namespace", dep.Namespace, "Serect.Name", dep.Name)
		err = r.Create(ctx, dep)
		if err != nil {
			log.Error(err, "Failed to create new Serect", "Serect.Namespace", dep.Namespace, "Serect.Name", dep.Name)
			return ctrl.Result{}, err
		}

		// Secrect created successfully - return and requeue
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		log.Error(err, "Failed to get Serect")
		return ctrl.Result{}, err
	}

	// Check if the statefulset already exists, if not create a new one
	found := &appsv1.StatefulSet{}
	err = r.Get(ctx, types.NamespacedName{Name: keydb.Name, Namespace: keydb.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		// Define a new statefulset
		dep := r.statefulsetForKeydb(keydb)

		log.Info("Creating a new Statefulset", "Statefulset.Namespace", dep.Namespace, "Statefulset.Name", dep.Name)
		err = r.Create(ctx, dep)
		if err != nil {
			log.Error(err, "Failed to create new Statefulset", "Statefulset.Namespace", dep.Namespace, "Statefulset.Name", dep.Name)
			return ctrl.Result{}, err
		}
		// Statefulset created successfully - return and requeue
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		log.Error(err, "Failed to get Statefulset")
		return ctrl.Result{}, err
	}

	// Ensure the deployment size is the same as the spec
	size := keydb.Spec.Replicas
	if *found.Spec.Replicas != size {
		found.Spec.Replicas = &size
		err = r.Update(ctx, found)
		if err != nil {
			log.Error(err, "Failed to update Statefulset", "Statefulset.Namespace", found.Namespace, "Statefulset.Name", found.Name)
			return ctrl.Result{}, err
		}
		// Spec updated - return and requeue
		return ctrl.Result{Requeue: true}, nil
	}

	// Create Service

	// Check if the Service already exists, if not create a new one
	foundServices := &corev1.Service{}
	err = r.Get(ctx, types.NamespacedName{Name: keydb.Name, Namespace: keydb.Namespace}, foundServices)
	if err != nil && errors.IsNotFound(err) {
		// Define a new Service
		dep := r.serviceForKeydb(keydb)

		log.Info("Creating a new Service", "Service.Namespace", dep.Namespace, "Service.Name", dep.Name)
		err = r.Create(ctx, dep)
		if err != nil {
			log.Error(err, "Failed to create new Service", "Service.Namespace", dep.Namespace, "Service.Name", dep.Name)
			return ctrl.Result{}, err
		}
		// Service created successfully - return and requeue
		return ctrl.Result{Requeue: true}, nil

	} else if err != nil {
		log.Error(err, "Failed to get Service")
		return ctrl.Result{}, err
	}

	// Ensure the service services is the same as the spec
	svcType := corev1.ServiceType(keydb.Spec.ServiceType)
	if *&foundServices.Spec.Type != svcType {
		foundServices.Spec.Type = svcType
		err = r.Update(ctx, foundServices)
		if err != nil {
			log.Error(err, "Failed to update Services", "Services.Namespace", foundServices.Namespace, "Services.Name", foundServices.Name)
			return ctrl.Result{}, err
		}
		// Spec updated - return and requeue
		return ctrl.Result{Requeue: true}, nil
	}

	// Update the keydb status with the pod names
	// List the pods for this keydb's statefulset
	podList := &corev1.PodList{}
	listOpts := []client.ListOption{
		client.InNamespace(keydb.Namespace),
		client.MatchingLabels(labelsForKeydb(keydb.Name)),
	}
	if err = r.List(ctx, podList, listOpts...); err != nil {
		log.Error(err, "Failed to list pods", "KeyDB.Namespace", keydb.Namespace, "keydb.Name", keydb.Name)
		return ctrl.Result{}, err
	}
	podNames := getPodNames(podList.Items)

	fmt.Println("List podNames: ", podNames)

	// Update status.Nodes if needed
	if !reflect.DeepEqual(podNames, keydb.Status.Nodes) {
		keydb.Status.Nodes = podNames
		err := r.Status().Update(ctx, keydb)
		if err != nil {
			log.Error(err, "Failed to update KeyDb status")
			return ctrl.Result{}, err
		}
	}

	// // // List the service for this keydb's statefulset
	// serviceList := &corev1.ServiceList{}
	// servicelistOpts := []client.ListOption{
	// 	client.InNamespace(keydb.Namespace),
	// 	client.MatchingLabels(labelsForKeydb(keydb.Name)),
	// }
	// if err = r.List(ctx, serviceList, servicelistOpts...); err != nil {
	// 	log.Error(err, "Failed to list service", "KeyDB.Namespace", keydb.Namespace, "keydb.Name", keydb.Name)
	// 	return ctrl.Result{}, err
	// }

	// var serviceNames []string
	// for _, serviceName := range serviceList.Items {
	// 	fmt.Println("Service Name is:  ", serviceName)
	// 	serviceNames = append(serviceNames, serviceName.Name)
	// }

	// fmt.Println("List servicesList: ", serviceList.Items[0].Name)
	// fmt.Println("List servicesList new: ", serviceNames)

	// // Update status.Service if needed
	// if !reflect.DeepEqual(serviceList.Items, keydb.Status.Services) {
	// 	//keydb.Status.Services[0] = serviceList.Items[0].Name
	// 	keydb.Status.Services = serviceNames
	// 	err := r.Status().Update(ctx, keydb)
	// 	if err != nil {
	// 		log.Error(err, "Failed to update KeyDb status")
	// 		return ctrl.Result{}, err
	// 	}
	// }

	return ctrl.Result{RequeueAfter: time.Second * 5}, nil
}

func (r *KeyDBReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&cachev1alpha1.KeyDB{}).
		Owns(&appsv1.StatefulSet{}).
		WithOptions(controller.Options{
			MaxConcurrentReconciles: 2,
		}).
		Complete(r)
}

// secretForKeydb return serect for keydb
func (r *KeyDBReconciler) secretForKeydb(m *cachev1alpha1.KeyDB) *corev1.Secret {
	objMetaSecretName := m.Name + "-utils"

	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:            objMetaSecretName,
			Namespace:       m.Namespace,
			Labels:          labelsForKeydb(m.Name),
			OwnerReferences: m.OwnerReferences,
		},
		Type:       "Opaque",
		StringData: generateSecretUtils(m.Name, m.Spec.Replicas),
	}
}

// serviceForKeydb return service for Keydb
func (r *KeyDBReconciler) serviceForKeydb(m *cachev1alpha1.KeyDB) *corev1.Service {
	name := m.Name
	namespace := m.Namespace

	keydbTargetPort := intstr.FromInt(6379)
	labels := labelsForKeydb(m.Name)

	serviceType := m.Spec.ServiceType

	// save service type
	var tmp corev1.ServiceType

	// if type ClusterIP
	if serviceType == "ClusterIP" {
		tmp = corev1.ServiceTypeClusterIP
	}

	// if type LoadBalancer
	if serviceType == "LoadBalancer" {
		tmp = corev1.ServiceTypeLoadBalancer
	}

	// return service
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:            name,
			Namespace:       namespace,
			Labels:          labels,
			OwnerReferences: m.OwnerReferences,
			Annotations:     m.Spec.ServiceAnnotations,
		},
		Spec: corev1.ServiceSpec{
			Selector: labels,
			Ports: []corev1.ServicePort{
				{
					Name:       "keydb",
					Port:       6379,
					TargetPort: keydbTargetPort,
					Protocol:   "TCP",
				},
			},
			Type: tmp,
		},
	}

}

// statefulsetForKeydb returns a Keydb statefulset object
func (r *KeyDBReconciler) statefulsetForKeydb(m *cachev1alpha1.KeyDB) *appsv1.StatefulSet {
	// Label for keydb
	ls := labelsForKeydb(m.Name)

	// Number of Pods
	replicas := m.Spec.Replicas

	// image tag, if not set will be latest
	imageTag := m.Spec.ImageTag

	mode := int32(0444)

	if len(imageTag) == 0 {
		imageTag = "latest"
	}

	dep := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      m.Name,
			Namespace: m.Namespace,
		},
		Spec: appsv1.StatefulSetSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: ls,
			},
			ServiceName: m.Name,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: ls,
				},
				Spec: corev1.PodSpec{
					Affinity:     m.Spec.Affinity,
					Tolerations:  m.Spec.Tolerations,
					NodeSelector: m.Spec.NodeSelector,
					Containers: []corev1.Container{{
						Image:           "eqalpha/keydb:" + imageTag,
						ImagePullPolicy: m.Spec.ImagePullPolicy,
						Name:            "keydb",
						Command:         []string{"/bin/bash", "-x", "/utils/server.sh"},
						Ports: []corev1.ContainerPort{{
							ContainerPort: 6379,
							Name:          "keydb",
						}},
						Resources: corev1.ResourceRequirements{
							Requests: m.Spec.Resources.Requests,
							Limits:   m.Spec.Resources.Limits,
						},
						VolumeMounts: []corev1.VolumeMount{{
							Name:      "utils",
							MountPath: "/utils",
							ReadOnly:  true,
						}, {
							Name:      "keydb-data",
							MountPath: "/data",
						}},
					}},
					Volumes: []corev1.Volume{{
						Name: "utils",
						VolumeSource: corev1.VolumeSource{
							Secret: &corev1.SecretVolumeSource{
								SecretName:  m.Name + "-utils",
								DefaultMode: &mode,
								Items: []corev1.KeyToPath{{
									Key:  "server.sh",
									Path: "server.sh",
								}},
							},
						},
					}, {
						Name: "keydb-data",
					}},
				},
			},
		},
	}
	// Set Keydb instance as the owner and controller
	ctrl.SetControllerReference(m, dep, r.Scheme)
	return dep
}

// labelsForKeydb returns the labels for selecting the resources
// belonging to the given Keydb CR name.
func labelsForKeydb(name string) map[string]string {
	return map[string]string{"app": "keydb", "keydb_cr": name}
}

// getPodNames returns the pod names of the array of pods passed in
func getPodNames(pods []corev1.Pod) []string {
	var podNames []string
	for _, pod := range pods {
		podNames = append(podNames, pod.Name)
	}
	return podNames
}

// MergeLabels merges all the label maps received as argument into a single new label map.
func MergeLabels(allLabels ...map[string]string) map[string]string {
	res := map[string]string{}

	for _, labels := range allLabels {
		if labels != nil {
			for k, v := range labels {
				res[k] = v
			}
		}
	}
	return res
}

// generateSelectorLabels create labels for app
func generateSelectorLabels(component, appLabel, name string) map[string]string {
	return map[string]string{
		"app.kubernetes.io/name":      name,
		"app.kubernetes.io/component": component,
		"app.kubernetes.io/part-of":   appLabel,
	}
}

// generateSecretUtils create server.sh for app
func generateSecretUtils(svcName string, numNodes int32) map[string]string {
	utilValues := templateSecretsUtils

	num := fmt.Sprint(numNodes - 1)

	// Replace with number nodes
	utilValues = strings.ReplaceAll(utilValues, "var-number-nodes", num)

	// Replace with name node
	utilValues = strings.ReplaceAll(utilValues, "var-keydb-sample", svcName)

	//value := strings.ReplaceAll(SerectUtils, "keydb-ultis", svcName)
	return map[string]string{
		"server.sh": utilValues,
	}
}

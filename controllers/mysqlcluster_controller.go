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

package controllers

import (
	"context"
	"github.com/imind-lab/mysql-operator/controllers/utils"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	dbv1beta1 "github.com/imind-lab/mysql-operator/api/v1beta1"
)

// MySQLClusterReconciler reconciles a MySQLCluster object
type MySQLClusterReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=db.imind.tech,resources=mysqlclusters,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=db.imind.tech,resources=mysqlclusters/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=db.imind.tech,resources=mysqlclusters/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the MySQLCluster object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.1/pkg/reconcile
func (r *MySQLClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	mysqlCluster := &dbv1beta1.MySQLCluster{}
	if err := r.Get(ctx, req.NamespacedName, mysqlCluster); err != nil {
		logger.Error(err, "The mysql cluster has been deleted")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	service := utils.NewService(mysqlCluster)
	if err := controllerutil.SetControllerReference(mysqlCluster, service, r.Scheme); err != nil {
		return ctrl.Result{}, err
	}

	svc := &corev1.Service{}
	if err := r.Get(ctx, types.NamespacedName{Name: mysqlCluster.Name, Namespace: mysqlCluster.Namespace}, svc); err != nil {
		if errors.IsNotFound(err) {
			if err := r.Create(ctx, service); err != nil {
				logger.Error(err, "create service failed")
				return ctrl.Result{}, err
			}
		} else {
			return ctrl.Result{}, err
		}
	}

	configMap := utils.NewConfigMap(mysqlCluster)
	if err := controllerutil.SetControllerReference(mysqlCluster, configMap, r.Scheme); err != nil {
		return ctrl.Result{}, err
	}

	cm := &corev1.ConfigMap{}
	if err := r.Get(ctx, types.NamespacedName{Name: mysqlCluster.Name, Namespace: mysqlCluster.Namespace}, cm); err != nil {
		if errors.IsNotFound(err) {
			if err := r.Create(ctx, configMap); err != nil {
				logger.Error(err, "create configmap failed")
				return ctrl.Result{}, err
			}
		}
	}

	statefulSet := utils.NewStatefulSet(mysqlCluster)
	if err := controllerutil.SetControllerReference(mysqlCluster, statefulSet, r.Scheme); err != nil {
		return ctrl.Result{}, err
	}

	sts := &appsv1.StatefulSet{}
	if err := r.Get(ctx, types.NamespacedName{Name: mysqlCluster.Name, Namespace: mysqlCluster.Namespace}, sts); err != nil {
		if errors.IsNotFound(err) {
			if err := r.Create(ctx, statefulSet); err != nil {
				logger.Error(err, "create statefulset failed")
				return ctrl.Result{}, err
			}
		} else {
			return ctrl.Result{}, err
		}
	} else {
		if *sts.Spec.Replicas != *statefulSet.Spec.Replicas || sts.Spec.Template.Spec.Containers[0].Image != statefulSet.Spec.Template.Spec.Containers[0].Image {
			*sts.Spec.Replicas = *statefulSet.Spec.Replicas
			sts.Spec.Template.Spec.Containers[0].Image = statefulSet.Spec.Template.Spec.Containers[0].Image
			if err := r.Update(ctx, sts); err != nil {
				logger.Error(err, "update statefulset failed")
				return ctrl.Result{}, err
			}
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *MySQLClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&dbv1beta1.MySQLCluster{}).
		Complete(r)
}

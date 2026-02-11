/*
Copyright 2026.

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
	"encoding/json"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/util/retry"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	intstr "k8s.io/apimachinery/pkg/util/intstr"

	examplev1 "github.com/msfurusa/json-server-controller/api/v1"
)

// JsonServerReconciler reconciles a JsonServer object
type JsonServerReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.23.1/pkg/reconcile
func (r *JsonServerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := logf.FromContext(ctx)

	var js examplev1.JsonServer
	if err := r.Get(ctx, req.NamespacedName, &js); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Validate JSON config; if invalid, update status and return
	var tmp interface{}
	if err := json.Unmarshal([]byte(js.Spec.JsonConfig), &tmp); err != nil {
		js.Status.State = "Error"
		js.Status.Message = "Error: spec.jsonConfig is not a valid json object"
		if uerr := r.Status().Update(ctx, &js); uerr != nil {
			logger.Error(uerr, "failed to update status")
			return ctrl.Result{}, uerr
		}
		return ctrl.Result{}, nil
	}

	name := js.Name
	namespace := js.Namespace

	// ConfigMap (CreateOrUpdate)
	cm := &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace}}
	_, err := controllerutil.CreateOrUpdate(ctx, r.Client, cm, func() error {
		if cm.Data == nil {
			cm.Data = map[string]string{}
		}
		cm.Data["db.json"] = js.Spec.JsonConfig
		return controllerutil.SetControllerReference(&js, cm, r.Scheme)
	})
	if err != nil {
		logger.Error(err, "failed to create or update configmap")
		js.Status.State = "Error"
		js.Status.Message = "Error: json-server unexpected failure"
		_ = r.Status().Update(ctx, &js)
		return ctrl.Result{}, err
	}

	// Deployment
	replicas := int32(1)
	if js.Spec.Replicas != nil {
		replicas = *js.Spec.Replicas
	}

	labels := map[string]string{"app": name}

	image := js.Spec.Image
	if image == "" {
		image = "backplane/json-server"
	}

	desiredDeploy := appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace, Labels: labels},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{MatchLabels: labels},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: labels},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Name:            "json-server",
						Image:           image,
						ImagePullPolicy: corev1.PullIfNotPresent,
						Args:            []string{"/data/db.json"},
						Ports:           []corev1.ContainerPort{{ContainerPort: 3000}},
						VolumeMounts:    []corev1.VolumeMount{{Name: "data", MountPath: "/data"}},
					}},
					Volumes: []corev1.Volume{{
						Name: "data",
						VolumeSource: corev1.VolumeSource{
							ConfigMap: &corev1.ConfigMapVolumeSource{
								LocalObjectReference: corev1.LocalObjectReference{Name: name},
								Items:                []corev1.KeyToPath{{Key: "db.json", Path: "db.json"}},
							},
						},
					}},
				},
			},
		},
	}

	// Deployment (CreateOrUpdate)
	deploy := &appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace}}
	err = retry.RetryOnConflict(retry.DefaultRetry, func() error {
		if err := r.Get(ctx, client.ObjectKeyFromObject(deploy), deploy); err != nil {
			if apierrors.IsNotFound(err) {
				deploy = &appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace}}
			} else {
				return err
			}
		}
		_, err := controllerutil.CreateOrUpdate(ctx, r.Client, deploy, func() error {
			deploy.Labels = labels
			deploy.Spec = desiredDeploy.Spec
			return controllerutil.SetControllerReference(&js, deploy, r.Scheme)
		})
		return err
	})
	if err != nil {
		logger.Error(err, "failed to create or update deployment")
		js.Status.State = "Error"
		js.Status.Message = "Error: json-server unexpected failure"
		_ = r.Status().Update(ctx, &js)
		return ctrl.Result{}, err
	}

	// Service
	desiredSvc := corev1.Service{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace},
		Spec:       corev1.ServiceSpec{Selector: labels, Ports: []corev1.ServicePort{{Port: 3000, TargetPort: intstr.FromInt(3000), Protocol: corev1.ProtocolTCP}}},
	}

	// Service (CreateOrUpdate)
	svc := &corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace}}
	err = retry.RetryOnConflict(retry.DefaultRetry, func() error {
		if err := r.Get(ctx, client.ObjectKeyFromObject(svc), svc); err != nil {
			if apierrors.IsNotFound(err) {
				svc = &corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace}}
			} else {
				return err
			}
		}
		_, err := controllerutil.CreateOrUpdate(ctx, r.Client, svc, func() error {
			svc.Spec = desiredSvc.Spec
			return controllerutil.SetControllerReference(&js, svc, r.Scheme)
		})
		return err
	})
	if err != nil {
		logger.Error(err, "failed to create or update service")
		js.Status.State = "Error"
		js.Status.Message = "Error: json-server unexpected failure"
		_ = r.Status().Update(ctx, &js)
		return ctrl.Result{}, err
	}

	// Refresh deployment status for scale subresource
	if err := r.Get(ctx, req.NamespacedName, deploy); err != nil {
		logger.Error(err, "failed to fetch deployment status")
		return ctrl.Result{}, err
	}

	selector := metav1.FormatLabelSelector(&metav1.LabelSelector{MatchLabels: labels})

	// Success
	js.Status.State = "Synced"
	js.Status.Message = "Synced succesfully!"
	js.Status.Replicas = deploy.Status.Replicas
	js.Status.Selector = selector
	if err := r.Status().Update(ctx, &js); err != nil {
		logger.Error(err, "failed to update status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *JsonServerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&examplev1.JsonServer{}).
		Named("jsonserver").
		Complete(r)
}

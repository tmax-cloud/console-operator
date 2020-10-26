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
	"reflect"

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/source"

	hypercloudv1 "github.com/tmax-cloud/console-operator/api/v1"
)

// ConsoleReconciler reconciles a Console object
type ConsoleReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=hypercloud.tmaxcloud.com,resources=consoles,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=hypercloud.tmaxcloud.com,resources=consoles/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=*,resources=*,verbs=*

func (r *ConsoleReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("console", req.NamespacedName)

	// your logic here
	log.Info("Reconciling Console")

	// Fetch the Console instance
	console := &hypercloudv1.Console{}
	err := r.Get(ctx, req.NamespacedName, console)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			log.Info("Console resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		log.Error(err, "Failed to get Console")
		return ctrl.Result{}, err
	}
	r.Create(ctx, r.serviceAccount(console))
	r.Create(ctx, r.clusterRole(console))
	r.Create(ctx, r.clusterRoleBinding(console))
	err = r.Create(ctx, r.jobForTls(console))
	if err != nil && errors.IsAlreadyExists(err) {
		log.Info(err.Error())
		log.Info("Failed to create job")
	}
	// Define a new deployment
	found := &appsv1.Deployment{}
	err = r.Get(ctx, types.NamespacedName{Name: console.Name, Namespace: console.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		// Define a new deployment
		dep := r.deploymentForConsole(console)
		log.Info("Creating a new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
		err = r.Create(context.TODO(), dep)
		if err != nil && errors.IsAlreadyExists(err) {
			return ctrl.Result{Requeue: true}, nil
		} else if err != nil {
			log.Error(err, "Failed to create new Deployment.", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
			return ctrl.Result{}, err
		}
		// Deployment created successfully - return and requeue
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		log.Error(err, "Failed to get Deployment")
	}

	// Checking Deployment Status
	depStat := r.deploymentForConsole(console)
	if !reflect.DeepEqual(depStat.Spec, found.Spec) {
		found.Spec = depStat.Spec
		err = r.Update(ctx, found)
		if err != nil {
			log.Error(err, "Failed to update Deployment.", "Deployment.Namespace", found.Namespace, "Deployment.Name", found.Name)
			return ctrl.Result{}, err
		}
	}

	foundService := &corev1.Service{}
	err = r.Get(ctx, types.NamespacedName{Name: console.Name, Namespace: console.Namespace}, foundService)
	if err != nil && errors.IsNotFound(err) {
		svc := r.serviceForConsole(console)
		log.Info("Creating a new Service", "Service.Namespace", svc.Namespace, "Service,Name", svc.Name)
		err = r.Create(ctx, svc)
		if err != nil && errors.IsAlreadyExists(err) {
			return ctrl.Result{Requeue: true}, nil
		} else if err != nil {
			log.Error(err, "Failed to craete new service", "Service.Namespace", svc.Namespace, "Serivce.Name", svc.Name)
			return ctrl.Result{}, err
		}
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		log.Error(err, "Failed to get Service")
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

func (r *ConsoleReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&hypercloudv1.Console{}).
		Watches(
			&source.Kind{Type: &hypercloudv1.Console{}},
			&handler.EnqueueRequestForObject{},
		).
		Watches(
			&source.Kind{Type: &appsv1.Deployment{}},
			&handler.EnqueueRequestForOwner{
				IsController: true,
				OwnerType:    &hypercloudv1.Console{},
			},
		).
		Watches(
			&source.Kind{Type: &corev1.Service{}},
			&handler.EnqueueRequestForOwner{
				IsController: true,
				OwnerType:    &hypercloudv1.Console{},
			},
		).
		Complete(r)
}

func (r *ConsoleReconciler) jobForTls(console *hypercloudv1.Console) *batchv1.Job {
	bo := bool(true)
	in := int64(2000)
	return &batchv1.Job{
		TypeMeta: metav1.TypeMeta{APIVersion: batchv1.SchemeGroupVersion.Version, Kind: "Job"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      console.Name,
			Namespace: console.Namespace,
		},
		Spec: batchv1.JobSpec{
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name: console.Name,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						corev1.Container{
							Name:            "create",
							Image:           "docker.io/jettech/kube-webhook-certgen:v1.3.0",
							ImagePullPolicy: corev1.PullIfNotPresent,
							Args: []string{
								"create",
								"--host=console,console.$(POD_NAMESPACE).svc",
								"--namespace=$(POD_NAMESPACE)",
								"--secret-name=console-https-secret",
							},
							Env: []corev1.EnvVar{
								corev1.EnvVar{
									Name: "POD_NAMESPACE",
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											FieldPath: "metadata.namespace",
										},
									},
								},
							},
						},
					},
					RestartPolicy:      corev1.RestartPolicyOnFailure,
					ServiceAccountName: console.Name,
					SecurityContext: &corev1.PodSecurityContext{
						RunAsNonRoot: &bo,
						RunAsUser:    &in,
					},
				},
			},
		},
	}
}

func (r *ConsoleReconciler) serviceAccount(console *hypercloudv1.Console) *corev1.ServiceAccount {
	return &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      console.Name,
			Namespace: console.Namespace,
		},
	}
}

func (r *ConsoleReconciler) clusterRole(console *hypercloudv1.Console) *rbacv1.ClusterRole {
	return &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name:      console.Name,
			Namespace: console.Namespace,
		},
		Rules: []rbacv1.PolicyRule{
			rbacv1.PolicyRule{
				APIGroups: []string{"*"},
				Resources: []string{"*"},
				Verbs:     []string{"*"},
			},
		},
	}
}

func (r *ConsoleReconciler) clusterRoleBinding(console *hypercloudv1.Console) *rbacv1.ClusterRoleBinding {
	return &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: console.Name,
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      console.Name,
				Namespace: console.Namespace,
			},
		},
		RoleRef: rbacv1.RoleRef{
			Kind:     "ClusterRole",
			Name:     console.Name,
			APIGroup: "rbac.authorization.k8s.io",
		},
	}
}

func (r *ConsoleReconciler) serviceForConsole(console *hypercloudv1.Console) *corev1.Service {
	p := make([]corev1.ServicePort, 0)
	servicePort := corev1.ServicePort{
		Name:       "tcp-port",
		Port:       433,
		TargetPort: intstr.FromInt(6433),
	}
	p = append(p, servicePort)
	consoleSvc := &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      console.Name,
			Namespace: console.Namespace,
			Labels:    map[string]string{"app": "console"},
		},
		Spec: corev1.ServiceSpec{
			Ports:    p,
			Type:     console.Spec.App.ServiceType,
			Selector: map[string]string{"app": "console"},
		},
	}
	return consoleSvc
}

func (r *ConsoleReconciler) deploymentForConsole(console *hypercloudv1.Console) *appsv1.Deployment {
	cnts := make([]corev1.Container, 0)

	cnt := corev1.Container{
		Name:            "console",
		Image:           console.Spec.App.Repository + ":" + console.Spec.App.Tag,
		ImagePullPolicy: corev1.PullIfNotPresent,
		Command: []string{
			"/opt/bridge/bin/bridge",
			"--public-dir=/opt/bridge/static",
			"--listen=https://0.0.0.0:6443",
			"--base-address=https://0.0.0.0:6443",
			"--tls-cert-file=/var/https-cert/tls.crt",
			"--tls-key-file=/var/https-cert/tls.key",
			"--user-auth=disabled",
		},
		Ports: []corev1.ContainerPort{
			corev1.ContainerPort{
				ContainerPort: 6443,
				Protocol:      "TCP",
			},
		},
		VolumeMounts: []corev1.VolumeMount{
			corev1.VolumeMount{
				MountPath: "/var/https-cert",
				Name:      "https-cert",
				ReadOnly:  true,
			},
		},
		TerminationMessagePath:   "/dev/termination-log",
		TerminationMessagePolicy: "File",
	}
	cnts = append(cnts, cnt)
	podTempSpec := corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{"app": "console"},
		},
		Spec: corev1.PodSpec{
			ServiceAccountName: console.Name,
			Containers:         cnts,
			Volumes: []corev1.Volume{
				corev1.Volume{
					Name: "https-cert",
					VolumeSource: corev1.VolumeSource{
						Secret: &corev1.SecretVolumeSource{
							SecretName: "console-https-secret",
						},
					},
				},
			},
		},
	}
	dep := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      console.Name,
			Namespace: console.Namespace,
			Labels:    map[string]string{"app": "console"},
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": "console"},
			},
			Replicas: &console.Spec.App.Replicas,
			Template: podTempSpec,
		},
	}
	return dep
}

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
	"os"

	"github.com/go-logr/logr"
	hypercloudv1 "github.com/tmax-cloud/console-operator/api/v1"
	"gopkg.in/yaml.v2"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ConsoleReconciler reconciles a Console object
type ConsoleReconciler struct {
	client.Client
	Log           logr.Logger
	Scheme        *runtime.Scheme
	DynamicConfig string
}

// +kubebuilder:rbac:groups=hypercloud.tmaxcloud.com,resources=consoles,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=hypercloud.tmaxcloud.com,resources=consoles/status,verbs=get;update;patch
// +kubebuilder:rbac:groups="",resources=*,verbs=*

func (r *ConsoleReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {

	ctx := context.Background()
	log := r.Log.WithValues("console", req.NamespacedName)

	log.Info("Reconciling Console")

	var console hypercloudv1.Console

	if err := r.Get(ctx, req.NamespacedName, &console); err != nil {
		log.Info("Unable to fetch Console", "Error", err)

		// Delete Cr,Crb
		var crb rbacv1.ClusterRoleBinding
		if err := r.Get(ctx, client.ObjectKey{Name: req.Name}, &crb); err == nil {
			return r.removeCrb(ctx, &crb, log)
		}
		var cr rbacv1.ClusterRole
		if err := r.Get(ctx, client.ObjectKey{Name: req.Name}, &cr); err == nil {
			return r.removeCr(ctx, &cr, log)
		}

		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	log.Info(console.Namespace + "/" + console.Name)

	yamlConsole := console.DeepCopy()

	config := yamlConsole.Spec.Configuration
	yamlFile, _ := yaml.Marshal(config)
	fy, _ := os.Create(r.DynamicConfig)
	_, err := fy.Write(yamlFile)
	if err != nil {
		return ctrl.Result{Requeue: false}, err
	}
	err = yaml.Unmarshal(yamlFile, &console)
	if err != nil {
		return ctrl.Result{}, err
	}
	// fmt.Printf("YAML: \n %v \n", console)

	console.Status.Number = 0
	console.Status.Routers = ""
	for name, router := range config.Routers {
		console.Status.Number = console.Status.Number + 1
		temp := "[" + name + " : " + router.Server + " " + router.Path + "]  "
		console.Status.Routers = console.Status.Routers + temp
	}
	if err := r.Status().Update(ctx, &console); err != nil {
		log.Error(err, "unable to update CronJob status")
		return ctrl.Result{}, err
	}
	// // var i int
	// i := 0
	// for name, router := range config.Routers {
	// 	i = i + 1
	// 	_ = router
	// 	console.Status.Routers = console.Status.Routers + " , " + name
	// 	console.Status.Number = i
	// }
	// err := r.Status().Update(ctx, &console)
	// if err != nil {
	// 	return reconcile.Result{}, err
	// }
	// jsonFile, _ := json.Marshal(console.DeepCopyObject())
	// fj, _ := os.Create("/home/jinsoo/console-json.json")
	// fj.Write(jsonFile)
	// json.Unmarshal(jsonFile, &console)
	// fmt.Printf("JSON: \n %v \n", console)

	// log.Info("wrote  file : %v ", n1)
	// viper.AddConfigPath("/root/")
	// viper.SafeWriteConfig()

	// sa, err := r.desiredServiceAccount(console)
	// if err != nil {
	// 	return ctrl.Result{}, err
	// }

	// cr, err := r.desiredClusterRole(console)
	// if err != nil {
	// 	return ctrl.Result{}, err
	// }

	// crb, err := r.desiredClusterRoleBinding(console)
	// if err != nil {
	// 	return ctrl.Result{}, err
	// }

	// job, err := r.desiredJob(console)
	// if err != nil {
	// 	return ctrl.Result{}, err
	// }

	// applyOpts := []client.PatchOption{client.ForceOwnership, client.FieldOwner("console-controller")}

	// if err := r.Patch(ctx, &sa, client.Apply, applyOpts...); err != nil {
	// 	log.Error(err, "error is occur when sa")
	// 	sa := &corev1.ServiceAccount{}
	// 	r.Get(ctx, req.NamespacedName, sa)
	// 	return r.removeSa(ctx, sa, log)
	// }

	// // if err := r.Create(ctx, &cr, client.Apply, applyOpts...); err != nil {
	// if err := r.Update(ctx, &cr); err != nil {
	// 	log.Error(err, "error is occure when cr")
	// 	cr := &rbacv1.ClusterRole{}
	// 	r.Get(ctx, client.ObjectKey{Name: req.Name}, cr)
	// 	return r.removeCr(ctx, cr, log)
	// }

	// // if err := r.Patch(ctx, &crb, client.Apply, applyOpts...); err != nil {
	// if err := r.Update(ctx, &crb); err != nil {
	// 	log.Error(err, "error is occure when crb")
	// 	crb := &rbacv1.ClusterRoleBinding{}
	// 	r.Get(ctx, client.ObjectKey{Name: req.Name}, crb)
	// 	return r.removeCrb(ctx, crb, log)
	// }

	// err = r.Patch(ctx, &job, client.Apply, applyOpts...)
	// if err != nil {
	// 	log.Error(err, "error is occur when job")
	// 	job := &batchv1.Job{}
	// 	r.Get(ctx, req.NamespacedName, job)
	// 	return r.removeJob(ctx, job, log)
	// }

	// var serviceAddr string
	// checkSvc := &corev1.Service{}
	// r.Get(ctx, req.NamespacedName, checkSvc)
	// // r.getUrl(ctx, *checkSvc, lolg)
	// // console.Status.LeaderService = string(checkSvc.Spec.Type)
	// if checkSvc.Spec.Type == corev1.ServiceTypeLoadBalancer {
	// 	console.Status.TYPE = string(corev1.ServiceTypeLoadBalancer)
	// 	if checkSvc.Status.LoadBalancer.Ingress == nil || len(checkSvc.Status.LoadBalancer.Ingress) == 0 {
	// 		serviceAddr = "Undefined"
	// 	}
	// 	serviceAddr = "https://" + checkSvc.Status.LoadBalancer.Ingress[0].IP
	// } else {
	// 	console.Status.TYPE = string(corev1.ServiceTypeNodePort)
	// 	if checkSvc.Spec.Ports == nil || len(checkSvc.Spec.Ports) == 0 {
	// 		serviceAddr = "Undefined"
	// 	}
	// 	serviceAddr = "https://<NodeIP>:" + strconv.Itoa(int(checkSvc.Spec.Ports[0].NodePort))
	// }
	// console.Status.URL = serviceAddr
	// log.Info(console.Status.TYPE)
	// log.Info(console.Status.STATUS)

	// checkDepl := &appsv1.Deployment{}
	// r.Get(ctx, req.NamespacedName, checkDepl)
	// if checkDepl.Status.Conditions != nil && len(checkDepl.Status.Conditions) != 0 {
	// 	if checkDepl.Status.Conditions[0].Type == appsv1.DeploymentAvailable {
	// 		console.Status.STATUS = "Ready"
	// 	} else {
	// 		console.Status.STATUS = "Failed"
	// 	}
	// } else {
	// 	console.Status.STATUS = "Undefined"
	// }
	// log.Info(console.Status.STATUS)

	// r.Status().Update(ctx, &console)

	log.Info("reconciled console")
	return ctrl.Result{}, nil
}

func (r *ConsoleReconciler) SetupWithManager(mgr ctrl.Manager) error {

	return ctrl.NewControllerManagedBy(mgr).
		For(&hypercloudv1.Console{}).
		// Owns(&v1.ConfigMap{}).
		Complete(r)
}

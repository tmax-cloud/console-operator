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
	var err error
	ctx := context.Background()
	log := r.Log.WithValues("console", req.NamespacedName)

	log.Info("Reconciling Console")

	var console hypercloudv1.Console

	if err := r.Get(ctx, req.NamespacedName, &console); err != nil {
		log.Info("Unable to get Console", "Error", err)

		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	log.Info(console.Namespace + "/" + console.Name)

	yamlConsole := console

	config := yamlConsole.Spec.Configuration
	yamlFile, _ := yaml.Marshal(config)
	fy, _ := os.Create(r.DynamicConfig)
	_, err = fy.Write(yamlFile)
	if err != nil {
		return ctrl.Result{Requeue: false}, err
	}

	console.Status.Number = 0
	console.Status.Routers = ""
	for name, router := range config.Routers {
		console.Status.Number = console.Status.Number + 1
		temp := "[" + name + " : " + router.Server + " " + router.Path + "]  "
		console.Status.Routers = console.Status.Routers + temp
	}
	if err := r.Status().Update(ctx, &console); err != nil {
		log.Info("unable to update Console status", "Error :", err)
		return ctrl.Result{}, err
	}

	log.Info("reconciled console")
	return ctrl.Result{}, nil
}

func (r *ConsoleReconciler) SetupWithManager(mgr ctrl.Manager) error {

	return ctrl.NewControllerManagedBy(mgr).
		For(&hypercloudv1.Console{}).
		Complete(r)
}

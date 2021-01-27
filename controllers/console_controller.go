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
	"os"

	"github.com/go-logr/logr"
	hypercloudv1 "github.com/tmax-cloud/console-operator/api/v1"
	"gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	configFileName = "dynamic-config.yaml"
)

// ConsoleReconciler reconciles a Console object
type ConsoleReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
	PWD    string
	Config map[string]*hypercloudv1.Configuration
	// NameSpace string
}

// +kubebuilder:rbac:groups=hypercloud.tmaxcloud.com,resources=consoles,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=hypercloud.tmaxcloud.com,resources=consoles/status,verbs=get;update;patch
// +kubebuilder:rbac:groups="",resources=*,verbs=*

func (r *ConsoleReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	var console hypercloudv1.Console
	var err error

	key := req.Name + "@" + req.Namespace
	fileName := req.Name + "@" + req.Namespace + ".yaml"
	pwd := r.PWD

	ctx := context.Background()
	log := r.Log.WithValues("console", req.NamespacedName)

	log.Info("Reconciling Console")

	if err := r.Get(ctx, req.NamespacedName, &console); err != nil {
		log.Info("Unable to get Console", "Error", err)
		log.Info(fmt.Sprintf("Delete config file. fileName : %v", fileName))

		os.Remove(pwd + fileName)
		delete(r.Config, key)
		err = r.createProxyFile(pwd + configFileName)
		if err != nil {
			return ctrl.Result{Requeue: false}, err
		}

		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	log.Info(console.Namespace + "/" + console.Name)

	// Create ConfigFile correspoding to each console CR
	config := console.Spec.Configuration.DeepCopy()
	err = r.createConfigFile(pwd+fileName, config)
	if err != nil {
		return ctrl.Result{Requeue: false}, err
	}
	// END

	// // Create a ConfigFile which adding ALL console cr
	// r.Config[key] = console.Spec.Configuration.DeepCopy()
	// err = r.createConfigFile(pwd+"all.yaml", r.Config)
	// if err != nil {
	// 	return ctrl.Result{Requeue: false}, err
	// }
	// // END

	// Create a ConfigFile without console Name, only show router config
	r.Config[key] = console.Spec.Configuration.DeepCopy()
	err = r.createProxyFile(pwd + configFileName)
	if err != nil {
		return ctrl.Result{Requeue: false}, err
	}
	// END

	//Update console status
	console.Status.Number = 0
	console.Status.Routers = ""
	for name, router := range config.Routers {
		console.Status.Number = console.Status.Number + 1
		temp := "[" + name + " : " + router.Server + " " + router.Path + "]  "
		console.Status.Routers = console.Status.Routers + temp
	}
	if err := r.Status().Update(ctx, &console); err != nil {
		log.Info("unable to update Console status")
		return ctrl.Result{}, err
	}

	count := len(r.Config)
	var showName []string
	for name := range r.Config {
		showName = append(showName, name)
	}
	log.Info(fmt.Sprintf("Console Controller has the %v files -> %v", count, showName))
	// END

	// // [Deplicated]
	// // Create a configMap named dynamic-config where has a config file named "dynamic-config.yaml".
	// currentConfigMap := &corev1.ConfigMap{}
	// newConfigMap, err := r.createConfigMap()
	// if err != nil {
	// 	return ctrl.Result{Requeue: false}, err
	// }
	// err = r.Get(ctx, types.NamespacedName{Name: "dynamic-config", Namespace: r.NameSpace}, currentConfigMap)
	// if err == nil {
	// 	if !reflect.DeepEqual(currentConfigMap.BinaryData, newConfigMap.BinaryData) {
	// 		log.Info(fmt.Sprintf("ConfigMap will be changed: %v", "dynamic-config"))
	// 		if err = r.Delete(ctx, currentConfigMap); err != nil {
	// 			log.Info("Failed to delete old configmap", "error", err.Error())
	// 		}
	// 		if r.Update(ctx, newConfigMap) != nil {
	// 			return ctrl.Result{}, err
	// 		}
	// 	}
	// } else {
	// 	if errors.IsNotFound(err) {
	// 		if r.Create(ctx, newConfigMap) != nil {
	// 			log.Error(err, err.Error())
	// 			return ctrl.Result{}, err
	// 		}
	// 	} else {
	// 		log.Error(err, err.Error())
	// 		return ctrl.Result{}, err
	// 	}
	// }
	// // configmap END

	log.Info("reconciled console")
	return ctrl.Result{}, nil
}

func (r *ConsoleReconciler) SetupWithManager(mgr ctrl.Manager) error {

	return ctrl.NewControllerManagedBy(mgr).
		For(&hypercloudv1.Console{}).
		Owns(&corev1.ConfigMap{}).
		Complete(r)
}

func (r *ConsoleReconciler) createConfigFile(fileName string, config interface{}) error {
	file, err := os.Create(fileName)
	if err != nil {
		return err
	}
	yamlConfig, err := yaml.Marshal(config)
	if err != nil {
		return err
	}
	_, err = file.Write(yamlConfig)
	if err != nil {
		return err
	}
	return nil
}

func (r *ConsoleReconciler) createProxyFile(fileName string) error {
	temp := hypercloudv1.Configuration{
		Routers: make(map[string]*hypercloudv1.Router),
	}
	for _, conf := range r.Config {
		for name, router := range conf.Routers {
			temp.Routers[name] = router
		}
	}
	err := r.createConfigFile(fileName, temp)
	if err != nil {
		return err
	}
	return nil
}

// // [DEPLICATED] When we use  configMap, the updating time is too long (may be 1min )
// func (r *ConsoleReconciler) createConfigMap() (*corev1.ConfigMap, error) {
// 	temp := hypercloudv1.Configuration{
// 		Routers: make(map[string]*hypercloudv1.Router),
// 	}
// 	for _, conf := range r.Config {
// 		for name, router := range conf.Routers {
// 			temp.Routers[name] = router
// 		}
// 	}
// 	yamlConfig, err := yaml.Marshal(temp)
// 	if err != nil {
// 		return nil, err
// 	}

// 	proxyConfigMap := &corev1.ConfigMap{
// 		TypeMeta: metav1.TypeMeta{
// 			Kind:       "ConfigMap",
// 			APIVersion: "v1",
// 		},
// 		ObjectMeta: metav1.ObjectMeta{
// 			Name:      "dynamic-config",
// 			Namespace: r.NameSpace,
// 		},
// 		BinaryData: map[string][]byte{
// 			"dynamic-config.yaml": yamlConfig,
// 		},
// 	}
// 	return proxyConfigMap, nil
// }

package console

import (
	"context"
	"reflect"

	consolev1alpha1 "github.com/tmax-cloud/console-operator/pkg/apis/console/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_console")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new Console Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileConsole{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("console-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource Console
	err = c.Watch(&source.Kind{Type: &consolev1alpha1.Console{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner Console
	// err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
	// 	IsController: true,
	// 	OwnerType:    &consolev1alpha1.Console{},
	// })
	// if err != nil {
	// 	return err
	// }

	return nil
}

// blank assignment to verify that ReconcileConsole implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileConsole{}

// ReconcileConsole reconciles a Console object
type ReconcileConsole struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a Console object and makes changes based on the state read
// and what is in the Console.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileConsole) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling Console")

	// Fetch the Console instance
	instance := &consolev1alpha1.Console{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		reqLogger.Error(err, "Failed to get console object")
		return reconcile.Result{}, err
	}
	err = r.Console(instance)
	if err != nil {
		reqLogger.Error(err, "Failed to create/update bookstore resources")
		return reconcile.Result{}, err
	}
	_ = r.client.Status().Update(context.TODO(), instance)
	return reconcile.Result{}, nil

	// Define a new Pod object
	// pod := newPodForCR(instance)

	// // Set Console instance as the owner and controller
	// if err := controllerutil.SetControllerReference(instance, pod, r.scheme); err != nil {
	// 	return reconcile.Result{}, err
	// }

	// Check if this Pod already exists
	// found := &corev1.Pod{}
	// err = r.client.Get(context.TODO(), types.NamespacedName{Name: pod.Name, Namespace: pod.Namespace}, found)
	// if err != nil && errors.IsNotFound(err) {
	// 	reqLogger.Info("Creating a new Pod", "Pod.Namespace", pod.Namespace, "Pod.Name", pod.Name)
	// 	err = r.client.Create(context.TODO(), pod)
	// 	if err != nil {
	// 		return reconcile.Result{}, err
	// 	}

	// 	// Pod created successfully - don't requeue
	// 	return reconcile.Result{}, nil
	// } else if err != nil {
	// 	return reconcile.Result{}, err
	// }

	// // Pod already exists - don't requeue
	// reqLogger.Info("Skip reconcile: Pod already exists", "Pod.Namespace", found.Namespace, "Pod.Name", found.Name)
	// return reconcile.Result{}, nil
}

func (r *ReconcileConsole) Console(console *consolev1alpha1.Console) error {
	reqLogger := log.WithValues("Namespace", console.Namespace)

	// Service
	consoleSvc := getConsoleSvc(console)
	csvc := &corev1.Service{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: "console-svc", Namespace: console.Namespace}, csvc)
	if err != nil {
		if errors.IsNotFound(err) {
			controllerutil.SetControllerReference(console, consoleSvc, r.scheme)
			err = r.client.Create(context.TODO(), consoleSvc)
			if err != nil {
				return err
			}
		} else {
			reqLogger.Info("failed to get console service")
			return err
		}
	} else if !reflect.DeepEqual(consoleSvc.Spec, csvc.Spec) {
		consoleSvc.ObjectMeta = csvc.ObjectMeta
		consoleSvc.Spec.ClusterIP = csvc.Spec.ClusterIP
		controllerutil.SetControllerReference(console, consoleSvc, r.scheme)
		err = r.client.Update(context.TODO(), consoleSvc)
		if err != nil {
			return err
		}
		reqLogger.Info("console service updated")
	}
	// Deployment
	consoleDep := getConsoleDeploy(console)
	cdep := &appsv1.Deployment{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: "console", Namespace: console.Namespace}, cdep)
	if err != nil {
		if errors.IsNotFound(err) {
			controllerutil.SetControllerReference(console, consoleDep, r.scheme)
			err = r.client.Create(context.TODO(), consoleDep)
			if err != nil {
				return err
			}
		} else {
			reqLogger.Info("failed to get console deployment")
			return err
		}
	} else if !reflect.DeepEqual(consoleDep.Spec, cdep.Spec) {
		consoleDep.ObjectMeta = cdep.ObjectMeta
		controllerutil.SetControllerReference(console, consoleDep, r.scheme)
		err = r.client.Update(context.TODO(), consoleDep)
		if err != nil {
			return err
		}
		reqLogger.Info("console deployment updated")
	}
	r.client.Status().Update(context.TODO(), console)
	return nil
}

func getConsoleSvc(console *consolev1alpha1.Console) *corev1.Service {
	p := make([]corev1.ServicePort, 0)
	servicePort := corev1.ServicePort{
		Name:       "tcp-port",
		Port:       console.Spec.ConsoleApp.Port,
		TargetPort: intstr.FromInt(console.Spec.ConsoleApp.TargetPort),
	}
	p = append(p, servicePort)
	consoleSvc := &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "console-svc",
			Namespace: console.Namespace,
			Labels:    map[string]string{"app": "console"},
		},
		Spec: corev1.ServiceSpec{
			Ports:    p,
			Type:     console.Spec.ConsoleApp.ServiceType,
			Selector: map[string]string{"app": "console"},
		},
	}
	return consoleSvc
}

func getConsoleDeploy(console *consolev1alpha1.Console) *appsv1.Deployment {

	cnts := make([]corev1.Container, 0)
	cnt := corev1.Container{
		Name:            "console",
		Image:           console.Spec.ConsoleApp.Repository + ":" + console.Spec.ConsoleApp.Tag,
		ImagePullPolicy: console.Spec.ConsoleApp.ImagePullPolicy,
	}
	cnts = append(cnts, cnt)
	podTempSpec := corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{"app": "console"},
		},
		Spec: corev1.PodSpec{
			Containers: cnts,
		},
	}
	dep := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "console",
			Namespace: console.Namespace,
			Labels:    map[string]string{"app": "console"},
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": "console"},
			},
			Replicas: &console.Spec.ConsoleApp.Replicas,
			Template: podTempSpec,
		},
	}
	return dep
}

// newPodForCR returns a busybox pod with the same name/namespace as the cr
// func newPodForCR(cr *consolev1alpha1.Console) *corev1.Pod {
// 	labels := map[string]string{
// 		"app": cr.Name,
// 	}
// 	return &corev1.Pod{
// 		ObjectMeta: metav1.ObjectMeta{
// 			Name:      cr.Name + "-pod",
// 			Namespace: cr.Namespace,
// 			Labels:    labels,
// 		},
// 		Spec: corev1.PodSpec{
// 			Containers: []corev1.Container{
// 				{
// 					Name:    "busybox",
// 					Image:   "busybox",
// 					Command: []string{"sleep", "3600"},
// 				},
// 			},
// 		},
// 	}
// }

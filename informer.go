package main

import (
	"fmt"

	"k8s.io/client-go/informers"
	coreinformers "k8s.io/client-go/informers/core/v1"
	klog "k8s.io/klog/v2"

	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/cache"
)

// ActivePassiveController logs the name and namespace of pods that are added,
// deleted, or updated
type ActivePassiveController struct {
	informerFactory informers.SharedInformerFactory
	podInformer     coreinformers.PodInformer
}

// Run starts shared informers and waits for the shared informer cache to
// synchronize.
func (c *ActivePassiveController) Run(stopCh chan struct{}) error {
	// Starts all the shared informers that have been created by the factory so
	// far.
	c.informerFactory.Start(stopCh)
	// wait for the initial synchronization of the local cache.
	if !cache.WaitForCacheSync(stopCh, c.podInformer.Informer().HasSynced) {
		return fmt.Errorf("failed to sync")
	}

	return nil
}

func (c *ActivePassiveController) podAdd(obj interface{}) {
	pod := obj.(*v1.Pod)
	klog.Infof("POD CREATED: %s/%s", pod.Namespace, pod.Name)
}

func (c *ActivePassiveController) podUpdate(old, new interface{}) {
	oldPod := old.(*v1.Pod)
	newPod := new.(*v1.Pod)

	if oldPod.Status.Phase == v1.PodRunning && newPod.Status.Phase != v1.PodRunning {
		// business logic - perform swicth over
		fmt.Println("the active pod will now become stand-by!")
	}

	klog.Infof(
		"POD UPDATED. %s/%s %s",
		oldPod.Namespace, oldPod.Name, newPod.Status.Phase,
	)
}

func (c *ActivePassiveController) podDelete(obj interface{}) {
	pod := obj.(*v1.Pod)
	klog.Infof("POD DELETED: %s/%s", pod.Namespace, pod.Name)
}

// NewActivePassiveController creates a ActivePassiveController
func NewActivePassiveController(informerFactory informers.SharedInformerFactory) (*ActivePassiveController, error) {
	podInformer := informerFactory.Core().V1().Pods()

	c := &ActivePassiveController{
		informerFactory: informerFactory,
		podInformer:     podInformer,
	}
	_, err := podInformer.Informer().AddEventHandler(
		// Your custom resource event handlers.
		cache.ResourceEventHandlerFuncs{

			// Called on creation
			// AddFunc: c.podAdd,

			// Called on resource update and every resyncPeriod on existing resources.
			UpdateFunc: c.podUpdate,

			// Called on resource deletion.
			// DeleteFunc: c.podDelete,
		},
	)
	if err != nil {
		return nil, err
	}

	return c, nil
}

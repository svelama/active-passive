package main

import (
	"flag"
	"os"
	"path/filepath"
	"time"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	klog "k8s.io/klog/v2"

	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/component-base/logs"
)

var kubeconfig string

func init() {
	flag.StringVar(&kubeconfig, "kubeconfig", filepath.Join(os.Getenv("HOME"), ".kube", "config"), "absolute path to the kubeconfig file")
}

func main() {

	flag.Parse()

	logs.InitLogs()
	defer logs.FlushLogs()

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		klog.Fatal(err)
	}

	// To watch events that are specific to a namespace
	// nsOptions := informers.WithNamespace("test")

	// To watch specific pod based on label
	labelOptions := informers.WithTweakListOptions(func(lo *v1.ListOptions) {
		lo.LabelSelector = "mode=active"
	})

	factory := informers.NewSharedInformerFactoryWithOptions(clientset, time.Hour*24, labelOptions)
	controller, err := NewActivePassiveController(factory)
	if err != nil {
		klog.Fatal(err)
	}

	stop := make(chan struct{})
	defer close(stop)
	err = controller.Run(stop)
	if err != nil {
		klog.Fatal(err)
	}
	select {}
}

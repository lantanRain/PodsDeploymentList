package main

import (
	"context"
	"errors"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	typev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/clientcmd"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func main() {

	ctx := context.Background()

	kubeconfig := filepath.Join(
		os.Getenv("HOME"), ".kube", "config",
	)
	namespace := "openfx-fn"
	k8sClient, err := getClient(kubeconfig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	svc, err := getServiceForDeployment(ctx, "APP_NAME", namespace, k8sClient)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(2)
	}

	pods, err := getPodsForSvc(ctx, svc, namespace, k8sClient)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(2)
	}
	log.Printf(pods.String())

}

func getClient(configLocation string) (typev1.CoreV1Interface, error) {
	kubeconfig := filepath.Clean(configLocation)
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		log.Fatal(err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return clientset.CoreV1(), nil
}

func getServiceForDeployment(ctx context.Context, deployment string, namespace string, k8sClient typev1.CoreV1Interface) (*corev1.Service, error) {
	listOptions := metav1.ListOptions{}
	svcs, err := k8sClient.Services(namespace).List(ctx, listOptions)
	if err != nil {
		log.Fatal(err)
	}
	for _, svc := range svcs.Items {
		if strings.Contains(svc.Name, deployment) {
			fmt.Fprintf(os.Stdout, "service name: %v\n", svc.Name)
			return &svc, nil
		}
	}
	return nil, errors.New("cannot find service for deployment")
}

func getPodsForSvc(ctx context.Context, svc *corev1.Service, namespace string, k8sClient typev1.CoreV1Interface) (*corev1.PodList, error) {
	set := labels.Set(svc.Spec.Selector)
	listOptions := metav1.ListOptions{LabelSelector: set.AsSelector().String()}
	pods, err := k8sClient.Pods(namespace).List(ctx, listOptions)
	for _, pod := range pods.Items {
		fmt.Fprintf(os.Stdout, "pod name: %v\n", pod.Name)
	}
	return pods, err
}

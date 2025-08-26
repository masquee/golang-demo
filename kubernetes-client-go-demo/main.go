package main

import (
	"context"
	"fmt"
	"log"
	"path/filepath"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func main() {
	// Build kubeconfig path
	var kubeconfig string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = filepath.Join(home, ".kube", "config")
	}

	// Create Kubernetes client configuration
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		log.Fatalf("Failed to build kubeconfig: %v", err)
	}

	// Create Kubernetes clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("Failed to create clientset: %v", err)
	}

	fmt.Println("=== Kubernetes Client-Go Demo ===")
	fmt.Println("Connected to Kubernetes cluster successfully!")

	// Demonstrate listing pods
	fmt.Println("\n1. Listing Pods:")
	listPods(clientset)

	// Demonstrate watching pods
	fmt.Println("\n2. Watching Pods (for 30 seconds):")
	watchPods(clientset)
}

// listPods demonstrates listing all pods in all namespaces
func listPods(clientset *kubernetes.Clientset) {
	ctx := context.Background()

	// List pods in all namespaces
	pods, err := clientset.CoreV1().Pods("").List(ctx, metav1.ListOptions{})
	if err != nil {
		log.Printf("Failed to list pods: %v", err)
		return
	}

	fmt.Printf("Found %d pods:\n", len(pods.Items))
	for _, pod := range pods.Items {
		fmt.Printf("  - %s/%s (Status: %s, Created: %s)\n",
			pod.Namespace,
			pod.Name,
			pod.Status.Phase,
			pod.CreationTimestamp.Format("2006-01-02 15:04:05"),
		)
	}
}

// watchPods demonstrates watching pod events
func watchPods(clientset *kubernetes.Clientset) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Create a watcher for pods in all namespaces
	watcher, err := clientset.CoreV1().Pods("").Watch(ctx, metav1.ListOptions{})
	if err != nil {
		log.Printf("Failed to create pod watcher: %v", err)
		return
	}
	defer watcher.Stop()

	fmt.Println("Watching for pod events... (Press Ctrl+C to stop)")

	for {
		select {
		case event, ok := <-watcher.ResultChan():
			if !ok {
				fmt.Println("Watcher channel closed")
				return
			}

			pod, ok := event.Object.(*corev1.Pod)
			if !ok {
				fmt.Printf("Unexpected object type: %T\n", event.Object)
				continue
			}

			timestamp := time.Now().Format("15:04:05")
			switch event.Type {
			case watch.Added:
				fmt.Printf("[%s] ADDED: %s/%s (Status: %s)\n",
					timestamp, pod.Namespace, pod.Name, pod.Status.Phase)
			case watch.Modified:
				fmt.Printf("[%s] MODIFIED: %s/%s (Status: %s)\n",
					timestamp, pod.Namespace, pod.Name, pod.Status.Phase)
			case watch.Deleted:
				fmt.Printf("[%s] DELETED: %s/%s\n",
					timestamp, pod.Namespace, pod.Name)
			case watch.Error:
				fmt.Printf("[%s] ERROR: %v\n", timestamp, event.Object)
			default:
				fmt.Printf("[%s] UNKNOWN EVENT TYPE: %v for %s/%s\n",
					timestamp, event.Type, pod.Namespace, pod.Name)
			}

		case <-ctx.Done():
			fmt.Println("Watch timeout reached")
			return
		}
	}
}

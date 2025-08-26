package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"path/filepath"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func main() {
	var kubeconfig string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = filepath.Join(home, ".kube", "config")
	}

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		log.Fatalf("Failed to build kubeconfig: %v", err)
	}

	// Add HTTP request/response logging
	config.WrapTransport = func(rt http.RoundTripper) http.RoundTripper {
		return &detailedLoggingTransport{rt: rt, requestCount: 0}
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("Failed to create clientset: %v", err)
	}

	fmt.Println("=== Kubernetes Client-Go Demo with HTTP Logging ===")
	fmt.Println("Connected to Kubernetes cluster successfully!")

	fmt.Println("\n1. Listing Pods (watch for the HTTP request details):")
	listPods(clientset)

	fmt.Println("\n2. Watching Pods (watch for the streaming HTTP connection):")
	watchPods(clientset)
}

// Enhanced transport with detailed HTTP logging
type detailedLoggingTransport struct {
	rt           http.RoundTripper
	requestCount int
}

func (t *detailedLoggingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	t.requestCount++

	fmt.Printf("\n" + strings.Repeat("=", 80) + "\n")
	fmt.Printf("HTTP REQUEST #%d\n", t.requestCount)
	fmt.Printf(strings.Repeat("=", 80) + "\n")

	// Log the full HTTP request
	reqDump, err := httputil.DumpRequestOut(req, true)
	if err != nil {
		fmt.Printf("Failed to dump request: %v\n", err)
	} else {
		fmt.Printf("REQUEST:\n%s\n", string(reqDump))
	}

	// Make the actual request
	resp, err := t.rt.RoundTrip(req)
	if err != nil {
		fmt.Printf("Request failed: %v\n", err)
		return resp, err
	}

	fmt.Printf(strings.Repeat("-", 40) + "\n")
	fmt.Printf("HTTP RESPONSE #%d\n", t.requestCount)
	fmt.Printf(strings.Repeat("-", 40) + "\n")

	// For watch requests, don't read the full body as it's a stream
	if req.URL.Query().Get("watch") == "true" {
		fmt.Printf("RESPONSE (Watch Stream - headers only):\n")
		respDump, err := httputil.DumpResponse(resp, false)
		if err != nil {
			fmt.Printf("Failed to dump response: %v\n", err)
		} else {
			fmt.Printf("%s\n", string(respDump))
		}
		fmt.Printf("Note: This is a streaming response for watch requests\n")
	} else {
		// For non-watch requests, read and display the response
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Printf("Failed to read response body: %v\n", err)
			return resp, err
		}

		// Create a new reader for the response body
		resp.Body = io.NopCloser(bytes.NewReader(bodyBytes))

		// Truncate very long responses for readability
		displayBody := string(bodyBytes)
		if len(displayBody) > 2000 {
			displayBody = displayBody[:2000] + "\n... (truncated)"
		}

		fmt.Printf("RESPONSE:\n")
		fmt.Printf("Status: %s\n", resp.Status)
		fmt.Printf("Headers:\n")
		for key, values := range resp.Header {
			for _, value := range values {
				fmt.Printf("  %s: %s\n", key, value)
			}
		}
		fmt.Printf("Body:\n%s\n", displayBody)
	}

	fmt.Printf(strings.Repeat("=", 80) + "\n\n")
	return resp, nil
}

func listPods(clientset *kubernetes.Clientset) {
	ctx := context.Background()

	pods, err := clientset.CoreV1().Pods("").List(ctx, metav1.ListOptions{})
	if err != nil {
		log.Printf("Failed to list pods: %v", err)
		return
	}

	fmt.Printf("Found %d pods:\n", len(pods.Items))
	for _, pod := range pods.Items {
		fmt.Printf("  - %s/%s (Status: %s)\n",
			pod.Namespace, pod.Name, pod.Status.Phase)
	}
}

func watchPods(clientset *kubernetes.Clientset) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	watcher, err := clientset.CoreV1().Pods("").Watch(ctx, metav1.ListOptions{})
	if err != nil {
		log.Printf("Failed to create pod watcher: %v", err)
		return
	}
	defer watcher.Stop()

	fmt.Println("Watching for pod events...")

	eventCount := 0
	for {
		select {
		case event, ok := <-watcher.ResultChan():
			if !ok {
				fmt.Println("Watcher channel closed")
				return
			}

			pod, ok := event.Object.(*corev1.Pod)
			if !ok {
				continue
			}

			eventCount++
			timestamp := time.Now().Format("15:04:05")
			fmt.Printf("[%s] Event #%d - %s: %s/%s\n",
				timestamp, eventCount, event.Type, pod.Namespace, pod.Name)

			// Stop after a few events to keep the demo manageable
			if eventCount >= 5 {
				fmt.Println("Received enough events for demo purposes")
				return
			}

		case <-ctx.Done():
			fmt.Println("Watch timeout reached")
			return
		}
	}
}

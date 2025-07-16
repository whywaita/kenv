package resolver

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/whywaita/keex/pkg/extractor"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

type Resolver struct {
	client    kubernetes.Interface
	namespace string
}

type Options struct {
	Context   string
	Namespace string
}

func New(opts Options) (*Resolver, error) {
	kubeconfig := os.Getenv("KUBECONFIG")
	if kubeconfig == "" {
		home := homedir.HomeDir()
		kubeconfig = filepath.Join(home, ".kube", "config")
	}

	// Build config
	configLoadingRules := &clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfig}
	configOverrides := &clientcmd.ConfigOverrides{}
	if opts.Context != "" {
		configOverrides.CurrentContext = opts.Context
	}

	clientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		configLoadingRules,
		configOverrides,
	)

	config, err := clientConfig.ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to build kubeconfig: %w", err)
	}

	// Create client
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	// Get namespace
	namespace := opts.Namespace
	if namespace == "" {
		namespace, _, err = clientConfig.Namespace()
		if err != nil {
			namespace = "default"
		}
	}

	return &Resolver{
		client:    clientset,
		namespace: namespace,
	}, nil
}

func NewFromClientset(clientset kubernetes.Interface, namespace string) *Resolver {
	return &Resolver{
		client:    clientset,
		namespace: namespace,
	}
}

func (r *Resolver) ResolveAll(envVars []extractor.EnvVar) ([]extractor.EnvVar, error) {
	ctx := context.Background()
	resolved := make([]extractor.EnvVar, 0, len(envVars))

	// Cache for secrets and configmaps
	secretCache := make(map[string]*corev1.Secret)
	configMapCache := make(map[string]*corev1.ConfigMap)

	for _, envVar := range envVars {
		switch envVar.Source {
		case extractor.SourceSecret:
			if envVar.SecretRef != nil {
				secret, ok := secretCache[envVar.SecretRef.Name]
				if !ok {
					var err error
					secret, err = r.client.CoreV1().Secrets(r.namespace).Get(ctx, envVar.SecretRef.Name, metav1.GetOptions{})
					if err != nil {
						fmt.Fprintf(os.Stderr, "Warning: failed to get secret %s: %v\n", envVar.SecretRef.Name, err)
						resolved = append(resolved, envVar)
						continue
					}
					secretCache[envVar.SecretRef.Name] = secret
				}

				if value, ok := secret.Data[envVar.SecretRef.Key]; ok {
					envVar.Value = string(value)
				} else {
					fmt.Fprintf(os.Stderr, "Warning: key %s not found in secret %s\n", envVar.SecretRef.Key, envVar.SecretRef.Name)
				}
			}
			resolved = append(resolved, envVar)

		case extractor.SourceConfigMap:
			if envVar.ConfigRef != nil {
				configMap, ok := configMapCache[envVar.ConfigRef.Name]
				if !ok {
					var err error
					configMap, err = r.client.CoreV1().ConfigMaps(r.namespace).Get(ctx, envVar.ConfigRef.Name, metav1.GetOptions{})
					if err != nil {
						fmt.Fprintf(os.Stderr, "Warning: failed to get configmap %s: %v\n", envVar.ConfigRef.Name, err)
						resolved = append(resolved, envVar)
						continue
					}
					configMapCache[envVar.ConfigRef.Name] = configMap
				}

				if value, ok := configMap.Data[envVar.ConfigRef.Key]; ok {
					envVar.Value = value
				} else {
					fmt.Fprintf(os.Stderr, "Warning: key %s not found in configmap %s\n", envVar.ConfigRef.Key, envVar.ConfigRef.Name)
				}
			}
			resolved = append(resolved, envVar)

		default:
			resolved = append(resolved, envVar)
		}
	}

	return resolved, nil
}

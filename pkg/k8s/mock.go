package k8s

import (
	"github.com/openware/pkg/kube"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	fake "k8s.io/client-go/kubernetes/fake"
)

const (
	mockNamespace = "odax"
)

func MockSecret(name, namespace string, data map[string][]byte) *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   namespace,
			Annotations: map[string]string{},
		},
		Data: data,
	}
}

// NewMockClient returns an initialized Mock Client object
func NewMockClient(obj ...runtime.Object) *kube.K8sClient {
	client := &kube.K8sClient{
		Client: fake.NewSimpleClientset(obj...),
	}

	return client
}

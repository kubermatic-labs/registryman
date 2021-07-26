package config

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGetFileName(t *testing.T) {
	secret := &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		Immutable: nil,
	}
	secret.SetName("secret_test_name")
	if name := getFileName(secret); name != "secret_test_name.yaml" {
		t.Errorf("unexpected file name: %s", name)
	}

}

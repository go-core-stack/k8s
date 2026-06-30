package apply

import (
	"testing"

	"github.com/onsi/gomega"
	uns "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func TestApplyObjectServerSide_EmptyName(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	obj := &uns.Unstructured{}
	obj.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "apps",
		Version: "v1",
		Kind:    "Deployment",
	})
	// Name intentionally left empty

	err := ApplyObjectServerSide(t.Context(), nil, obj, "test-manager")
	g.Expect(err).To(gomega.HaveOccurred())
	g.Expect(err.Error()).To(gomega.ContainSubstring("has no name"))
}

func TestApplyObjectServerSide_EmptyFieldManager(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	obj := &uns.Unstructured{}
	obj.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "apps",
		Version: "v1",
		Kind:    "Deployment",
	})
	obj.SetName("test-deployment")
	obj.SetNamespace("default")

	err := ApplyObjectServerSide(t.Context(), nil, obj, "")
	g.Expect(err).To(gomega.HaveOccurred())
	g.Expect(err.Error()).To(gomega.ContainSubstring("fieldManager is required"))
}

func TestApplyObjectServerSide_UnsupportedObject(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	// ServiceAccount with secrets is unsupported
	obj := &uns.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "ServiceAccount",
			"metadata": map[string]interface{}{
				"name":      "test-sa",
				"namespace": "default",
			},
			"secrets": []interface{}{
				map[string]interface{}{"name": "some-secret"},
			},
		},
	}

	err := ApplyObjectServerSide(t.Context(), nil, obj, "test-manager")
	g.Expect(err).To(gomega.HaveOccurred())
	g.Expect(err.Error()).To(gomega.ContainSubstring("unsupported"))
}

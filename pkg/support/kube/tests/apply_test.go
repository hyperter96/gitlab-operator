package kubetests

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/pkg/support/kube"
	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/pkg/support/kube/apply"
)

var _ = Describe("ApplyObject", func() {
	It("uses strategic merge patch meta", func() {
		obj := readObject("deployment-1")
		Expect(
			kube.ApplyObject(obj, apply.WithManager(Manager)),
		).To(Equal(kube.ObjectCreated))

		/* wait for the change to be populated */
		d := &appsv1.Deployment{
			ObjectMeta: v1.ObjectMeta{
				Name:      obj.GetName(),
				Namespace: obj.GetNamespace(),
			},
		}
		Eventually(getObject(d)).Should(Succeed())
		g := d.ObjectMeta.Generation

		obj = readObject("deployment-2")
		Expect(
			kube.ApplyObject(obj, apply.WithManager(Manager)),
		).To(Equal(kube.ObjectUpdated))

		d = &appsv1.Deployment{
			ObjectMeta: v1.ObjectMeta{
				Name:      obj.GetName(),
				Namespace: obj.GetNamespace(),
			},
		}
		Eventually(getObject(d, g+1)).Should(Succeed())

		Expect(d.ObjectMeta.Generation).To(BeNumerically(">", g))
		Expect(d.Spec.Template.Spec.Volumes).To(HaveLen(3))
		Expect(d.Spec.Template.Spec.Volumes[2].Name).To(Equal("dummy"))
		Expect(d.Spec.Template.Spec.Containers[0].VolumeMounts).To(HaveLen(3))
		Expect(d.Spec.Template.Spec.Containers[0].VolumeMounts[0].Name).To(Equal("dummy"))

		Eventually(deleteObject(d)).Should(Succeed())
	})

	It("patches the object when its last configuration is unknown", func() {
		obj := readObject("deployment-1")
		Eventually(createObject(obj)).Should(Succeed())

		/* wait for the change to be populated */
		d := &appsv1.Deployment{
			ObjectMeta: v1.ObjectMeta{
				Name:      obj.GetName(),
				Namespace: obj.GetNamespace(),
			},
		}
		Eventually(getObject(d)).Should(Succeed())
		g := d.ObjectMeta.Generation

		obj = readObject("deployment-1")
		Expect(
			kube.ApplyObject(obj, apply.WithManager(Manager)),
		).To(Equal(kube.ObjectUpdated))

		/* wait for the change to be populated */
		d = &appsv1.Deployment{
			ObjectMeta: v1.ObjectMeta{
				Name:      obj.GetName(),
				Namespace: obj.GetNamespace(),
			},
		}
		Eventually(getObject(d, g+1)).Should(Succeed())

		Expect(d.ObjectMeta.Generation).To(BeNumerically(">", g))
		Expect(d.Annotations).To(HaveKey(corev1.LastAppliedConfigAnnotation))
		Expect(d.Annotations[corev1.LastAppliedConfigAnnotation]).NotTo(BeEmpty())

		Eventually(deleteObject(d)).Should(Succeed())
	})

	It("does not patch the object when it is not changed", func() {
		obj := readObject("job-1")
		Expect(
			kube.ApplyObject(obj, apply.WithManager(Manager)),
		).To(Equal(kube.ObjectCreated))

		/* wait for the change to be populated */
		j := &batchv1.Job{
			ObjectMeta: v1.ObjectMeta{
				Name:      obj.GetName(),
				Namespace: obj.GetNamespace(),
			},
		}
		Eventually(getObject(j)).Should(Succeed())
		g := j.ObjectMeta.Generation

		obj = readObject("job-1")
		Expect(
			kube.ApplyObject(obj, apply.WithManager(Manager)),
		).To(Equal(kube.ObjectUnchanged))

		Expect(j.ObjectMeta.Generation).To(Equal(g))

		Eventually(deleteObject(obj)).Should(Succeed())
	})

	It("does not modify the object when it is not semantically changed", func() {
		obj := readObject("deployment-1")
		Expect(
			kube.ApplyObject(obj, apply.WithManager(Manager)),
		).To(Equal(kube.ObjectCreated))

		/* wait for the change to be populated */
		d := &appsv1.Deployment{
			ObjectMeta: v1.ObjectMeta{
				Name:      obj.GetName(),
				Namespace: obj.GetNamespace(),
			},
		}
		Eventually(getObject(d)).Should(Succeed())
		g := d.ObjectMeta.Generation

		obj = readObject("deployment-1")
		Expect(
			kube.ApplyObject(obj, apply.WithManager(Manager)),
		).To(Equal(kube.ObjectUnchanged))

		d = &appsv1.Deployment{
			ObjectMeta: v1.ObjectMeta{
				Name:      obj.GetName(),
				Namespace: obj.GetNamespace(),
			},
		}
		Eventually(getObject(d, g)).Should(Succeed())

		Expect(d.ObjectMeta.Generation).To(Equal(g))

		Eventually(deleteObject(d)).Should(Succeed())
	})

	It("can work with unstructured objects", func() {
		obj := readObject("deployment-1", UnstructuredYAMLCodec)
		Expect(
			kube.ApplyObject(obj, apply.WithManager(Manager)),
		).To(Equal(kube.ObjectCreated))

		/* wait for the change to be populated */
		d := &appsv1.Deployment{
			ObjectMeta: v1.ObjectMeta{
				Name:      obj.GetName(),
				Namespace: obj.GetNamespace(),
			},
		}
		Eventually(getObject(d)).Should(Succeed())
		g := d.ObjectMeta.Generation

		obj = readObject("deployment-2", UnstructuredYAMLCodec)
		Expect(
			kube.ApplyObject(obj, apply.WithManager(Manager)),
		).To(Equal(kube.ObjectUpdated))

		d = &appsv1.Deployment{
			ObjectMeta: v1.ObjectMeta{
				Name:      obj.GetName(),
				Namespace: obj.GetNamespace(),
			},
		}
		Eventually(getObject(d, g+1)).Should(Succeed())

		Expect(d.ObjectMeta.Generation).To(BeNumerically(">", g))
		Expect(d.Spec.Template.Spec.Volumes).To(HaveLen(3))
		Expect(d.Spec.Template.Spec.Volumes[2].Name).To(Equal("dummy"))

		Eventually(deleteObject(obj)).Should(Succeed())
	})

	/*
	 * Testing unregistered types has proven to be difficult here. These types
	 * must be recognized by the mock Kubernetes API Server but not registered
	 * in the scheme.
	 */
})

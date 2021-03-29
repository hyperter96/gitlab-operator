package gitlab

import (
	"strconv"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	gitlabv1beta1 "gitlab.com/gitlab-org/gl-openshift/gitlab-operator/api/v1beta1"
	"gitlab.com/gitlab-org/gl-openshift/gitlab-operator/controllers/helpers"
	"gitlab.com/gitlab-org/gl-openshift/gitlab-operator/controllers/settings"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Helpers", func() {

	localUser, _ := strconv.ParseInt(settings.LocalUser, 10, 64)

	if namespace == "" {
		namespace = "default"
	}

	mockGitLab := &gitlabv1beta1.GitLab{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: namespace,
		},
		Spec: gitlabv1beta1.GitLabSpec{
			Chart: gitlabv1beta1.GitLabChartSpec{
				Version: chartVersion,
			},
		},
	}

	Context("PostgreSQL", func() {
		adapter := helpers.NewCustomResourceAdapter(mockGitLab)

		It("should specify the SecurityContext and ServiceAccount of the StatefulSet", func() {

			statefulset := PostgresStatefulSet(adapter)

			Expect(statefulset.Spec.Template.Spec.ServiceAccountName).To(Equal(settings.AppServiceAccount))

			Expect(statefulset.Spec.Template.Spec.SecurityContext).NotTo(BeNil())
			Expect(statefulset.Spec.Template.Spec.SecurityContext.FSGroup).NotTo(BeNil())
			Expect(statefulset.Spec.Template.Spec.Containers[0].SecurityContext.RunAsUser).NotTo(BeNil())

			fsGroup := statefulset.Spec.Template.Spec.SecurityContext.FSGroup
			runAsUser := statefulset.Spec.Template.Spec.Containers[0].SecurityContext.RunAsUser

			Expect(*fsGroup).To(Equal(localUser))
			Expect(*runAsUser).To(Equal(localUser))
		})
	})

	Context("Redis", func() {
		adapter := helpers.NewCustomResourceAdapter(mockGitLab)

		It("should specify the SecurityContext and ServiceAccount of the StatefulSet", func() {

			statefulset := RedisStatefulSet(adapter)

			Expect(statefulset.Spec.Template.Spec.ServiceAccountName).To(Equal(settings.AppServiceAccount))

			Expect(statefulset.Spec.Template.Spec.SecurityContext).NotTo(BeNil())
			Expect(statefulset.Spec.Template.Spec.SecurityContext.FSGroup).NotTo(BeNil())
			Expect(statefulset.Spec.Template.Spec.SecurityContext.RunAsUser).NotTo(BeNil())

			fsGroup := statefulset.Spec.Template.Spec.SecurityContext.FSGroup
			runAsUser := statefulset.Spec.Template.Spec.SecurityContext.RunAsUser

			Expect(*fsGroup).To(Equal(localUser))
			Expect(*runAsUser).To(Equal(localUser))
		})
	})
})

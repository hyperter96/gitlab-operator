package controllers

import (
	"context"
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	gitlabv1beta1 "gitlab.com/gitlab-org/gl-openshift/gitlab-operator/api/v1beta1"
	"gitlab.com/gitlab-org/gl-openshift/gitlab-operator/controllers/helpers"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	ctx = context.Background()
)

func newGitLab(releaseName string) *gitlabv1beta1.GitLab {
	return &gitlabv1beta1.GitLab{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps.gitlab.com/v1beta1",
			Kind:       "GitLab",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      releaseName,
			Namespace: Namespace,
			Labels: map[string]string{
				"chart": fmt.Sprintf("gitlab-%s", helpers.AvailableChartVersions()[0]),
			},
		},
		Spec: gitlabv1beta1.GitLabSpec{
			AutoScaling: &gitlabv1beta1.AutoScalingSpec{},
			Database: &gitlabv1beta1.DatabaseSpec{
				Volume: gitlabv1beta1.VolumeSpec{
					Capacity: "100",
				},
			},
			Redis: &gitlabv1beta1.RedisSpec{},
			Volume: gitlabv1beta1.VolumeSpec{
				Capacity: "100",
			},
		},
	}
}

func createObject(obj runtime.Object, ignoreAlreadyExists bool) error {
	err := k8sClient.Create(ctx, obj)
	if errors.IsAlreadyExists(err) && ignoreAlreadyExists {
		err = nil
	}
	return err
}

func getObject(name string, obj runtime.Object) error {
	lookupKey := types.NamespacedName{
		Name:      name,
		Namespace: Namespace,
	}
	return k8sClient.Get(ctx, lookupKey, obj)
}

func getObjectPromise(name string, obj runtime.Object) func() error {
	return func() error {
		return getObject(name, obj)
	}
}

func listObjects(query string, obj runtime.Object) error {
	labelSelector, err := labels.Parse(query)
	if err != nil {
		return err
	}

	listOptions := &client.ListOptions{
		Namespace:     Namespace,
		LabelSelector: labelSelector,
	}

	if err := k8sClient.List(ctx, obj, listOptions); err != nil {
		return err
	}

	return nil
}

func listObjectsPromise(query string, obj runtime.Object) func() error {
	return func() error {
		if err := listObjects(query, obj); err != nil {
			return err
		}
		switch l := obj.(type) {
		case *batchv1.JobList:
			if len(l.Items) == 0 {
				return fmt.Errorf("Job not found [%s]", query)
			}
		}
		return nil
	}
}

func deleteObject(name string, obj runtime.Object, ignoreNotExistis bool) error {
	if err := getObject(name, obj); err != nil {
		if errors.IsNotFound(err) {
			err = nil
		}
		return err
	}
	return k8sClient.Delete(ctx, obj)
}

func deleteObjectPromise(name string, obj runtime.Object) func() error {
	return func() error {
		return deleteObject(name, obj, true)
	}
}

func listConfigMapsPromise(query string) func() []corev1.ConfigMap {
	return func() []corev1.ConfigMap {
		createdCfgMaps := &corev1.ConfigMapList{}
		if err := listObjects(query, createdCfgMaps); err != nil {
			return nil
		}
		return createdCfgMaps.Items
	}
}

func updateJobStatusPromise(query string, success bool) func() error {
	return func() error {
		createdJobs := &batchv1.JobList{}

		if err := listObjects(query, createdJobs); err != nil {
			return err
		}

		if len(createdJobs.Items) == 0 {
			return fmt.Errorf("Job not found [%s]", query)
		}

		if success {
			createdJobs.Items[0].Status.Succeeded = 1
			createdJobs.Items[0].Status.Failed = 0
		} else {
			createdJobs.Items[0].Status.Succeeded = 0
			createdJobs.Items[0].Status.Failed = 1
		}

		return k8sClient.Status().Update(ctx, &createdJobs.Items[0])
	}
}

func appLabels(releaseName, appName string) string {
	return fmt.Sprintf("release=%s,app=%s", releaseName, appName)
}

func componentLabels(releaseName, componentName string) string {
	return fmt.Sprintf("release=%s,app.kubernetes.io/component=%s", releaseName, componentName)
}

func createGitLabResource(releaseName string) {
	By("Creating a new GitLab resource")
	Expect(createObject(newGitLab(releaseName), true)).Should(Succeed())

	By("Checking GitLab resource is created")
	Eventually(getObjectPromise(releaseName, &gitlabv1beta1.GitLab{}),
		PollTimeout, PollInterval).Should(Succeed())
}

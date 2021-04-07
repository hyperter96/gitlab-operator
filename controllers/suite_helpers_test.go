package controllers

import (
	"context"
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	gitlabv1beta1 "gitlab.com/gitlab-org/gl-openshift/gitlab-operator/api/v1beta1"
	"gitlab.com/gitlab-org/gl-openshift/gitlab-operator/controllers/gitlab"
	"gitlab.com/gitlab-org/gl-openshift/gitlab-operator/helm"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	ctx          = context.Background()
	chartVersion = gitlab.AvailableChartVersions()[0]
	emptyValues  = helm.EmptyValues()
)

func newGitLab(releaseName string, chartValues helm.Values) *gitlabv1beta1.GitLab {
	return &gitlabv1beta1.GitLab{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps.gitlab.com/v1beta1",
			Kind:       "GitLab",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      releaseName,
			Namespace: Namespace,
		},
		Spec: gitlabv1beta1.GitLabSpec{
			Chart: gitlabv1beta1.GitLabChartSpec{
				Version: chartVersion,
				Values: gitlabv1beta1.ChartValues{
					Object: chartValues.AsMap(),
				},
			},
		},
	}
}

func createObject(obj client.Object, ignoreAlreadyExists bool) error {
	err := k8sClient.Create(ctx, obj)
	if errors.IsAlreadyExists(err) && ignoreAlreadyExists {
		err = nil
	}
	return err
}

func getObject(name string, obj client.Object) error {
	lookupKey := types.NamespacedName{
		Name:      name,
		Namespace: Namespace,
	}
	return k8sClient.Get(ctx, lookupKey, obj)
}

func getObjectPromise(name string, obj client.Object) func() error {
	return func() error {
		return getObject(name, obj)
	}
}

func listObjects(query string, obj client.ObjectList) error {
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

func listObjectsPromise(query string, obj client.ObjectList, expectedSize int) func() error {
	return func() error {
		if err := listObjects(query, obj); err != nil {
			return err
		}
		switch l := obj.(type) {
		case *batchv1.JobList:
			if len(l.Items) < expectedSize {
				return fmt.Errorf("Only %d Jobs found with [%s]", expectedSize, query)
			}
		}
		return nil
	}
}

func deleteObject(name string, obj client.Object, ignoreNotExistis bool) error {
	if err := getObject(name, obj); err != nil {
		if errors.IsNotFound(err) {
			err = nil
		}
		return err
	}
	return k8sClient.Delete(ctx, obj)
}

func deleteObjectPromise(name string, obj client.Object) func() error {
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

		for _, j := range createdJobs.Items {
			if success {
				j.Status.Succeeded = 1
				j.Status.Failed = 0
			} else {
				j.Status.Succeeded = 0
				j.Status.Failed = 1
			}
			if err := k8sClient.Status().Update(ctx, &j); err != nil {
				return err
			}
		}

		return nil
	}
}

func appLabels(releaseName, appName string) string {
	return fmt.Sprintf("release=%s,app=%s", releaseName, appName)
}

func componentLabels(releaseName, componentName string) string {
	return fmt.Sprintf("release=%s,app.kubernetes.io/component=%s", releaseName, componentName)
}

func createGitLabResource(releaseName string, chartValues helm.Values) {
	By("Creating a new GitLab resource")
	Expect(createObject(newGitLab(releaseName, chartValues), true)).Should(Succeed())

	By("Checking GitLab resource is created")
	Eventually(getObjectPromise(releaseName, &gitlabv1beta1.GitLab{}),
		PollTimeout, PollInterval).Should(Succeed())
}

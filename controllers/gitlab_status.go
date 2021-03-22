package controllers

import (
	"context"
	"reflect"

	"github.com/prometheus/common/log"
	gitlabv1beta1 "gitlab.com/gitlab-org/gl-openshift/gitlab-operator/api/v1beta1"
	"gitlab.com/gitlab-org/gl-openshift/gitlab-operator/controllers/helpers"
	gitlabutils "gitlab.com/gitlab-org/gl-openshift/gitlab-operator/controllers/utils"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
)

// EndpointMembers returns a list of members
var EndpointMembers []string

func (r *GitLabReconciler) reconcileGitlabStatus(ctx context.Context, adapter helpers.CustomResourceAdapter) error {

	lookupKey := types.NamespacedName{Namespace: adapter.Namespace(), Name: adapter.ReleaseName()}

	r.Log.V(1).Info("Updating GitLab resource status", "resource", lookupKey)

	// get current Gitlab resource
	gitlab := &gitlabv1beta1.GitLab{}
	err := r.Get(ctx, lookupKey, gitlab)
	if err != nil {
		return err
	}

	// Check if the postgres statefulset exists
	if r.isPostgresDeployed(ctx, adapter) && !r.isWebserviceDeployed(ctx, adapter) {
		gitlab.Status.Phase = "Initializing"
		gitlab.Status.Stage = "Waiting for database"
	}

	// Check if the webservice deployment exists
	if r.isWebserviceDeployed(ctx, adapter) {
		// Find webservice pod(s)
		pods := r.getEndpointMembers(ctx, adapter, adapter.ReleaseName()+"-webservice-default")
		if len(pods) == 0 {
			gitlab.Status.Phase = "Initializing"
			gitlab.Status.Stage = "Gitlab is initializing"
		}

		if len(pods) > 0 {
			gitlab.Status.Phase = "Running"
			gitlab.Status.Stage = ""

			// Temporarily disabled.
			/*
				gitlab.Status.HealthCheck = getReadinessStatus(ctx, adapter)
			*/
		}
	}

	// Check if the status of the gitlab resource has changed
	if !reflect.DeepEqual(adapter.Resource().Status, gitlab.Status) {
		// Update status if the status has changed
		if err := r.setGitlabStatus(ctx, gitlab); err != nil {
			return err
		}
	}

	return nil
}

// Temporarily disabled.
/*
func getReadinessStatus(ctx context.Context, adapter helpers.CustomResourceAdapter) *gitlabv1beta1.HealthCheck {
	var err error
	status := &ReadinessStatus{}

	resp, err := http.Get(fmt.Sprintf("http://%s:8181/-/readiness?all=1", adapter.ReleaseName()+"-webservice"))
	if err != nil {
		// log.Error(err, "Unable to retrieve status")
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		// Read readiness status response
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			// log.Error(err, "Unable to read status")
			return nil
		}

		if err = json.Unmarshal(body, status); err != nil {
			// log.Error(err, "Unable to convert response to struct")
			return nil
		}
	}

	return parseStatus(status)
}
*/

// ReadinessStatus shows status of Gitlab services
type ReadinessStatus struct {
	// Returns status of Gitlab rails app
	WorkhorseStatus string `json:"status,omitempty"`
	// RedisStatus reports status of redis
	RedisStatus []ServiceStatus `json:"redis_check,omitempty"`
	// DatabaseStatus reports status of postgres
	DatabaseStatus []ServiceStatus `json:"db_check,omitempty"`
}

// ServiceStatus shows status of a Gitlab
// dependent service .e.g. Postgres, Redis, Gitaly
type ServiceStatus struct {
	Status string `json:"status,omitempty"`
}

// Temporarily disabled.
// Retrieve health of a subsystem
/*
func parseStatus(status *ReadinessStatus) *gitlabv1beta1.HealthCheck {
	var result gitlabv1beta1.HealthCheck

	if status.WorkhorseStatus != "" {
		result.Workhorse = status.WorkhorseStatus
	}
	// Get redis status
	if len(status.RedisStatus) == 1 {
		result.Redis = status.RedisStatus[0].Status
	}

	// Get postgresql status
	if len(status.DatabaseStatus) == 1 {
		result.Postgres = status.DatabaseStatus[0].Status
	}

	return &result
}
*/

func (r *GitLabReconciler) isPostgresDeployed(ctx context.Context, adapter helpers.CustomResourceAdapter) bool {
	labels := gitlabutils.Label(adapter.ReleaseName(), "postgresql", gitlabutils.GitlabType)

	postgres := &appsv1.StatefulSet{}
	err := r.Get(ctx, types.NamespacedName{Namespace: adapter.Namespace(), Name: labels["app.kubernetes.io/instance"]}, postgres)
	return !reflect.DeepEqual(*postgres, appsv1.Deployment{}) || !errors.IsNotFound(err)
}

func (r *GitLabReconciler) isWebserviceDeployed(ctx context.Context, adapter helpers.CustomResourceAdapter) bool {
	webservice := &appsv1.Deployment{}
	err := r.Get(ctx, types.NamespacedName{Name: adapter.ReleaseName() + "-webservice-default", Namespace: adapter.Namespace()}, webservice)
	return !reflect.DeepEqual(*webservice, appsv1.Deployment{}) || !errors.IsNotFound(err)
}

// setGitlabStatus sets status of custom resource
func (r *GitLabReconciler) setGitlabStatus(ctx context.Context, object runtime.Object) error {
	return r.Status().Update(ctx, object)
}

func (r *GitLabReconciler) getEndpointMembers(ctx context.Context, adapter helpers.CustomResourceAdapter, endpoint string) []string {
	members := []string{}

	ep := &corev1.Endpoints{}
	err := r.Get(ctx, types.NamespacedName{Name: endpoint, Namespace: adapter.Namespace()}, ep)
	if err != nil {
		log.Error(err, "Error getting endpoints")
	}

	for _, subset := range ep.Subsets {
		// Get members that are ready
		for _, addr := range subset.Addresses {
			members = append(members, addr.TargetRef.Name)
		}
	}

	return members
}

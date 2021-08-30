package helm

import (
	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func (q *cachingQuery) JobByName(name string) *batchv1.Job {
	key := q.cacheKey(name, gvkJob, nil)
	result := q.runQuery(key,
		func() interface{} {
			objects, err := q.template.GetObjects(
				NewJobSelector(
					func(d *batchv1.Job) bool {
						return d.ObjectMeta.Name == name
					},
				),
			)

			if err != nil {
				return nil
			}

			return unsafeConvertJobs(objects)
		},
	)

	jobs := result.([]*batchv1.Job)

	if len(jobs) == 0 {
		return nil
	}

	return jobs[0]
}

func (q *cachingQuery) JobsByLabels(labels map[string]string) []*batchv1.Job {
	key := q.cacheKey(anything, gvkJob, labels)
	result := q.runQuery(key,
		func() interface{} {
			objects, err := q.template.GetObjects(
				NewJobSelector(
					func(d *batchv1.Job) bool {
						return matchLabels(d.ObjectMeta.Labels, labels)
					},
				),
			)

			if err != nil {
				return nil
			}

			return unsafeConvertJobs(objects)
		},
	)

	return result.([]*batchv1.Job)
}

func (q *cachingQuery) JobByComponent(component string) *batchv1.Job {
	jobs := q.JobsByLabels(map[string]string{
		appLabel: component,
	})

	if len(jobs) == 0 {
		return nil
	}

	return jobs[0]
}

func unsafeConvertJobs(objects []runtime.Object) []*batchv1.Job {
	jobs := make([]*batchv1.Job, len(objects))
	for i, o := range objects {
		jobs[i] = o.(*batchv1.Job)
	}

	return jobs
}

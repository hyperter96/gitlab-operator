package helm

import (
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
)

func (q *cachingQuery) CronJobByName(name string) *batchv1beta1.CronJob {
	key := q.cacheKey(name, gvkJob, nil)
	result := q.runQuery(key,
		func() interface{} {
			objects, err := q.template.GetObjects(
				NewCronJobSelector(
					func(d *batchv1beta1.CronJob) bool {
						return d.ObjectMeta.Name == name
					},
				),
			)

			if err != nil {
				return nil
			}

			return unsafeConvertCronJobs(objects)
		},
	)

	jobs := result.([]*batchv1beta1.CronJob)

	if len(jobs) == 0 {
		return nil
	}

	return jobs[0]
}

func unsafeConvertCronJobs(objects []runtime.Object) []*batchv1beta1.CronJob {
	jobs := make([]*batchv1beta1.CronJob, len(objects))
	for i, o := range objects {
		jobs[i] = o.(*batchv1beta1.CronJob)
	}

	return jobs
}

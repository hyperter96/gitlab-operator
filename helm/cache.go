package helm

import (
	"fmt"
)

func (q *cachingQuery) cacheKey(nameOrComponent, gvk string, labels map[string]string) string {
	return fmt.Sprintf("%s.%s[%s]", nameOrComponent, gvk, labels)
}

func (q *cachingQuery) readCache(key string) interface{} {
	q.locker.Lock()
	defer q.locker.Unlock()

	if result, ok := q.cache[key]; ok {
		return result
	}

	return nil
}

func (q *cachingQuery) updateCache(key string, objects interface{}) {
	if objects == nil {
		return
	}

	q.locker.Lock()
	defer q.locker.Unlock()

	q.cache[key] = objects
}

func (q *cachingQuery) clearCache() {
	q.locker.Lock()
	defer q.locker.Unlock()

	q.cache = make(map[string]interface{})
}

func cacheBackdoor(q Query) *map[string]interface{} {
	if cq, ok := (q).(*cachingQuery); ok {
		return &cq.cache
	}

	return nil
}

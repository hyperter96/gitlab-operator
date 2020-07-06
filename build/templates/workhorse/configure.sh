set -e
mkdir -p /init-secrets-workhorse/gitlab-workhorse
cp -v -r -L /init-config/gitlab-workhorse/secret /init-secrets-workhorse/gitlab-workhorse/secret
mkdir -p /init-secrets-workhorse/redis
cp -v -r -L /init-config/redis/redis-password /init-secrets-workhorse/redis/
set -e
mkdir -p /init-secrets/gitaly /init-secrets/shell
cp -v -r -L /init-config/.gitlab_shell_secret  /init-secrets/shell/.gitlab_shell_secret
cp -v -r -L /init-config/gitaly_token  /init-secrets/gitaly/gitaly_token
mkdir -p /init-secrets/redis
cp -v -r -L /init-config/redis_password  /init-secrets/redis/redis_password

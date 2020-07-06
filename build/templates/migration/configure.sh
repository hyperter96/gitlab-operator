set -e
config_dir="/init-config"
secret_dir="/init-secrets"

for secret in postgres rails-secrets migrations gitaly ; do
  mkdir -p "${secret_dir}/${secret}"
  cp -v -r -L "${config_dir}/${secret}/." "${secret_dir}/${secret}/"
done
for secret in redis minio objectstorage ldap omniauth smtp ; do
  if [ -e "${config_dir}/${secret}" ]; then
    mkdir -p "${secret_dir}/${secret}"
    cp -v -r -L "${config_dir}/${secret}/." "${secret_dir}/${secret}/"
  fi
done

if [ -e /config/accesskey ] ; then
  sed -e 's@ACCESS_KEY@'"$(cat /config/accesskey)"'@' -e 's@SECRET_KEY@'"$(cat /config/secretkey)"'@' /config/config.yml > /registry/config.yml
else
  cp -v -r -L /config/config.yml  /registry/config.yml
fi
# Place the `http.secret` value from the kubernetes secret
sed -i -e 's@HTTP_SECRET@'"$(cat /config/httpSecret)"'@' /registry/config.yml
# Insert any provided `storage` block from kubernetes secret
if [ -d /config/storage ]; then
  # Copy contents of storage secret(s)
  mkdir -p /registry/storage
  cp -v -r -L /config/storage/* /registry/storage/
  # Ensure there is a new line in the end
  echo '' >> /registry/storage/config
  # Default `delete.enabled: true` if not present.
  ## Note: busybox grep doesn't support multiline, so we chain `egrep`.
  if ! $(egrep -A1 '^delete:\s*$' /registry/storage/config | egrep -q '\s{2,4}enabled:') ; then
    echo 'delete:' >> /registry/storage/config
    echo '  enabled: true' >> /registry/storage/config
  fi
  # Indent /registry/storage/config 2 spaces before inserting into config.yml
  sed -i 's/^/  /' /registry/storage/config
  # Insert into /registry/config.yml after `storage:`
  sed -i '/storage:/ r /registry/storage/config' /registry/config.yml
  # Remove the now extraneous `config` file
  rm /registry/storage/config
fi
# Set to known path, to used ConfigMap
cat /config/certificate.crt > /registry/certificate.crt
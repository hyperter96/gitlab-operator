sed -e 's@ACCESS_KEY@'"$(cat /config/accesskey)"'@' -e 's@SECRET_KEY@'"$(cat /config/secretkey)"'@' /config/config.json > /minio/config.json

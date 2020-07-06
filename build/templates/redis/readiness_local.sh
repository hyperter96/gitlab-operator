password_aux=`cat ${REDIS_PASSWORD_FILE}`
export REDIS_PASSWORD=$password_aux
response=$(
  timeout -s 9 $1 \
  redis-cli \
    -a $REDIS_PASSWORD --no-auth-warning \
    -h localhost \
    -p $REDIS_PORT \
    ping
)
if [ "$response" != "PONG" ]; then
  echo "$response"
  exit 1
fi

password_aux=`cat ${REDIS_MASTER_PASSWORD_FILE}`
export REDIS_MASTER_PASSWORD=$password_aux
response=$(
  timeout -s 9 $1 \
  redis-cli \
    -a $REDIS_MASTER_PASSWORD --no-auth-warning \
    -h $REDIS_MASTER_HOST \
    -p $REDIS_MASTER_PORT_NUMBER \
    ping
)
if [ "$response" != "PONG" ]; then
  echo "$response"
  exit 1
fi

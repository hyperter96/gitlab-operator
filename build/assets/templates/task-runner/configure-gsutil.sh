# The following script is used to configure gsutil when creating backups
# It provides inputs to the `gsutil config -e` prompt as follows:
# 1) Path to service account JSON key file
# 2) Do not set permissions for key file
# 3) GCP Project ID
# 4) Decline anonymous usage statistics
printf "$GOOGLE_APPLICATION_CREDENTIALS\nN\n\nN\n" | gsutil config -e
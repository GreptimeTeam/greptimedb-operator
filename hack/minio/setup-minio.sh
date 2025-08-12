#!/bin/bash

set -e

BUCKET_NAME=testbucket
USER_NAME=testuser
USER_PASSWORD=testpassword
ACCESS_KEY=testaccesskey
SECRET_KEY=testsecretkey
# You may need to execute the port-forwarding command to expose the minio service to the local machine.
# kubectl port-forward svc/minio-dev 9000:9000 9001:9001
MINIO_ENDPOINT=http://localhost:9000
MINIO_ROOT_USER=minioadmin
MINIO_ROOT_PASSWORD=minioadmin

# Configure mc client to connect to MinIO.
echo "Configuring mc client..."
mc alias set local $MINIO_ENDPOINT $MINIO_ROOT_USER $MINIO_ROOT_PASSWORD

# If the bucket already exists, exit 0.
if mc ls local/$BUCKET_NAME; then
  echo "Bucket already exists, exiting..."
  exit 0
fi

# Create test bucket.
echo "Creating test bucket..."
mc mb local/$BUCKET_NAME

# Set bucket policy to public read (optional).
echo "Setting bucket policy..."
mc anonymous set public local/$BUCKET_NAME

# Create a test user and access key.
echo "Creating test user..."
mc admin user add local $USER_NAME $USER_PASSWORD

# Create access key.
echo "Creating access key..."
mc admin user svcacct add local $USER_NAME --access-key $ACCESS_KEY --secret-key $SECRET_KEY

# Attach readwrite policy to the user.
mc admin policy attach local readwrite --user $USER_NAME

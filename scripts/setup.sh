#!/bin/bash -x
cd /data
aws s3 cp s3://${BUCKET}/${NAME}.tgz /
tar xzf /${NAME}.tgz

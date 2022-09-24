#!/bin/bash -x
cd /data
aws s3 cp s3://${BUCKET}/${NAME}.tgz .
tar xzvf ${NAME}.tgz

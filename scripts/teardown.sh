#!/bin/bash -x
cd /data
tar czf /${NAME}.tgz * --exlcude logs --exlcude cache
aws s3 cp /${NAME}.tgz s3://${BUCKET}/
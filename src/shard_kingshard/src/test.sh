#!/bin/bash

org_sql=$1
echo ${org_sql}
num=$2
echo ${num}

if [ ! -f ${org_sql} ];then
    echo "no exist ${org_sql}"
    exit 1
fi

for ((i=0; i<${num}; i++))
do
    file1="${org_sql}".${i}
    file="${org_sql}".${i}.tmp
    cat ${org_sql} > ${file}

    sed -i "_bak" "s/mbk_orders/mbk_shard_orders_0/" ${file1}
done

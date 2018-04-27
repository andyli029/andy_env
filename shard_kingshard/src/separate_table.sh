#!/bin/bash

echo "abandoned，please use the go file"
exit 1;

if [ $# != 1 ] ; then
echo "USAGE: $0 YAMLNAME"
exit 1;
fi

yaml=$1

addr=`cat ${yaml} |grep 'addr:'| sed 's/^.*addr: //g'`
echo ${addr}

user=`cat ${yaml} |grep 'user:'|sed 's/^.*user: //g'`
echo ${user}


password=`cat ${yaml} |grep 'password:'|sed 's/^.*password: //g'`
echo ${password}

dbname=`cat ${yaml} |grep 'dbname:'|sed 's/^.*dbname: //g'`
echo ${dbname}

table=`cat ${yaml} |grep 'table:'|sed 's/^.*table: //g'`
echo ${table}

sqlfile=`cat ${yaml} |grep 'sqlfile:'|sed 's/^.*sqlfile: //g'`
echo ${sqlfile}

num=`cat ${yaml} |grep 'num:'|sed 's/^.*num: //g'`
echo ${num}

exit


exit

org_sql=$1
num=$2
#echo ${org_sql}
#echo ${num}

if [ ! -f ${org_sql} ];then
    echo "no exist ${org_sql}"
    exit 1
fi

for ((i=0; i<${num}; i++))
do
    file="${org_sql}"
    newfile="${org_sql}".${i}
    #file="${org_sql}".${i}.tmp
    cat ${org_sql} > ${newfile}

    cat ${file}| while read line
    #for line in `cat ${file}`
    do
        #echo $line
        tablename=`echo ${line} |grep 'CREATE TABLE'| sed 's/^.*TABLE \`//g' | sed 's/\` (*$//g'`
        if [[ ${tablename} =~ "mbk" ]];then

            newtabname=`echo ${tablename}|sed 's/mbk/mbk_shard/g'`
            newtabname="${newtabname}"_"${i}"

            #echo ${tablename}
            #echo ${newtabname}
            #echo ${newfile}
            # the below -i is in mac, urgly!!!!
            sed -i "_bak" "s/${tablename}/${newtabname}/g" ${newfile}
            #echo $?
        else
            #echo "invalid."
            continue
        fi

    done

    bakfile="${newfile}"_bak
    rm -rf ${bakfile}
    #echo
done

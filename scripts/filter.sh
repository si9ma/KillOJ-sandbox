#!/bin/bash
# use to filter needed system call

conv() {
    for item in $1
    do
        echo -n "\"$item\","
    done
}

all_sys=`cat all_sys`
remain_sys=$all_sys
test_sys=
cnt=0
for sys in $all_sys
do
    cnt=`echo $cnt + 1 | bc`
    test_sys=`sort <(echo "$remain_sys") <(echo "$sys") | uniq -u`
    parm=`conv "$test_sys"`
    statement=`printf "\"fk\":{%s}," $parm`
    sed -i '/"fk":{/c\'"$statement" init.go
    make
    timeout 3s bash -c "./kbox run --dir="/tmp/kbox/test" --cmd="/Main" --expected="fk" --input="fk" --timeout=10 --seccomp --memory=25600000 | tee output"
    grep "bad" output || remain_sys=$test_sys && echo "remove $sys"
    grep "bad" output && echo "save $sys"
    echo "remain:"`echo "$remain_sys" | wc -l` "loop:$cnt"
    echo "$remain_sys" > remain
done

conv "$remain_sys"

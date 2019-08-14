#!/bin/bash

function test_connection() {
    nginx_svc_ip_cluster3=$(kubectl --context=cluster3 get svc -l app=nginx-demo | awk 'FNR == 2 {print $3}')
    netshoot_pod=$(kubectl --context=cluster2 get pods -l app=netshoot | awk 'FNR == 2 {print $1}')

    echo "Testing connectivity between clusters - $netshoot_pod cluster2 --> $nginx_svc_ip_cluster3 nginx service cluster3" >&2

    attempt_counter=0
    max_attempts=0
    until $(kubectl --context=cluster2 exec -it ${netshoot_pod} -- curl --connect-timeout 3 --output /dev/null -m 30 --silent --head --fail ${nginx_svc_ip_cluster3}); do
        if [[ ${attempt_counter} -eq ${max_attempts} ]];then
          echo "NOT CONNECTED"
          return 1
        fi
        attempt_counter=$(($attempt_counter+1))
    done
    echo "CONNECTED"
}



test_connection


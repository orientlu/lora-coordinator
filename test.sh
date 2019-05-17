#!/bin/bash
# by orientlu

#set -x
version="t.0.0.1"

coordinator="lora-coo"
mqtt_brokers_name=(
mosquitto1
mosquitto2
)
gatewaybridge_name="bridge"
redis_name="redis_db"
network_name="test_coo"



if [[ -n "$1" && "$1" == "help" ]]; then
    echo -e "usage\n\t./test.sh  -- will run&clean\n\t./test.sh  -- run will run,not clean\n\t./test.sh  -- clean only clean\nuse BASH"
    exit 0
fi


if [[ -z "$1" || "$1" == "run" ]]; then

    echo -n "create test network "
    docker network create -d bridge ${network_name}
    docker network ls

    echo  "start mqtt"
    mqtt_url_list=""
    for((i=0; i<${#mqtt_brokers_name[*]};i++))
    do
        docker run --name ${mqtt_brokers_name[$i]} -d --network ${network_name}  -p 127.0.0.1:$((1883+i)):1883 ansi/mosquitto
        mqtt_url_list+="tcp://${mqtt_brokers_name[$i]}:1883,"
    done

    echo  "start redis"
    docker run --name ${redis_name} -d  --network ${network_name} -p 127.0.0.1:6379:6379 redis:3.0.7-alpine

    echo "start gateway-bridge"
    docker run --name ${gatewaybridge_name} -d --network ${network_name} -p 1700:1700/udp\
        --env "INTEGRATION.MQTT.AUTH.GENERIC.SERVER=tcp://${mqtt_brokers_name[0]}:1883"\
        orientludocker/lora-gateway-bridge:${version}

    #make docker

    docker run  -i --rm --name ${coordinator}\
        --network ${network_name}\
        --env "LORA_COORDINATOR_REDIS_URL=redis://${redis_name}:6379"\
        --env "LORA_COORDINATOR_MQTT_SERVERS=${mqtt_url_list}"\
        --env "LORA_COORDINATOR_GENERAL_LOG_LEVEL=6"\
        --mount type=bind,source=$(pwd),target=/etc/lora-coordinator\
        orientludocker/lora-coordinator:${version}
fi



if [[ -z "$1" || "$1" == "clean" ]]; then
    for((i=0; i<${#mqtt_brokers_name[*]};i++))
    do
        echo -n "stop "
        docker stop ${mqtt_brokers_name[$i]}
        echo -n "remove "
        docker rm ${mqtt_brokers_name[$i]}
    done

    echo -n "stop "
    docker stop ${redis_name}
    echo -n "remove "
    docker rm ${redis_name}

    echo -n "stop "
    docker stop ${gatewaybridge_name}
    echo -n "remove "
    docker rm ${gatewaybridge_name}


    echo -n "emove network "
    docker network rm ${network_name}
    docker network ls
fi


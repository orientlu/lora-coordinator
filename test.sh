#!/bin/bash
# by orientlu

version="0.0.1"
mqtt_brokers=(
mosquitto1
mosquitto2
)
redis=redis_db


if [[ -n "$1" && "$1" == "help" ]]; then
    echo -e "usage\n\t./test.sh  -- will run&clean\n\t./test.sh  -- run will run,not clean\n\t./test.sh  -- clean only clean\nuse BASH"
    exit 0
fi

if [[ -z "$1" || "$1" == "run" ]]; then
    mqtt_cmd=" "
    for((i=0; i<${#mqtt_brokers[*]};i++))
    do
        mqtt_cmd+="--link ${mqtt_brokers[$i]}:${mqtt_brokers[$i]} "
        docker run -d --name ${mqtt_brokers[$i]} -d  -p 127.0.0.1:$((1883+i)):1883 ansi/mosquitto
    done

    docker run -d --name ${redis} -d -p 127.0.0.1:6379:6379 redis:3.0.7-alpine

    #make docker

    docker run  -i --rm --name lor-coo ${mqtt_cmd} --link ${redis}:${redis} --mount type=bind,source=$(pwd),target=/etc/lora-coordinator lora-coordinator:${version}
fi


if [[ -z "$1" || "$1" == "clean" ]]; then
    for((i=0; i<${#mqtt_brokers[*]};i++))
    do
        echo -n "stop: "
        docker stop ${mqtt_brokers[$i]}
        echo -n "remove: "
        docker rm ${mqtt_brokers[$i]}
    done

    echo -n "stop: "
    docker stop ${redis}
    echo -n "remove: "
    docker rm ${redis}
fi


#!/bin/bash

pwd
for((i=1;i<=10;i++)); 
do  
  echo `date` $i >&2
  sleep 1
done

sleep 10
echo "finish"
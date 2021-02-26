#!/bin/bash

for i in {1..1000}; do curl -i http://$1:3000/data; sleep 1; done

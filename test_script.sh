#!/bin/bash

for i in {1..1000}; do curl -i http://10.0.1.248:3000/data; sleep 1; done

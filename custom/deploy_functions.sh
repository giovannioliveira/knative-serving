#!/bin/bash
for i in {0..423}
do
  kn service create "simtask-$i" --image docker.io/giovanniapsoliveira/temul --port 8080
  sleep 10
done

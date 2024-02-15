#!/bin/bash
for i in {0..423}
do
  kn service create "temul-$i" --image docker.io/giovanniapsoliveira/temul --port 8080
done
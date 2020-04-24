#!/bin/bash

for i in {1..10}; do
  go test ./app -run TestAppStateDeterminism -Enabled -Commit -NumBlocks=100 -BlockSize=200 -Seed ${i} -v -timeout 24h
done
#!/bin/bash
protoc -I=. --gogoslick_out=plugins=grpc:. ./_examples/normal_starting/messages.proto

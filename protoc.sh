protoc -I=. --gogoslick_out=plugins=grpc:. -I=./vendor ./rigger/protos.proto
protoc -I=. --gogoslick_out=plugins=grpc:.  ./messages.proto

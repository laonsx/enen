
default :
	go build

publish :
	./scripts/version.sh VERSION README.md main.go ./cmd/version.go
	git commit -a -m "publish version `cat VERSION`."
	echo "publish version `cat VERSION` success."

proto :
	protoc --go_out=./ ./common/pb/*.proto

run :
	./scripts/start.sh

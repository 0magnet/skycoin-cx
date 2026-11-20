
build:
	go build -o packageloader

test_loader:
	go test ./loader -v -coverprofile cover.out
	go tool cover -html=cover.out -o cover.html
	xdg-open cover.html

test_encoder:
	make build
	sudo rm -r ./encoder/test_bolt/
	./packageloader -load -name Test -path ./loader/test_folder/test_valid_program
	go test ./encoder -v -coverprofile cover.out
	go tool cover -html=cover.out -o cover.html
	xdg-open cover.html

test_tree:
	make build
	./packageloader -load -name TestTree -path ./tree/test_tree
	go test ./tree -v -coverprofile cover.out
	go tool cover -html=cover.out -o cover.html
	xdg-open cover.html

test_graph:
	make build
	./packageloader -load -name TestTree -path ./tree/test_tree
	go test ./graph -v -coverprofile cover.out
	go tool cover -html=cover.out -o cover.html
	xdg-open cover.html
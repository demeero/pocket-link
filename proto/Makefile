generate:
	docker build --target gen -t pocket-link-proto-gen .
	docker run --name pocket-link-proto-gen pocket-link-proto-gen
	docker cp pocket-link-proto-gen:/proto/gen/go ./gen
	docker rm pocket-link-proto-gen

lint:
	docker build --target lint -t pocket-link-proto-gen .
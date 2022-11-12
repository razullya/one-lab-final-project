protoc:
	protoc --proto_path=proto --go_out=. --go_opt=paths=import --go-grpc_out=. --go-grpc_opt=paths=import --experimental_allow_proto3_optional post.proto parse.proto

open:
	docker exec -it post_db psql -U postgres

start-postgre:
	docker run --name=post_db -e POSTGRES_PASSWORD='12345' -p 5436:5432 -d --rm postgres

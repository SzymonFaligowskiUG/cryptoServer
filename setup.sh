docker run --name scylla -d -p 8000:8000 scylladb/scylla:latest --alternator-port=8000 --alternator-write-isolation=always
go run cmd/seed/main.go
go run cmd/server/main.go &

build:
	go build -o ngx-collect cmd/ngx-collect/main.go

clean:
	rm -rf data/*
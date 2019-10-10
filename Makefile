install:
	go build -o bin/ansible-vault-run main.go
	mv bin/ansible-vault-run /usr/local/bin/

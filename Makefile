all:main
	./main

main:main.go
	go build -o main main.go

clean:
	rm -rf main



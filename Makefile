watch-styles:
	@echo "Watching styles..."
	npx tailwindcss -i ./src/input.css -o ./public/styles.css --watch

build-styles:
	@echo "Building styles..."
	npx tailwindcss -i ./src/input.css -o ./public/styles.css --minify

build:
	@echo "Building app..."
	go build -ldflags="-w -s"

run:
	go run main.go

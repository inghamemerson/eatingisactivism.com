watch-styles:
	@echo "Watching styles..."
	npx postcss --no-map ./src/input.css -o ./public/styles.css --watch --verbose

build-styles:
	@echo "Building styles..."
	npx postcss ./src/input.css -o ./public/styles.css --verbose --map

build:
	@echo "Building app..."
	go build -ldflags="-w -s"

run:
	go run main.go

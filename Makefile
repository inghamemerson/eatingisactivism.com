watch-css:
	@echo "Watching styles..."
	npx postcss --no-map ./src/input.css -o ./public/styles.css --watch --verbose

build-css:
	@echo "Building styles..."
	npx postcss ./src/input.css -o ./public/styles.css --verbose --map

watch-js:
	@echo "Watching scripts..."
	bun build --entrypoints ./src/main.js ./src/util.js --outdir ./public --watch

build-js:
	@echo "Building scripts..."
	bun build --entrypoints ./src/main.js ./src/util.js --outdir ./public --minify-whitespace --minify-syntax

build:
	@echo "Building app..."
	go build -ldflags="-w -s"

run:
	go run main.go

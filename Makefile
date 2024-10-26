# Makefile

BINARY_NAME=gostart

# Default url: http://localhost:7331

run: build 
	@./bin/main.exe

build:
	go mod tidy && \
   	templ generate && \
	go generate && \
	go build -ldflags="-w -s" -o ./bin/main.exe

dev/templ:
	templ generate --watch --proxy="http://localhost:7000" --open-browser=false -v

dev/tailwind:
	npx tailwindcss -i ./static/css/index.css -o static/css/output/styles.css --minify --watch

dev/air:
	air -c .air.toml

dev/esbuild:
	npx esbuild ./static/js/index.js --bundle --outdir=./static/js/output --minify --watch

dev/prettier:
	npx prettier . --write

dev: 
	make -j5 dev/templ dev/prettier dev/esbuild dev/tailwind dev/air

# prettier screws up the minification if last
# esbuild needs to be before tailwind to generate the proper classes, e.g. keeps generating spinner instead of dots even with correct classes
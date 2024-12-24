# Makefile

BINARY_NAME=gorant

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
	npx esbuild ./static/js/index.js ./static/js/sse.js ./static/js/post.js ./static/js/post-partial.js ./static/js/settings.js ./static/js/register-login.js ./static/js/htmx-bundle.js --bundle --outdir=./static/js/output --minify --watch

dev/prettier:
	npx prettier . --write ./static/js

# unused 
dev/biome:
	npx @biomejs/biome check --write ./static/js/

dev/keycloak:
# run keycloak and maildev in containers
#	docker run -p 8080:8080 -e KC_BOOTSTRAP_ADMIN_USERNAME=admin -e KC_BOOTSTRAP_ADMIN_PASSWORD=admin quay.io/keycloak/keycloak:26.0.7 start-dev
#	docker start 5b9d991aadc6087b4042a806626abe0be69a46efeca8381ec6617c79911dcf3f && \
#	docker pull maildev/maildev && docker run -p 1080:1080 -p 1025:1025 maildev/maildev
# maildev smtp server @ http://localhost:1025/ and gui @ http://localhost:1080/
	docker compose -f ./docker-compose-mail.yaml build && docker compose -f ./docker-compose-mail.yaml up

# used only when needed
dev/eslint:
	npx eslint

# prettier screws up the minification if last
# esbuild needs to be before tailwind to generate the proper classes, e.g. keeps generating spinner instead of dots even with correct classes
dev: 
	make -j6 dev/keycloak dev/templ dev/prettier dev/esbuild dev/tailwind dev/air 


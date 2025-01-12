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

dev/tailwind:
	npx tailwindcss -i ./static/css/index.css -o static/css/output/styles.css --minify --watch

# only difference here with the Dockerfile one is sourcemap 
dev/esbuild:
	npx esbuild ./static/js/admin/upload.ts ./static/js/index.ts ./static/js/sse.ts ./static/js/post.ts ./static/js/post-partial.ts ./static/js/settings.ts ./static/js/search.ts ./static/js/register-login.ts ./static/js/htmx-bundle.ts ./static/js/post-form.ts --bundle --sourcemap --outdir=./static/js/output --minify --watch

dev/templ:
	templ generate --watch --cmd="go run ." --proxy="http://localhost:7000" --open-browser=false -v

dev/prettier:
	npx prettier . --write ./static/js

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
	make -j4 dev/templ dev/prettier dev/esbuild dev/tailwind 

key: 
	make dev/keycloak 


############################
# Removed Wgo and Air - no need for those since Templ does reloading for all .go files
###########################

# dev/air:
# 	air -c .air.toml

# dev/wgo:
# 	wgo run . -xdir templates node_modules
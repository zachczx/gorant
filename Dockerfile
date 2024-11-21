FROM golang:1.23.3 AS first
ENV GO111MODULE=on
WORKDIR /app
COPY ./go.mod ./go.sum tailwind.config.js package.json package-lock.json ./
COPY ./posts ./posts
COPY ./templates ./templates
COPY ./database ./database
COPY ./users ./users
COPY ./static ./static
RUN go mod download

# Removed this command because the @latest one suffices. The version variable one looks more complicated than necessary.
# RUN go install github.com/a-h/templ/cmd/templ@$(go list -m -f '{{ .Version }}' github.com/a-h/templ)
# Technically this needn't be here if `templ generate` is done in dev, but no harm doing this. 
RUN go install github.com/a-h/templ/cmd/templ@latest

# Copy the source code. Note the slash at the end, as explained in
# https://docs.docker.com/reference/dockerfile/#copy
COPY *.go ./

# Env --- https://github.com/coollabsio/coolify/issues/1918
ARG LISTEN_ADDR
ENV LISTEN_ADDR={$LISTEN_ADDR} 

# Build
RUN templ generate && \
    CGO_ENABLED=0 GOOS=linux go build -o /app/gorant

####################################################################################

FROM node:22 AS second
WORKDIR /app
COPY --from=first /app/tailwind.config.js /app/package.json /app/gorant /app/package-lock.json /app/
COPY --from=first /app/templates /app/templates
COPY --from=first /app/static /app/static
# COPY package*.json ./
RUN npm install
RUN npx esbuild ./static/js/comment-form.js ./static/js/index.js --bundle --outdir=./static/js/output --minify &&\
    npx tailwindcss -i ./static/css/index.css -o static/css/output/styles.css --minify &&\
    npx brotli-cli compress --glob /app/static/css/output/styles.css /app/static/js/ext/htmx.min.js /app/static/js/output/comment-form.js /app/static/js/output/index.js

####################################################################################

FROM alpine:3.20.3
WORKDIR /app
COPY --from=second /app/gorant ./gorant
COPY --from=second /app/static ./static
ENV LISTEN_ADDR=${LISTEN_ADDR}
EXPOSE ${LISTEN_ADDR}

# Run
CMD ["/app/gorant"]


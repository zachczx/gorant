# FROM golang:1.23.4 AS first
FROM golang:alpine AS first
ENV GO111MODULE=on
# Changed to CGO_ENABLED=1 for C, so that I can run webp conversion. 
ENV CGO_ENABLED=1
RUN apk add build-base
WORKDIR /app
COPY ./go.mod ./go.sum tailwind.config.js package.json package-lock.json ./
COPY ./posts/ ./posts/
COPY ./templates/ ./templates/
COPY ./upload/ ./upload/
COPY ./database/ ./database/
COPY ./live/ ./live/
COPY ./users/ ./users/
COPY ./static/ ./static/
RUN go mod download

# Removed this command because the @latest one suffices. The version variable one looks more complicated than necessary.
# RUN go install github.com/a-h/templ/cmd/templ@$(go list -m -f '{{ .Version }}' github.com/a-h/templ)
# Technically this needn't be here if `templ generate` is done in dev, but no harm doing this. 
RUN go install github.com/a-h/templ/cmd/templ@v0.3.819

# Copy the source code. Note the slash at the end, as explained in
# https://docs.docker.com/reference/dockerfile/#copy
COPY *.go ./

# Env --- https://github.com/coollabsio/coolify/issues/1918
ARG LISTEN_ADDR
ENV LISTEN_ADDR={$LISTEN_ADDR} 

# Build
RUN templ generate && \
    GOOS=linux go build -o /app/gorant

####################################################################################

FROM node:22.12 AS second
WORKDIR /app
COPY --from=first /app/tailwind.config.js /app/package.json /app/gorant /app/package-lock.json /app/
COPY --from=first /app/templates /app/templates
COPY --from=first /app/static /app/static
# COPY package*.json ./
RUN npm install
RUN npx esbuild ./static/js/admin/upload.ts ./static/js/index.ts ./static/js/sse.ts ./static/js/post.ts ./static/js/post-partial.ts ./static/js/settings.ts ./static/js/register-login.ts ./static/js/htmx-bundle.ts ./static/js/tiptap.ts --bundle --outdir=./static/js/output --minify &&\       
    npx tailwindcss -i ./static/css/index.css -o static/css/output/styles.css --minify &&\
    npx brotli-cli compress --glob /app/static/css/output/styles.css /app/static/js/ext/htmx.min.js /app/static/js/output/comment-form.js /app/static/js/output/index.js

####################################################################################

# Only the builder requires golang:alpine (410mb+) vs alpine (50mb)
FROM alpine
WORKDIR /app
COPY --from=second /app/gorant ./gorant
COPY --from=second /app/static ./static
ENV LISTEN_ADDR=${LISTEN_ADDR}
EXPOSE ${LISTEN_ADDR}

# Run
CMD ["/app/gorant"]


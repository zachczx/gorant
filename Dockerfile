FROM golang:1.23.3 AS first
ENV GO111MODULE=on
WORKDIR /app
COPY ./go.mod ./go.sum tailwind.config.js package.json package-lock.json ./starter.db ./
COPY ./posts ./posts
COPY ./templates ./templates
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
    CGO_ENABLED=0 GOOS=linux go build -o /app/gostart

####################################################################################

FROM node:22 AS second
WORKDIR /app
COPY --from=first /app/tailwind.config.js /app/starter.db /app/package.json /app/gostart /app/package-lock.json /app/
COPY --from=first /app/templates /app/templates
COPY --from=first /app/static /app/static
# COPY package*.json ./
RUN npm install
RUN npx esbuild ./static/js/comment-form.js --bundle --outdir=./static/js/output --minify &&\
    npx tailwindcss -i ./static/css/index.css -o static/css/output/styles.css --minify

####################################################################################

FROM alpine:3.20.3
WORKDIR /app
COPY --from=second /app/gostart ./gostart
COPY --from=second /app/static ./static
COPY --from=second /app/starter.db ./starter.db
ENV LISTEN_ADDR=${LISTEN_ADDR}
EXPOSE ${LISTEN_ADDR}

# Run
CMD ["/app/gostart"]


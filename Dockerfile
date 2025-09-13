FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY server/ ./server/
RUN cd server && go build -o mole-server -ldflags="-w -s" .

FROM alpine:latest

# install ca-certificates, certbot with dns plugins, and dcron for ssl and cron jobs
RUN apk --no-cache add ca-certificates tzdata certbot dcron py3-pip \
    && pip3 install certbot-dns-cloudflare certbot-dns-route53 certbot-dns-digitalocean

WORKDIR /root/

# copy server binary and entrypoint script
COPY --from=builder /app/server/mole-server .
COPY entrypoint.sh .

# make entrypoint script executable
RUN chmod +x entrypoint.sh

# create directories for certificates and data
RUN mkdir -p /etc/letsencrypt /var/lib/letsencrypt /var/log/letsencrypt

# create cron directories
RUN mkdir -p /var/spool/cron/crontabs

# expose the port
EXPOSE 80

ENTRYPOINT ["./entrypoint.sh"]
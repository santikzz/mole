#!/bin/sh

# start crond in background
crond -b

# automatically set certificate paths based on domain
if [ "$MOLE_USE_HTTPS" = "true" ]; then
    export MOLE_CERT_FILE="/etc/letsencrypt/live/$MOLE_DOMAIN/fullchain.pem"
    export MOLE_KEY_FILE="/etc/letsencrypt/live/$MOLE_DOMAIN/privkey.pem"
fi

# check if certificates exist
if [ "$MOLE_USE_HTTPS" = "true" ] && [ ! -f "$MOLE_CERT_FILE" ]; then
    echo "generating ssl certificates for $MOLE_DOMAIN..."
    
    # generate certificates for main domain and wildcard
    certbot certonly --standalone \
        -d "$MOLE_DOMAIN" \
        -d "*.$MOLE_DOMAIN" \
        --email "$MOLE_EMAIL" \
        --agree-tos \
        --non-interactive \
        --expand
    
    if [ $? -eq 0 ]; then
        echo "certificates generated successfully"
    else
        echo "failed to generate certificates, starting without https"
        export MOLE_USE_HTTPS=false
    fi
fi

# setup certificate renewal cron job
if [ "$MOLE_USE_HTTPS" = "true" ]; then
    echo "0 12 * * * /usr/bin/certbot renew --quiet && echo 'certificates renewed'" > /var/spool/cron/crontabs/root
    echo "certificate renewal cron job added"
fi

# create log directory
mkdir -p /var/log

# start the mole server with logging
echo "Starting mole server with logging to /var/log/mole.log"
exec "./mole-server" 2>&1 | tee /var/log/mole.log
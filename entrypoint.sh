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
    
    # build domain list with main domain and subdomains
    DOMAINS="-d $MOLE_DOMAIN"
    
    # add specific subdomains from list
    if [ -n "$MOLE_SUBDOMAINS" ]; then
        IFS=','
        for subdomain in $MOLE_SUBDOMAINS; do
            DOMAINS="$DOMAINS -d $subdomain.$MOLE_DOMAIN"
        done
        unset IFS
    fi
    
    echo "requesting certificates for domains: $DOMAINS"
    echo "note: certbot will temporarily use port 80 for verification"
    
    # use standalone mode with specific port to avoid conflicts
    certbot certonly --standalone \
        --preferred-challenges http \
        --http-01-port 80 \
        $DOMAINS \
        --email "$MOLE_EMAIL" \
        --agree-tos \
        --non-interactive \
        --expand
    
    if [ $? -eq 0 ]; then
        echo "certificates generated successfully"
        echo "certificate files should be at: $MOLE_CERT_FILE"
        ls -la /etc/letsencrypt/live/$MOLE_DOMAIN/ || echo "certificate directory not found"
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
exec "./mole-server" --port 80 2>&1 | tee /var/log/mole.log
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
    
    # build domain list
    DOMAINS="-d $MOLE_DOMAIN"
    
    if [ "$MOLE_USE_WILDCARD" = "true" ]; then
        echo "using wildcard certificate with DNS challenge"
        DOMAINS="$DOMAINS -d *.$MOLE_DOMAIN"
        
        # create dns credentials file based on provider
        case "$MOLE_DNS_PROVIDER" in
            "cloudflare")
                echo "dns_cloudflare_email = $MOLE_DNS_EMAIL" > /root/.cloudflare.ini
                echo "dns_cloudflare_api_key = $MOLE_DNS_API_KEY" >> /root/.cloudflare.ini
                chmod 600 /root/.cloudflare.ini
                DNS_PLUGIN="--dns-cloudflare --dns-cloudflare-credentials /root/.cloudflare.ini"
                ;;
            "route53")
                DNS_PLUGIN="--dns-route53"
                ;;
            "digitalocean")
                echo "dns_digitalocean_token = $DO_AUTH_TOKEN" > /root/.digitalocean.ini
                chmod 600 /root/.digitalocean.ini
                DNS_PLUGIN="--dns-digitalocean --dns-digitalocean-credentials /root/.digitalocean.ini"
                ;;
            *)
                echo "unsupported dns provider: $MOLE_DNS_PROVIDER"
                echo "falling back to standalone mode without wildcard"
                DOMAINS="-d $MOLE_DOMAIN"
                DNS_PLUGIN="--standalone"
                ;;
        esac
        
        certbot certonly $DNS_PLUGIN \
            $DOMAINS \
            --email "$MOLE_EMAIL" \
            --agree-tos \
            --non-interactive \
            --expand
    else
        echo "using specific subdomains with HTTP challenge"
        
        # add specific subdomains from list
        if [ -n "$MOLE_SUBDOMAINS" ]; then
            IFS=','
            for subdomain in $MOLE_SUBDOMAINS; do
                DOMAINS="$DOMAINS -d $subdomain.$MOLE_DOMAIN"
            done
            unset IFS
        fi
        
        certbot certonly --standalone \
            $DOMAINS \
            --email "$MOLE_EMAIL" \
            --agree-tos \
            --non-interactive \
            --expand
    fi
    
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
exec "./mole-server" --port 80 2>&1 | tee /var/log/mole.log
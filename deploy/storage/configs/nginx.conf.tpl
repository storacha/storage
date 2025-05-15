server {
    listen 80;
    server_name ${app}.${domain};

    location / {
        return 301 https://$${host}$${request_uri};
    }
}

server {
    listen 443 ssl;
    server_name ${app}.${domain};

    ssl_certificate /etc/letsencrypt/live/${app}.${domain}/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/${app}.${domain}/privkey.pem;

    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_prefer_server_ciphers on;
    ssl_ciphers ECDHE-RSA-AES256-GCM-SHA512:DHE-RSA-AES256-GCM-SHA512:ECDHE-RSA-AES256-GCM-SHA384:DHE-RSA-AES256-GCM-SHA384:ECDHE-RSA-AES256-SHA384;
    ssl_session_timeout 1d;
    ssl_session_cache shared:SSL:10m;
    ssl_session_tickets off;
    ssl_stapling on;
    ssl_stapling_verify on;

    # Security headers
    add_header Strict-Transport-Security "max-age=63072000; includeSubDomains; preload";
    add_header X-Content-Type-Options nosniff;
    add_header X-Frame-Options DENY;
    add_header X-XSS-Protection "1; mode=block";

    location / {
        proxy_pass http://localhost:3000;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $${http_upgrade};
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $${host};
        proxy_set_header X-Real-IP $${remote_addr};
        proxy_set_header X-Forwarded-For $${proxy_add_x_forwarded_for};
        proxy_set_header X-Forwarded-Proto $${scheme};
        proxy_cache_bypass $${http_upgrade};
    }
}
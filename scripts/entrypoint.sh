#!/bin/sh

cat logo.txt
CMD=./mittag

if [ -n "$PUID" ] || [ -n "$PGID" ]; then
    USER=appuser
    HOME=/app

    if ! grep -q "$USER" /etc/passwd; then
        groupadd -g "$PGID" "$USER"
        useradd -d "$HOME" -g "$PGID" -M -N -u "$PUID" "$USER"
    fi

    chown "$USER":"$USER" "$HOME" -R
    printf "\nUID: %s GID: %s\n\n" "$PUID" "$PGID"
    exec su -c - "$USER" "$CMD"
else
    printf "\nWARNING: Running docker as root\n\n"
    exec "$CMD"
fi

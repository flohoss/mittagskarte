#!/bin/sh

cat logo.txt
CMD=./mittag

if [ -n "$PUID" ] || [ -n "$PGID" ]; then
    USER=appuser
    HOME=/app

    if ! grep -q "$USER" /etc/passwd; then
        # https://docs.fedoraproject.org/en-US/fedora/latest/system-administrators-guide/basic-system-configuration/Managing_Users_and_Groups/
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

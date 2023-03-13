#!/bin/sh

# Uninstall kobomail
rm -f /etc/udev/rules.d/97-kobomail.rules
rm -f /etc/ssl/certs/ca-certificates.crt
rm -rf /usr/local/kobomail/
rm -f /usr/local/Kobo/imageformats/libns.so
rm -rf /mnt/onboard/.adds/kobomail
rm -f /mnt/onboard/.adds/nm/kobomail

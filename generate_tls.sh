#!/bin/bash

mkdir ./.tls

sudo openssl genpkey -algorithm RSA -out ./.tls/yambol.key
sudo openssl req -new -x509 -key ./.tls/yambol.key -out ./.tls/yambol.crt -days 365

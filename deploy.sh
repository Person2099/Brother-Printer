#!/bin/bash
set -e

go build -o printer-app
sudo mv printer-app /usr/local/bin/
sudo cp -r /home/sys-infra/Brother-Printer/. /opt/printer-app/
sudo systemctl restart printer-app

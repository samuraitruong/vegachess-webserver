#!/bin/sh

# Exit immediately if a command exits with a non-zero status
set -e
git config --global user.name "Vega Publish"
git config --global user.email "vega@ci.com"

# Default environment variables if not provided
FTP_USER=${FTP_USER:-ftpuser}
FTP_PASS=${FTP_PASS:-password}

# Check if the FTP user exists; if not, create it
if ! id -u "$FTP_USER" >/dev/null 2>&1; then
  echo "Creating FTP user: $FTP_USER"
  useradd -m "$FTP_USER"
  mkdir -p /home/$FTP_USER && chown -R $FTP_USER:$FTP_USER /home/$FTP_USER
  chown $FTP_USER:$FTP_USER /data
fi

# Set the password for the FTP user
echo "$FTP_USER:$FTP_PASS" | chpasswd
echo "Password for user '$FTP_USER' set."

# Ensure the /data directory exists
if [ ! -d "/data" ]; then
  echo "Creating FTP data directory: /data"
  mkdir -p /data
fi

# Set proper permissions for the /data directory
echo "Setting permissions for /data"
chown "$FTP_USER":"$FTP_USER" /data
chmod 755 /data
#git clone git@github.com:Hobsons-Bay-Chess-Club-AU/tournaments.git
#ls
# Start the vsftpd service
echo "Starting vsftpd"
vsftpd /etc/vsftpd.conf & watcher

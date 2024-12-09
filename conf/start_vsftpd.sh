#!/bin/sh

#Remove all ftp users
grep '/ftp/' /etc/passwd | cut -d':' -f1 | xargs -r -n1 deluser

#Create users
#USERS='name1|password1|[folder1][|uid1] name2|password2|[folder2][|uid2]'
#may be:
# user|password foo|bar|/home/foo
#OR
# user|password|/home/user/dir|10000
#OR
# user|password||10000

#Default user 'ftp' with password 'alpineftp'

if [ -z "$USERS" ]; then
  USERS="alpineftp|alpineftp"
fi

echo "Set up user(s)"
for i in $USERS ; do
  NAME=$(echo $i | cut -d'|' -f1)
  GROUP=$NAME
  PASS=$(echo $i | cut -d'|' -f2)
  FOLDER=$(echo $i | cut -d'|' -f3)
  UID=$(echo $i | cut -d'|' -f4)

  if [ -z "$FOLDER" ]; then
    FOLDER="/ftp/$NAME"
  fi

  if [ ! -z "$UID" ]; then
    UID_OPT="-u $UID"
    #Check if the group with the same ID already exists
    GROUP=$(getent group $UID | cut -d: -f1)
    if [ ! -z "$GROUP" ]; then
      GROUP_OPT="-G $GROUP"
    fi
  fi

  echo "Add user"
  echo -e "$PASS\n$PASS" | adduser -h $FOLDER -s /sbin/nologin $UID_OPT $GROUP_OPT $NAME
  echo "Make their home folder"
  mkdir -p $FOLDER
  echo "Make sure they own that folder"
  chown $NAME:$GROUP $FOLDER
  unset NAME PASS FOLDER UID
done
echo "Created user(s)"

if [ -z "$MIN_PORT" ]; then
  MIN_PORT=21000
fi

if [ -z "$MAX_PORT" ]; then
  MAX_PORT=21010
fi

if [ ! -z "$ADDRESS" ]; then
  ADDR_OPT="-opasv_address=$ADDRESS"
fi

# Used to run custom commands inside container
if [ ! -z "$1" ]; then
  echo "Run custom command"
  exec "$@"
else
  echo "Start vsftpd"
  vsftpd -opasv_min_port=$MIN_PORT -opasv_max_port=$MAX_PORT $ADDR_OPT /etc/vsftpd/vsftpd.conf
  [ -d /var/run/vsftpd ] || mkdir /var/run/vsftpd
  pgrep vsftpd | tail -n 1 > /var/run/vsftpd/vsftpd.pid
  echo "Run pidproxy armed with the pid of vsftpd"
  exec pidproxy /var/run/vsftpd/vsftpd.pid true
  echo "Exited"
fi
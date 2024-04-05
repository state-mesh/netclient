#!/bin/bash

TOKEN="eyJzZXJ2ZXIiOiJhcGkuc2VsZi1zY2FsZS5jbHVzdGVyY2F0LmNvbSIsInZhbHVlIjoiQ0kzT1BFWVBXR0dNTzNETlVYTjM3M1BUSzc0NVZQS04ifQ=="
PORT=51820


run(){
   for i in {1..30}
   do
      echo "spining Up container ${i}"
      PORT=$((PORT+${i}))
      sudo docker run -d  --privileged -p ${PORT}:${PORT}/udp -e TOKEN=${TOKEN} -e HOST_NAME=netclient${i} -v /etc/netclient${i}:/etc/netclient --name netclient${i} abhi9686/netclient:NET-1082
      sleep 2
   done
}
cleanUp(){
   # remove all containers
   docker container rm $(docker container ls -aq) -f
   # delete config directories
   rm -rf /etc/netclient*
}

main(){

   while getopts :sr flag; do
	case "${flag}" in
   s)
      run
   ;;
   r)
      cleanUp
   ;;
   esac
done
}

main "${@}"
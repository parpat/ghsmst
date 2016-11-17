for ((  i = 1 ;  i <= $1;  i++  ))
do
#FOR SEPARATE TERMINALS:
#gnome-terminal -x sh -c "docker run -d --rm --name ghs$i -P -v /home/parth/gocode/src/github.com/parpat/ghsmst/logs:/logs ghsmst; bash"
#
docker run -d --name ghs$i -P -v /home/parth/gocode/src/github.com/parpat/ghsmst/logs:/logs ghsmst
#sleep 0.1
done

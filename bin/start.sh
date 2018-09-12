#!/bin/sh
# 
# leftTicket        Startup script for leftTicket
#
# Centos6.5/7 /etc/init.d/leftTicket
#
# chkconfig: 
# processname: leftTicket
# config: ./conf
# pidfile: ./bin/run/tsdbFacade.pid
# description: tsdbFacade is a reverse proxy server
#

# Source function library.
. /etc/init.d/functions

appdir=/spring.d/go/src/github.com/SpringDRen/leftTicket

name="$appdir/leftTicket"
prog=`/bin/basename $name`

mkdir -p $appdir/bin/run/
lockfile="$appdir/bin/run/$prog.lock"
pidfile="$appdir/bin/run/$prog.pid"
RETVAL=0

findproc() {
    pidof $prog
}

start() {
    echo -n $"Starting $prog: "

    daemon --pidfile=${pidfile} "${name} >> /dev/null &"
    RETVAL=$?
    sleep 1
    echo
    [ $RETVAL -eq 0 ] && (findproc > $pidfile  && touch $lockfile)
    return $RETVAL
}

stop() {
    echo -n $"Stopping $prog: "
    killproc -p ${pidfile} ${prog}
    RETVAL=$?
    echo
    [ $RETVAL = 0 ] && rm -f ${lockfile} ${pidfile}
}

rh_status() {
    #status -p ${pidfile} -b ${tsdbFacade} ${tsdbFacade}
    status -p $pidfile -l $lockfile $name
}


# See how we were called.
case "$1" in
    start)
        rh_status >/dev/null 2>&1 && exit 0
        start
        ;;
    stop)
        stop
        ;;
    status)
        rh_status
        RETVAL=$?
        ;;
    restart)
        stop
        start
        ;;
    *)
        echo $"Usage: $prog {start|stop|restart|status}"
        RETVAL=2
esac

exit $RETVAL

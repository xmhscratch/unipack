#!/bin/sh

if [ -t 1 ]; then
	COLCYAN="\033[36m"
	COLGREEN="\033[32m"
	COLRESET="\033[0m"
else
	COLCYAN=""
	COLGREEN=""
	COLRESET=""
fi

print_hello() {
    msg1=${1:-}
    msg2=${2:-}
    shift $#

    # printf "${COLGREEN}%s${COLRESET} ${COLCYAN}%s${COLRESET}\n" $msg1 $msg2;
    printf "%s%s%s %s%s%s\n" $COLGREEN $msg1 $COLRESET $COLCYAN $msg2 $COLRESET;
}

print_hello "olleh dlrow"

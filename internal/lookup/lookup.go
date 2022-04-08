package lookup

import (
	"log"
	"os/exec"
	"strings"
)

func ExtLookupUser(user string, lookup string, l *log.Logger) (u string, err error) {

	// todo: this whole package needs some love, quite some love
	switch lookup {
	case "GECOS":
		u, err = lookupGECOS(user, l)
	default:
		u = user
	}

	return u, err
}

func lookupGECOS(u string, l *log.Logger) (string, error) {

	out, err := exec.Command("/usr/bin/getent", "passwd", u).Output()
	if err != nil {
		return u, err
	}
	fields := strings.Split(string(out), ":")

	return fields[4], nil
}

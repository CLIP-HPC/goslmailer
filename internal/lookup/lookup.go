package lookup

import (
	"log"
	"os/exec"
	"strings"
)

func ExtLookupUser(user string, conn string) (u string) {

	switch conn {
	case "msteams":
		u = lookupGECOS(user)
	default:
		u = user
	}

	return u
}

func lookupGECOS(u string) string {

	out, err := exec.Command("/usr/bin/getent", "passwd", u).Output()
	if err != nil {
		log.Fatalln("Getent lookup FAILED. Aborting!")
	}
	fields := strings.Split(string(out), ":")

	return fields[4]
}

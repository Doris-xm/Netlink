package util

import (
	"regexp"
	"strconv"
	"strings"
)

func CheckInvalidIpv4(ip string) bool {
	// legal IP address format: 192.168.1.1/24
	re := regexp.MustCompile(`^([0-9]{1,3}\.){3}[0-9]{1,3}(/([8-9]|1[0-9]|2[0-9]|3[0-2]))?$`)

	if !re.MatchString(ip) {
		return false
	}

	ipParts := strings.Split(ip, "/")
	ipAddress := ipParts[0]

	// check each part of the IP address
	parts := strings.Split(ipAddress, ".")
	if len(parts) != 4 {
		return false
	}
	for _, part := range parts {
		if val, err := strconv.Atoi(part); err != nil || val < 0 || val > 255 {
			return false
		}
	}

	// We reserve the IP address range 192.168.10
	if parts[0] == "192" && parts[1] == "168" && parts[2] == "10" {
		return false
	}

	return true
}

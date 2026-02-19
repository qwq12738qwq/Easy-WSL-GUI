//go:build windows
// +build windows

package installWSL

import (
	"fmt"
	"strconv"
	"strings"
)

type WSLGroup struct {
	Name  string
	GID   int
	Users []string
}

func GetWSLUserGroups(distroName string) ([]WSLGroup, error) {
	if distroName == "" {
		return nil, fmt.Errorf("distro name is empty")
	}

	info := WSLinfo{
		Linux_Version: distroName,
	}

	out, err := Start_cmd(info, "UserGroups")
	if err != nil {
		return nil, err
	}

	content := Reduce_Unicode(out)
	if strings.TrimSpace(content) == "" {
		return []WSLGroup{}, nil
	}

	lines := strings.Split(content, "\n")
	groups := make([]WSLGroup, 0, len(lines))

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.Split(line, ":")
		if len(parts) < 3 {
			continue
		}

		name := parts[0]
		gidStr := parts[2]
		membersStr := ""
		if len(parts) >= 4 {
			membersStr = parts[3]
		}

		gid, err := strconv.Atoi(gidStr)
		if err != nil {
			continue
		}

		members := []string{}
		if membersStr != "" {
			for _, m := range strings.Split(membersStr, ",") {
				m = strings.TrimSpace(m)
				if m != "" {
					members = append(members, m)
				}
			}
		}

		groups = append(groups, WSLGroup{
			Name:  name,
			GID:   gid,
			Users: members,
		})
	}

	return groups, nil
}

func GetWSLUsers(distroName string) ([]string, error) {
	if distroName == "" {
		return nil, fmt.Errorf("distro name is empty")
	}

	info := WSLinfo{
		Linux_Version: distroName,
	}

	out, err := Start_cmd(info, "UserList")
	if err != nil {
		return nil, err
	}

	content := Reduce_Unicode(out)
	if strings.TrimSpace(content) == "" {
		return []string{}, nil
	}

	lines := strings.Split(content, "\n")
	var users []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// /etc/passwd format: username:password:UID:GID:gecos:home:shell
		parts := strings.Split(line, ":")
		if len(parts) >= 3 && parts[0] != "" {
			uid, err := strconv.Atoi(parts[2])
			if err == nil {
				// Filter out system users (UID < 1000) and nobody (UID 65534)
				// Usually regular users start from UID 1000
				if uid >= 1000 && uid != 65534 {
					users = append(users, parts[0])
				}
			}
		}
	}

	return users, nil
}

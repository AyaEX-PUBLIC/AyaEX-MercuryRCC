package main

// Hosted gameservers can run in a hardened container (Gameserver/Dockerfile,
// built on Docker Hardened Images) instead of as plain processes, when
// GAMESERVER_IMAGE is set. Renders stay native: they need the X display for
// OpenGL.

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	c "github.com/TwiN/go-color"
	. "github.com/tp-link-extender/RCCService/Shared"
)

// gameserverImage is the image hosted gameservers run in; empty runs them natively
func gameserverImage() string {
	return os.Getenv("GAMESERVER_IMAGE")
}

// containerised reports whether instances of a kind run in a container
func containerised(kind string) bool {
	return kind == "host" && gameserverImage() != "" && nativeRCC()
}

func envOr(name, fallback string) string {
	if v := os.Getenv(name); v != "" {
		return v
	}
	return fallback
}

func gameserverContainer(port int) string {
	return "mercury-gameserver-" + strconv.Itoa(port)
}

// checkGameserverImage warns at startup if the image hasn't been built.
// Hosting then fails rather than falling back to unconfined processes.
func checkGameserverImage() {
	image := gameserverImage()
	if image == "" {
		return
	}
	if err := exec.Command("docker", "image", "inspect", image).Run(); err != nil {
		Log(c.InRed("GAMESERVER_IMAGE " + image + " isn't built, so games can't start: run Gameserver/build.sh"))
		return
	}
	Log(c.InGreen("Hosted gameservers run in the " + image + " container image"))
}

// runContainer runs one hosted gameserver's Aya.Server in a hardened
// container until it exits: an unprivileged user, no capabilities or
// privilege escalation, a read-only filesystem, resource limits, and only
// Aya's build and resources (read-only) and the DataStore mounted. It shares
// the host network, as players reach it through the Orbiter on localhost
// and it calls the Site and this proxy on localhost.
func (i *rccInstance) runContainer() {
	exe, err := filepath.Abs(rccExe())
	if err != nil {
		Log(c.InRed("Failed to resolve RCC_EXE: " + err.Error()))
		return
	}
	datastore, err := filepath.Abs(envOr("GAMESERVER_DATASTORE_DIR", "../data/datastore"))
	if err == nil {
		err = os.MkdirAll(datastore, 0o700)
	}
	if err != nil {
		Log(c.InRed("Failed to prepare GAMESERVER_DATASTORE_DIR: " + err.Error()))
		return
	}

	name := gameserverContainer(i.port)
	args := []string{
		"run", "--rm", "--interactive", "--init",
		"--name", name, "--label", "mercury.gameserver=" + strconv.Itoa(i.port),
		"--network", "host",
		"--user", fmt.Sprintf("%d:%d", os.Getuid(), os.Getgid()),
		"--cap-drop", "ALL", "--security-opt", "no-new-privileges",
		"--read-only", "--tmpfs", "/tmp:rw,nosuid,nodev,size=256m", "--env", "HOME=/tmp",
		"--memory", envOr("GAMESERVER_MEMORY", "2g"),
		"--cpus", envOr("GAMESERVER_CPUS", "2"),
		"--pids-limit", envOr("GAMESERVER_PIDS", "512"),
		"--volume", filepath.Dir(exe) + ":/aya:ro",
		"--volume", datastore + ":/datastore",
	}

	extra := strings.Fields(envOr("RCC_ARGS", "--websocket --noconsole --insecure"))
	for n, arg := range extra {
		if arg == "--contentPath" && n+1 < len(extra) {
			// content/ and platform/ sit side by side; mounted at the same path
			resources := filepath.Dir(filepath.Clean(extra[n+1]))
			args = append(args, "--volume", resources+":"+resources+":ro")
		}
	}
	args = append(args, gameserverImage(), "--port", strconv.Itoa(i.port))
	args = append(args, extra...)
	args = append(args, "--datastorePath", "/datastore/Aya.DataStore.sqlite3")

	// a container left behind by a crash would still hold the name
	_ = exec.Command("docker", "rm", "--force", name).Run()

	cmd := exec.Command("docker", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	// with --noconsole, Aya.Server reads commands from stdin; an open pipe that's
	// never written to keeps that loop blocked instead of spinning on EOF
	stdin, err := cmd.StdinPipe()
	if err != nil {
		Log(c.InRed("Failed to create gameserver stdin pipe: " + err.Error()))
		return
	}
	defer stdin.Close()

	Log(c.InPurple(fmt.Sprintf("Starting host instance in container %s (%s)", name, gameserverImage())))
	if err := cmd.Run(); err != nil {
		fmt.Println(err)
	}
}

// retire stops a containerised instance once its games are over, so every
// game starts in a fresh container
func (p *pool) retire(i *rccInstance) {
	if !containerised(i.kind) || i.countJobs() > 0 {
		return
	}
	i.kill()
	go func() {
		_ = exec.Command("docker", "stop", "--time", "5", gameserverContainer(i.port)).Run()
	}()
}

// forget removes a stopped instance from the pool, freeing its port
func (p *pool) forget(i *rccInstance) {
	p.mu.Lock()
	defer p.mu.Unlock()
	for n, ins := range p.instances {
		if ins == i {
			p.instances = append(p.instances[:n], p.instances[n+1:]...)
			break
		}
	}
	delete(p.reserved, i.port)
}

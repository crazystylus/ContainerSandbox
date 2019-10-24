package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"syscall"
	"time"
)

// docker 			run	 <duration>	<cmd>	<params>
// go run main.go	run	<cmd>	<params>

func main() {
	switch os.Args[1] {
	case "run":
		run()
	case "child":
		child()
	default:
		panic("bad command")
	}
}

func run() {
	// Mount tempfs on /sys/fs/cgroup
	syscall.Mount("cgroup_root", "/sys/fs/cgroup", "tmpfs", 0, "")

	// Creating CGroup required mounts
	mounts := []string{"pids", "memory", "cpu,cpuacct", "blkio", "cpuset", "devices"}
	for _, mnt := range mounts {
		must(os.MkdirAll(filepath.Join("/sys/fs/cgroup/", mnt), os.ModePerm))
		must(syscall.Mount(mnt, filepath.Join("/sys/fs/cgroup/", mnt), "cgroup", 0, mnt))
	}

	fmt.Printf("Running %v as %d\n", os.Args[3:], os.Getpid())
	cmd := exec.Command("/proc/self/exe", append([]string{"child"}, os.Args[2:]...)...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS |
			syscall.CLONE_NEWPID |
			syscall.CLONE_NEWIPC |
			syscall.CLONE_NEWNS |
			syscall.CLONE_NEWNET,
		Unshareflags: syscall.CLONE_NEWNS,
	}
	cmd.Run()

	//Cleaning up cgroup mounts
	for _, mnt := range mounts {
		must(syscall.Unmount(filepath.Join("/sys/fs/cgroup/", mnt), 0))
		must(os.RemoveAll(filepath.Join("/sys/fs/cgroup/", mnt)))
	}
	must(syscall.Unmount("/sys/fs/cgroup", 0))
}

func child() {
	fmt.Printf("Running %v as %d\n", os.Args[3:], os.Getpid())

	cg()
	syscall.Sethostname([]byte("jail"))
	//syscall.Chroot("ubuntu")

	syscall.Mount("proc", "/proc", "proc", 0, "")
	syscall.Mount("tempfs", "/dev", "tempfs", syscall.MS_NOSUID|syscall.MS_STRICTATIME, "mode=755")
	syscall.Chdir("/home/sandbox")
	iTime, err := strconv.ParseInt(os.Args[2], 10, 64)
	if err != nil {
		fmt.Printf("\nTimeout invalid\n")
		return
	}
	// Handling String to int converstion for timeout

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(iTime)*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, os.Args[3], os.Args[4:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{}
	cmd.SysProcAttr.Credential = &syscall.Credential{Uid: 1000, Gid: 1000}
	cmd.Run()
	if ctx.Err() == context.DeadlineExceeded {
		fmt.Printf("\n Deadline Exceeded \n")
	}
	syscall.Unmount("/proc", 0)
	syscall.Unmount("/dev", 0)
	//syscall.Unmount("/sys/class/net/", 0)
}

//Temporary manual cgroup function
func cg() {
	cgroups := "/sys/fs/cgroup/"
	pids := filepath.Join(cgroups, "pids")
	os.Mkdir(filepath.Join(pids, "ourContainer"), 0755)
	ioutil.WriteFile(filepath.Join(pids, "ourContainer/pids.max"), []byte("10"), 0700)
	//Limiting max pids to 10

	ioutil.WriteFile(filepath.Join(pids, "ourContainer/notify_on_release"), []byte("1"), 0700)

	ioutil.WriteFile(filepath.Join(pids, "ourContainer/cgroup.procs"), []byte(strconv.Itoa(os.Getpid())), 0700)
	// up here we write container PIDs to cgroup.procs

	mems := filepath.Join(cgroups, "memory")
	memLimit := "175M"
	os.Mkdir(filepath.Join(mems, "ourContainer"), 0755)

	ioutil.WriteFile(filepath.Join(mems, "ourContainer/memory.limit_in_bytes"), []byte(memLimit), 0700)
	//Limiting total memory sum of the ps tree to memLimit
	ioutil.WriteFile(filepath.Join(mems, "ourContainer/notify_on_release"), []byte("1"), 0700)

	ioutil.WriteFile(filepath.Join(mems, "ourContainer/cgroup.procs"), []byte(strconv.Itoa(os.Getpid())), 0700)
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}

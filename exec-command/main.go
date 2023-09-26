package main

import (
	"fmt"
	"os"
	"os/exec"
)

func main() {
	args := []string{"-a"}
	cmd := exec.Command("/usr/bin/uname", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		panic(err)
	}
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case err := <-done:
		fmt.Println("received from done")
		if err != nil {
			panic(err)
		}
	}
	fmt.Println("command exec finished")
}

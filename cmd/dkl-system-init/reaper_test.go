package main

import (
	"os"
	"os/exec"
	"sync"
	"testing"
)

func _TestReap(t *testing.T) {
	truePath, err := exec.LookPath("true")
	if err != nil {
		t.Log("true binary not found, ignoring this test.")
		return
	}

	go handleChildren()

	count := 1000

	wg := &sync.WaitGroup{}
	wg.Add(count)
	for i := 0; i < count; i++ {
		i := i
		go func() {
			cmd := exec.Command(truePath)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			cmd.Stdin = os.Stdin
			if err := cmd.Run(); err != nil {
				t.Errorf("[%d] %v", i, err)
			}
			wg.Done()
		}()
	}

	wg.Wait()

	cmd := exec.Command("sh", "-c", "ps aux |grep Z")
	cmd.Stdout = os.Stdout
	cmd.Run()

	t.Fail()
}

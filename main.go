// Copyright 2017 The Transcode Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

var (
	source = flag.String("source", "", "source of file")
	mock   = flag.Bool("mock", true, "print what will be done")
)

func mkdir(name string) {
	_, err := os.Stat(name)
	if err == nil {
		return
	}

	fmt.Printf("mkdir %s\n", name)
	if *mock {
		return
	}

	err = os.Mkdir(name, 0775)
	if err != nil {
		panic(err)
	}
}

func transcode(source, destination string) {
	_, err := os.Stat(destination)
	if err == nil {
		return
	}

	name := "ffmpeg"
	args := []string{
		"-i", source, "-c:v", "libx264", "-preset", "slow", "-crf", "26", "-c:a",
		"mp3", destination,
	}
	command := name
	for _, arg := range args {
		command += " " + arg
	}
	fmt.Println(command)
	if *mock {
		return
	}

	cmd := exec.Command(name, args...)
	err = cmd.Run()
	if err != nil {
		panic(err)
	}
}

func copy(source, destination string) {
	_, err := os.Stat(destination)
	if err == nil {
		return
	}

	fmt.Printf("cp %s %s\n", source, destination)
	if *mock {
		return
	}

	sourceFile, err := os.Open(source)
	if err != nil {
		panic(err)
	}
	defer sourceFile.Close()

	stat, err := sourceFile.Stat()
	if err != nil {
		panic(err)
	}

	if !stat.Mode().IsRegular() {
		return
	}

	destinationFile, err := os.Create(destination)
	if err != nil {
		panic(err)
	}
	defer destinationFile.Close()

	_, err = io.Copy(destinationFile, sourceFile)
	if err != nil {
		panic(err)
	}
}

func process(source, destination string) {
	file, err := os.Open(source)
	if err != nil {
		panic(err)
	}

	entries, err := file.Readdir(8)
	for err == nil {
		for _, entry := range entries {
			currentSource := fmt.Sprintf("%s%c%s", source, os.PathSeparator, entry.Name())
			currentDestination := fmt.Sprintf("%s%c%s", destination, os.PathSeparator, entry.Name())
			if entry.IsDir() {
				mkdir(currentDestination)
				process(currentSource, currentDestination)
				continue
			}
			if strings.HasSuffix(entry.Name(), ".avi") {
				transcode(currentSource, currentDestination)
				continue
			}
			copy(currentSource, currentDestination)
		}
		entries, err = file.Readdir(8)
	}
}

func main() {
	flag.Parse()

	if *source == "" {
		fmt.Println("source required")
		return
	}

	process(strings.TrimSuffix(*source, string(os.PathSeparator)), ".")
}

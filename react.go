// Commandline tool for executing scripts on directory changes
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
// You should have received a copy of the GNU General Public License
// along with this program. If not, see http://www.gnu.org/licenses/.
//
package main

import (
    "bytes"
    "code.google.com/p/go.exp/fsnotify"
    "flag"
    "fmt"
    "log"
    "os" // For creating files and directories
    "os/exec"
    "os/signal"
    "path"
    "path/filepath" // For walking directories
    "strings"
)

type Driver struct {
    wtc          *fsnotify.Watcher
    script       string
    FoundScripts bool
    Running      bool
}

func NewDriver() (d *Driver) {
    d = &Driver{}
    return
}

func (d *Driver) Drive(script, root string) {
    d.script = script
    // Directory watcher
    d.wtc = d.createWatcher()
    go d.distributeIOEvents(d.wtc.Event)
    filepath.Walk(root, d.mapper())
}

func (d *Driver) Close() {
    d.wtc.Close()
}

// Fatal if unable to create a watcher
func (d *Driver) createWatcher() *fsnotify.Watcher {
    watcher, err := fsnotify.NewWatcher()
    if err != nil {
        log.Fatal(err)
    }
    return watcher
}

// Distributes delete, modify and create events
func (d *Driver) distributeIOEvents(eventstream chan *fsnotify.FileEvent) {
    for event := range eventstream {
        name := event.Name
        if d.Running {
            log.Printf("Skipping: %s", name)
        } else {
            d.Running = true
            defer d.done()
            script := filepath.Join(path.Dir(name), d.script)
            log.Printf("%s %s", script, name)
            cmd := exec.Command(script, name)
            cmd.Stdin = strings.NewReader("some input")
            var out bytes.Buffer
            cmd.Stdout = &out
            err := cmd.Run()
            if err != nil {
                log.Printf(err.Error())
            }
            fmt.Printf("%s", out.String())
        }
    }
}

func (d *Driver) done() {
    d.Running = false
}

func (d *Driver) mapper() filepath.WalkFunc {
    return func(path string, info os.FileInfo, err error) error {
        return d.MapFilesAndDirs(path, info, err)
    }
}

// MapFilesAndDirs connects a watcher to the path if a script is present
func (d *Driver) MapFilesAndDirs(path string, info os.FileInfo, err error) error {
    if err != nil {
        return err
    }
    stat, err := os.Stat(path)
    if err != nil {
        log.Fatal(err)
    }
    if stat.IsDir() {
        script := filepath.Join(path, d.script)
        _, err := os.Stat(script)
        if err != nil {
            if os.IsNotExist(err) {
                // skip
            }
        } else {
            d.watchPath(path, d.wtc)
        }
    }
    return nil
}

func (d *Driver) watchPath(path string, watcher *fsnotify.Watcher) {
    err := watcher.Watch(path)
    if err != nil {
        log.Fatal(err)
    }
    log.Printf("Watching %s\n", path)
    d.FoundScripts = true
}

var script = flag.String("script", ".onchange", "on change script name")
var root = flag.String("root", ".", "Directory to start reaction in")
var printVersion = flag.Bool("v", false, "Prints version and exits")

func main() {
    flag.Parse()
    if *printVersion {
        fmt.Printf("0.1\n")
        os.Exit(0)
    }
    driver := NewDriver()
    driver.Drive(*script, *root)
    defer driver.Close()
    if driver.FoundScripts {
        waitForCtrlC("Press Ctrl-c to stop")
    }
    fmt.Printf("No scripts named '%v' found in '%v' or it's sub folders!\n", *script, *root)
}

// Waits for Ctrl-c before calling os.Exit(0)
func waitForCtrlC(message string) {
    fmt.Printf("%s\n", message)
    interrupt := make(chan os.Signal, 1)
    signal.Notify(interrupt, os.Interrupt)
    for {
        _ = <-interrupt
        os.Exit(0)
    }
}

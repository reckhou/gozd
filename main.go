// Use of this source code is governed by a BSD-style license

package main

import (
	"flag"
  "os"
  "fmt"
  "io"
  "os/signal"
  "io/ioutil"
  "path/filepath"
  "syscall"
  "strconv"
  "crypto/sha1"
  
  //"./daemon"
)

// configuration
var (
   optSendSignal= flag.String("s","","send signal to a master process: stop, quit, reopen, reload")
   optConfigFile= flag.String("c","","set configuration file" )
	 optHelp= flag.Bool("h",false,"this help")
)

func parseConfigFile(filePath string) bool {
	return true
}

func usage() {
  fmt.Println("[command] -conf=[config file]")
  flag.PrintDefaults()
}

func readStringFromFile(filepath string) (string, error) {
  contents, err := ioutil.ReadFile(filepath)
  return string(contents), err
}

func writeStringToFile (filepath string, contents string) error {
  return ioutil.WriteFile(filepath, []byte(contents), 0x777)
}

func getPidByConf (confPath string, prefix string) (int, error) {
  
  confPath,err := filepath.Abs(confPath)
  if (err != nil) {
    return 0, err
  }
  
  hashSha1 := sha1.New()
  io.WriteString(hashSha1, confPath)
  pidFilepath := filepath.Join(os.TempDir(), fmt.Sprintf("%v_%x.pid", prefix, hashSha1.Sum(nil)))
  
  pidString, err := readStringFromFile(pidFilepath)
  if (err != nil) {
    return 0, err
  }
  
  return strconv.Atoi(pidString)
}

func main() {
  // parse arguments
  flag.Parse()

  // -conf parse config
  if (!parseConfigFile(*optConfigFile)) {
    usage()
    os.Exit(1)
    return
  }

  // find master process id by conf
  pid,err := getPidByConf(*optConfigFile, "gozerodown")
  if (err != nil) {
    pid = 0 
  }
  
  // -s send signal to the process that has same config
  switch (*optSendSignal) {
    case "stop": 
    if (pid != 0) {
      p,err := os.FindProcess(pid)
      if (err == nil) {
        p.Signal(syscall.SIGTERM)
        // wait it end
        os.Exit(0)
      }
    }
    os.Exit(0)
    
    case "start","":
      // start daemon
    case "reopen","reload":
      if (pid != 0) {
        p,err := os.FindProcess(pid)
        if (err == nil) {
          p.Signal(syscall.SIGTERM)
          // wait it end
          // start daemon
        }
      }
      os.Exit(0)
  }

  // handle signals
  // Set up channel on which to send signal notifications.
  // We must use a buffered channel or risk missing the signal
  // if we're not ready to receive when the signal is sent.
  cSignal := make(chan os.Signal, 1)
  signal.Notify(cSignal, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGUSR2, syscall.SIGINT)

  // Block until a signal is received.
  for s := range cSignal {
    fmt.Println("Got signal:", s)
    switch (s) {
      case syscall.SIGHUP, syscall.SIGUSR2:
        // upgrade, reopen
      case syscall.SIGTERM, syscall.SIGINT:
        // quit
        os.Exit(0)
    }
  }
}
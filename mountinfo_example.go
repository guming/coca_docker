package main

import (
	"strings"
	"os"
	"bufio"
	"fmt"
)
func main(){
	subsystem:="memory"
	f, err := os.Open("/proc/self/mountinfo")
	if err != nil {
		fmt.Errorf("error is %v",err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		txt := scanner.Text()
		fmt.Println(txt)
		fields := strings.Split(txt, " ")
		for _, opt := range strings.Split(fields[len(fields)-1], ",") {
			if opt == subsystem {
				fmt.Printf(fields[4])
			}
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Errorf("error is %v",err)
	}

}


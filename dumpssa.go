package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func dumpSSA(platform string, before, after commit, fnname string) {
	fmt.Printf("dumping SSA for %v:\n", fnname)

	// make fnname into an easier to deal with filename
	filename := strings.ReplaceAll(fnname, "(", "_")
	filename = strings.ReplaceAll(filename, ")", "_")
	filename = strings.ReplaceAll(filename, "*", ".")

	for _, c := range []commit{before, after} {
		cmdgo := filepath.Join(c.dir, "bin", "go")
		cmd := exec.Command(cmdgo, "build", "std", "cmd")
		goos, goarch := parsePlatform(platform)
		cmd.Env = append(os.Environ(), "GOOS="+goos, "GOARCH="+goarch, "GOSSAFUNC="+fnname)
		cmd.Dir = filepath.Join(c.dir, "src")
		pipe, err := cmd.StderrPipe()
		check(err)
		err = cmd.Start()
		check(err)

		scan := bufio.NewScanner(pipe)
		for scan.Scan() {
			s := scan.Text()
			if len(s) == 0 {
				continue
			}
			const dumpedSSATo = "dumped SSA to "
			if strings.HasPrefix(s, dumpedSSATo) {
				relpath := s[len(dumpedSSATo):]
				if !strings.HasSuffix(relpath, "ssa.html") {
					panic("wrote ssa to non-ssa.html file")
				}
				src := filepath.Join(c.dir, "src", relpath)
				prefix := ""
				if platform != "" {
					prefix = fmt.Sprintf("%s_%s_", goos, goarch)
				}
				dst := filepath.Join(c.dir, "src", strings.TrimSuffix(relpath, "ssa.html")+prefix+filename+".html")
				err = os.Rename(src, dst)
				check(err)
				fmt.Println(dst)
			}
		}
		check(scan.Err())

		err = cmd.Wait()
		check(err)
	}
	fmt.Println()
}

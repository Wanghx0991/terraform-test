package util

import (
	"bytes"
	"os/exec"
)

func DoCmd(command string) (string, string, error) {
	//函数返回一个*Cmd，用于使用给出的参数执行name指定的程序
	Cmd := exec.Command("/bin/bash", "-c", command)
	var out, Err bytes.Buffer
	Cmd.Stdout = &out
	Cmd.Stderr = &Err
	//Run执行c包含的命令，并阻塞直到完成。  这里stdout被取出，cmd.Wait()无法正确获取stdin,stdout,stderr，则阻塞
	err := Cmd.Run()
	return out.String(), Err.String(), err
}

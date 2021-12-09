package common

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"os/exec"
	"strings"
	"sync/atomic"
	"time"
)

const DefaultHangTimeout = 20 * 60

type ExternalProgramExecutor struct {
	execName   string
	arguments  *ExternalProgramArguments

	HangTimeout int64

	stdErrPipe io.ReadCloser
	stdOutPipe io.ReadCloser
	stdInPipe  io.WriteCloser
	process    *exec.Cmd

	done chan error

	latestTime int64
}

func NewExternalProgramExecutor(execName string, inputArgs []string, outputArgs []string) *ExternalProgramExecutor {
	return &ExternalProgramExecutor{
		execName: execName,
		arguments: &ExternalProgramArguments{
			InputArgs:  inputArgs,
			OutputArgs: outputArgs,
		},
		HangTimeout: DefaultHangTimeout,
	}
}

func (e *ExternalProgramExecutor) GetStdErrPipe() io.ReadCloser {
	return e.stdErrPipe
}

func (e *ExternalProgramExecutor) GetStdOutPipe() io.ReadCloser {
	return e.stdOutPipe
}

func (e *ExternalProgramExecutor) GetStdInPipe() io.WriteCloser {
	return e.stdInPipe
}

func (e *ExternalProgramExecutor) GetExitCode() int {
	return e.process.ProcessState.ExitCode()
}

func (e *ExternalProgramExecutor) GetProcess() *exec.Cmd {
	return e.process
}

func (e *ExternalProgramExecutor) ExecuteSynchronously() (string, error) {

	execBin, err := findExecPath(e.execName)
	if err != nil {
		log.Printf("external program (%v) not available (not installed) : %v\n", e.execName, err.Error())
		return "", err
	}

	// set command to exec execution
	command := e.arguments.getArguments()
	log.Printf("exeternal program command : %v %v", execBin, command)

	var outputBuffer, errorBuffer bytes.Buffer
	e.process = exec.Command(execBin, command...)
	e.process.Stdout = &outputBuffer
	e.process.Stderr = &errorBuffer

	// exec execution
	if err := e.process.Run(); err != nil {
		log.Printf("%v exec error, err=%v", e.execName, err)
		return "", err
	}

	return outputBuffer.String(), nil
}

func (e *ExternalProgramExecutor) ExecuteAsynchronously() <-chan error {

	e.done = make(chan error)

	execBin, err := findExecPath(e.execName)
	if err != nil {
		log.Printf("external program (%v) not available (not installed) : %v", e.execName, err.Error())
	}

	// set latestTime with current time
	atomic.StoreInt64(&e.latestTime, time.Now().Unix())

	// set command to exec ffmpeg
	command := e.arguments.getArguments()
	log.Printf("external program command : %v %v", execBin, command)

	proc := exec.Command(execBin, command...)
	errStream, err := proc.StderrPipe()
	if err != nil {
		log.Println("stderr not available : ", err.Error())
	} else {
		e.stdErrPipe = errStream
	}

	stdOut, err := proc.StdoutPipe()
	if err != nil {
		log.Println("stdout not available : ", err.Error())
	} else {
		e.stdOutPipe = stdOut
	}

	stdIn, err := proc.StdinPipe()
	if err != nil {
		log.Println("stdin not available : ", err.Error())
	} else {
		e.stdInPipe = stdIn
	}

	// exec fmpeg
	err = proc.Start()
	e.process = proc
	log.Printf("after proc start")

	go func(err error) {

		defer func() {
			close(e.done)
		}()

		if err != nil {
			e.done <- fmt.Errorf("execution start fail (cmd:%v, err:%v)", command, err)
			return
		}

		// on timer (10 minutes...)
		done := e.OnTimeout()

		defer func() {
			close(done)
		}()

		log.Printf("begin to wait....")
		err = proc.Wait()
		if err != nil {
			err = fmt.Errorf("failed finish execution. (cmd=%v) %w", command, err)
		}
		log.Printf("finish to wait....")

		done <- err
		e.done <- err

	}(err)

	return e.done
}

func (e *ExternalProgramExecutor) OnTimeout() chan error{

	done := make(chan error)

	go func() {
		for {
			select {
			case err, _ := <-done:
				log.Printf("timeout checker ended.. (err=%v)", err)
				return
			default:
				currentTime := time.Now().Unix()
				latestTime := atomic.LoadInt64(&e.latestTime)
				if latestTime + e.HangTimeout <= currentTime {
					err := e.GetProcess().Process.Kill()
					if err != nil {
						log.Printf("external program (%v) termination waring (err=%v)\n", e.execName, err)
					}
					log.Printf("external program (%v) timed out", e.execName)
					e.done <- ErrSigTerm.Copy(err)
					return
				} else {
					time.Sleep(time.Second)
				}
			}
		}
	}()

	return done
}

func (e *ExternalProgramExecutor) OnProgress(splitFunction func([]byte, bool) (int, []byte, error),
	scanFunction func(*bufio.Scanner, chan interface{}, interface{}), pipe io.ReadCloser, args interface{}) <-chan interface{} {

	progressOutput := make(chan interface{})

	go func() {
		defer func() {
			close(progressOutput)
		}()

		if pipe == nil || scanFunction == nil {
			progressOutput <- nil
			return
		}
		defer func(pipe io.ReadCloser) {
			_ = pipe.Close()
		}(pipe)

		scanner := bufio.NewScanner(pipe)
		scanner.Split(splitFunction)
		scanner.Buffer(make([]byte, 2), bufio.MaxScanTokenSize)

		for scanner.Scan() {
			// update latest time
			atomic.StoreInt64(&e.latestTime, time.Now().Unix())
			// run parsing logics
			scanFunction(scanner, progressOutput, args)
		}
	}()

	return progressOutput
}

func (e *ExternalProgramExecutor) OnMessages(splitFunction func([]byte, bool) (int, []byte, error),
	scanFunction func(*bufio.Scanner, chan string, interface{}), pipe io.ReadCloser, args interface{}) <-chan string {

	messageOutput := make(chan string)

	go func() {
		defer func() {
			close(messageOutput)
		}()

		if pipe == nil || scanFunction == nil{
			messageOutput <- ""
			return
		}
		defer func(pipe io.ReadCloser) {
			_ = pipe.Close()
		}(pipe)

		scanner := bufio.NewScanner(pipe)
		scanner.Split(splitFunction)
		scanner.Buffer(make([]byte, 2), bufio.MaxScanTokenSize)

		for scanner.Scan() {
			scanFunction(scanner, messageOutput, args)
		}
	}()

	return messageOutput
}

type ExternalProgramArguments struct {
	InputArgs  []string
	OutputArgs []string
}

func (e *ExternalProgramArguments) getArguments() []string {
	var args []string
	args = append(args, e.InputArgs...)
	args = append(args, e.OutputArgs...)
	return args
}

func findExecPath(execName string) (string, error) {

	// exec cmd
	command := exec.Command(GetWhereIs(), execName)

	var output bytes.Buffer
	command.Stdout = &output

	if err := command.Run(); err != nil {
		return "", err
	}

	return strings.Trim(output.String(), "\n"), nil
}

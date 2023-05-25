package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

var db = make(map[string]string)

func getCommandAndArgs(input string) (command string, args []string) {
	splited := strings.Split(strings.TrimSpace(string(input)), " ")
	return strings.ToUpper(splited[0]), splited[1:]
}

func validateArity(command string, args []string, arity int) error {
	if len(args) != arity {
		fmt.Printf("Wrong number of arguments for %s\n", command)
		return fmt.Errorf("Wrong number of arguments for %s", command)
	}
	return nil
}

func handleConnection(c net.Conn) {
	fmt.Printf("Serving %s\n", c.RemoteAddr().String())

	for {
		netData, err := bufio.NewReader(c).ReadString('\n')
		if err != nil {
			fmt.Println(err)
			return
		}

		switch command, args := getCommandAndArgs(netData); command {
		case "STOP":
			fmt.Println("Exiting TCP server!")
			c.Close()
			return
		case "PING":
			c.Write([]byte("PONG\n"))
		case "ECHO":
			c.Write([]byte("echo: \"" + strings.Join(args, " ") + "\"\n"))
		case "SET":
			err := validateArity(command, args, 2)
			if err != nil {
				c.Write([]byte(err.Error() + "\n"))
				continue
			}
			db[args[0]] = args[1]
			c.Write([]byte("OK\n"))
		case "GET":
			err := validateArity(command, args, 1)
			if err != nil {
				c.Write([]byte(err.Error() + "\n"))
				continue
			}
			c.Write([]byte(db[args[0]] + "\n"))
		default:
			fmt.Println("Unknown command: ", command)
			c.Write([]byte("Unknown command: " + command + "\n"))
		}
	}
}

func main() {
	arguments := os.Args
	if len(arguments) == 1 {
		fmt.Println("Please provide port number")
		return
	}

	PORT := ":" + arguments[1]
	l, err := net.Listen("tcp", PORT)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer l.Close()

	for {
		c, err := l.Accept()
		if err != nil {
			fmt.Println(err)
			return
		}

		sigc := make(chan os.Signal, 1)
		signal.Notify(sigc,
			syscall.SIGHUP,
			syscall.SIGINT,
			syscall.SIGTERM,
			syscall.SIGQUIT)
		go func() {
			fmt.Printf("Got signal: %v\n", <-sigc)
			c.Close()
			os.Exit(0)
		}()

		go handleConnection(c)
	}
}

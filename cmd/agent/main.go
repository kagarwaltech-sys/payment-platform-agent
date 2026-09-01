package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"payment-platform-agent/internal/agent"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	command := os.Getenv("MCP_SERVER_COMMAND")
	if command == "" {
		command = "go"
	}
	args := strings.Fields(os.Getenv("MCP_SERVER_ARGS"))
	if len(args) == 0 {
		args = []string{"run", "../payment-platform-mcp/cmd/server"}
	}

	runner, err := agent.Connect(ctx, command, args...)
	if err != nil {
		fatal(err)
	}
	defer runner.Close()

	tools, err := runner.Tools(ctx)
	if err != nil {
		fatal(err)
	}
	fmt.Printf("Connected to MCP server; tools: %d\n", len(tools))
	fmt.Println("Commands: get <payment-id> | pay <amount> <currency> <reference> | quit")

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}
		parts := strings.Fields(scanner.Text())
		if len(parts) == 0 {
			continue
		}
		switch parts[0] {
		case "quit", "exit":
			return
		case "get":
			if len(parts) != 2 {
				fmt.Println("usage: get <payment-id>")
				continue
			}
			data, err := runner.GetPayment(ctx, parts[1])
			if err != nil {
				fmt.Println("error:", err)
				continue
			}
			fmt.Println(agent.FormatPayment(data))
		case "pay":
			if len(parts) != 4 {
				fmt.Println("usage: pay <amount> <currency> <reference>")
				continue
			}
			fmt.Printf("Confirm payment of %s minor units %s for %s? (yes/no): ", parts[1], parts[2], parts[3])
			if !scanner.Scan() || strings.ToLower(strings.TrimSpace(scanner.Text())) != "yes" {
				fmt.Println("payment cancelled")
				continue
			}
			var amount int64
			if _, err := fmt.Sscan(parts[1], &amount); err != nil {
				fmt.Println("invalid amount")
				continue
			}
			data, err := runner.PayPayment(ctx, amount, parts[2], parts[3], true)
			if err != nil {
				fmt.Println("error:", err)
				continue
			}
			fmt.Println(agent.FormatPayment(data))
		default:
			fmt.Println("unknown command")
		}
	}
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}

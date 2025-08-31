package client

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/obsidian-agent-cli/internal/constant"
	"github.com/obsidian-agent-cli/internal/property"
	"github.com/obsidian-agent-cli/internal/proto"
	"github.com/obsidian-agent-cli/internal/utils"
)

func Run(cfg *property.Config) {
	cli, err := NewWSClient(cfg.URL, cfg.Token)
	if err != nil {
		log.Fatal(err)
	}
	defer cli.Close()

	fmt.Println("输入你的问题；指令：/reset 清空历史，/sys <prompt> 设置system，/exit 退出")

	// 会话状态
	var system string
	history := make([]proto.ChatMessage, 0, 32)

	// Ctrl+C 优雅退出
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	go func() { <-sig; fmt.Println("\nbye"); _ = cli.Close(); os.Exit(0) }()

	// REPL
	in := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("> ")
		if !in.Scan() {
			break
		}
		line := strings.TrimSpace(in.Text())
		if line == "" {
			continue
		}
		// 命令
		if strings.HasPrefix(line, "/exit") {
			break
		}
		if strings.HasPrefix(line, "/reset") {
			history = history[:0]
			system = ""
			fmt.Println(constant.COLOR_GRAY + "[reset] 已清空 system 与历史" + constant.COLOR_RESET)
			continue
		}
		if strings.HasPrefix(line, "/sys ") {
			system = strings.TrimSpace(strings.TrimPrefix(line, "/sys "))
			fmt.Println(constant.COLOR_GRAY + "[system] " + system + constant.COLOR_RESET)
			continue
		}

		// 组装多轮 messages
		msgs := make([]proto.ChatMessage, 0, len(history)+2)
		if strings.TrimSpace(system) != "" {
			msgs = append(msgs, proto.ChatMessage{Role: "system", Content: system})
		}
		msgs = append(msgs, history...)
		msgs = append(msgs, proto.ChatMessage{Role: "user", Content: line})

		reqID := "run-" + utils.RandID()
		req := proto.MsgRequest{
			Type:       "agent/run",
			ID:         reqID,
			Intent:     cfg.Intent,
			Question:   line,     // 兼容服务端旧版
			Messages:   msgs,     // 新：把历史发给服务端（若支持）
			Reserve:    cfg.Reserve,
			AllowTools: cfg.AllowTools,
		}

		// 发送本轮
		if err := cli.SendJSON(req); err != nil {
			fmt.Println(constant.COLOR_RED, "[send error]", err, constant.COLOR_RESET)
			continue
		}

		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.TimeoutSec)*time.Second)
		defer cancel()

		assistantBuf := &strings.Builder{}
		previewShown := false
		turnDone := make(chan struct{})

		go func() {
			defer close(turnDone)
			for {
				select {
				case <-ctx.Done():
					_ = cli.SendJSON(proto.MsgRequest{Type: "agent/cancel", ID: reqID})
					return
				default:
				}
				var m proto.MsgResponse
				if err := cli.ReadOne(&m); err != nil {
					return
				}
				switch m.Type {
				case "agent/preview.delta":
					if !previewShown {
						previewShown = true
						fmt.Printf("%s[preview]%s ", constant.COLOR_GRAY, constant.COLOR_RESET)
					}
					fmt.Print(strings.ReplaceAll(m.Text, "\n", " "))
				case "agent/full.delta":
					if previewShown && assistantBuf.Len() == 0 {
						fmt.Print("\n" + strings.Repeat("-", 72) + "\n")
						fmt.Printf("%s[assistant]%s ", constant.COLOR_CYAN, constant.COLOR_RESET)
					}
					if cfg.ShowSeq {
						fmt.Printf("%s{%d}%s", constant.COLOR_GRAY, m.Seq, constant.COLOR_RESET)
					}
					assistantBuf.WriteString(m.Text)
					fmt.Print(m.Text)
				case "tools/call.delta":
					fmt.Printf("\n%s[tool]%s %s\n", constant.COLOR_GRAY, constant.COLOR_RESET, m.Text)
				case "tools/call.result":
					js, _ := utils.JsonIndent(m.Result)
					fmt.Printf("\n%s[tool.result]%s\n%s\n", constant.COLOR_GRAY, constant.COLOR_RESET, js)
				case "agent/error":
					fmt.Printf("\n%s[error]%s %s (%s)\n", constant.COLOR_RED, constant.COLOR_RESET, m.ErrorMsg, m.ErrorCode)
					return
				case "agent/done":
					fmt.Println()
					return
				}
			}
		}()

		<-turnDone

		// 写回历史
		ans := assistantBuf.String()
		if ans != "" {
			history = append(history, proto.ChatMessage{Role: "user", Content: line})
			history = append(history, proto.ChatMessage{Role: "assistant", Content: ans})
		}
	}

	fmt.Println("done.")
}

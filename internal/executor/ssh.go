package executor

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/xiaolfeng/builder-cli/internal/config"
	"golang.org/x/crypto/ssh"
)

// SSHExecutor SSH 远程执行器
type SSHExecutor struct {
	*BaseExecutor
	host        string
	port        int
	username    string
	authMethod  ssh.AuthMethod
	commands    []string
	script      string
	localScript string
}

// NewSSHExecutor 创建 SSH 执行器
func NewSSHExecutor(taskName string, cfg config.TaskConfig, server *config.Server) (*SSHExecutor, error) {
	e := &SSHExecutor{
		BaseExecutor: NewBaseExecutor(taskName, TypeSSH),
		host:         server.Host,
		port:         server.Port,
		username:     server.Username,
		commands:     cfg.Commands,
		script:       cfg.Script,
		localScript:  cfg.LocalScript,
	}

	// 默认端口
	if e.port == 0 {
		e.port = 22
	}

	// 设置认证方式
	auth, err := e.createAuthMethod(server.Auth)
	if err != nil {
		return nil, fmt.Errorf("创建 SSH 认证失败: %w", err)
	}
	e.authMethod = auth

	// 设置超时
	if cfg.Timeout > 0 {
		e.SetTimeout(time.Duration(cfg.Timeout) * time.Second)
	} else {
		e.SetTimeout(10 * time.Minute) // SSH 默认 10 分钟
	}

	return e, nil
}

// createAuthMethod 创建 SSH 认证方式
func (e *SSHExecutor) createAuthMethod(auth config.ServerAuth) (ssh.AuthMethod, error) {
	switch auth.Type {
	case "password":
		return ssh.Password(auth.Password), nil

	case "key":
		keyPath := expandHomePath(auth.KeyPath)
		key, err := os.ReadFile(keyPath)
		if err != nil {
			return nil, fmt.Errorf("读取密钥文件失败: %w", err)
		}

		signer, err := ssh.ParsePrivateKey(key)
		if err != nil {
			return nil, fmt.Errorf("解析密钥失败: %w", err)
		}

		return ssh.PublicKeys(signer), nil

	default:
		return nil, fmt.Errorf("不支持的认证类型: %s", auth.Type)
	}
}

// Execute 执行 SSH 命令
func (e *SSHExecutor) Execute(ctx context.Context, handler OutputHandler) error {
	handler(fmt.Sprintf("🔗 连接服务器: %s@%s:%d", e.username, e.host, e.port), false)

	// 创建 SSH 配置
	sshConfig := &ssh.ClientConfig{
		User:            e.username,
		Auth:            []ssh.AuthMethod{e.authMethod},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // 注意：生产环境应验证主机密钥
		Timeout:         30 * time.Second,
	}

	// 连接服务器
	addr := fmt.Sprintf("%s:%d", e.host, e.port)
	client, err := ssh.Dial("tcp", addr, sshConfig)
	if err != nil {
		return fmt.Errorf("SSH 连接失败: %w", err)
	}
	defer client.Close()

	handler("✅ SSH 连接成功", false)
	handler("", false)

	// 根据配置执行不同的操作
	if e.localScript != "" {
		return e.executeLocalScript(ctx, client, handler)
	}

	if e.script != "" {
		return e.executeRemoteScript(ctx, client, handler)
	}

	return e.executeCommands(ctx, client, handler)
}

// executeCommands 执行命令列表
func (e *SSHExecutor) executeCommands(ctx context.Context, client *ssh.Client, handler OutputHandler) error {
	// 将所有命令合并为一条，用 && 连接
	// 这样可以保持工作目录等状态在命令之间传递
	if len(e.commands) == 0 {
		return nil
	}

	// 显示将要执行的命令
	for i, cmd := range e.commands {
		handler(fmt.Sprintf("📝 [%d/%d] 执行: %s", i+1, len(e.commands), cmd), false)
	}
	handler("", false)

	// 合并命令用 && 连接，确保前一条成功后才执行下一条
	combinedCmd := strings.Join(e.commands, " && ")

	if err := e.runCommand(ctx, client, combinedCmd, handler); err != nil {
		return fmt.Errorf("命令执行失败: %w", err)
	}

	return nil
}

// executeRemoteScript 执行远程脚本
func (e *SSHExecutor) executeRemoteScript(ctx context.Context, client *ssh.Client, handler OutputHandler) error {
	handler(fmt.Sprintf("📜 执行远程脚本: %s", e.script), false)
	return e.runCommand(ctx, client, e.script, handler)
}

// executeLocalScript 上传并执行本地脚本
func (e *SSHExecutor) executeLocalScript(ctx context.Context, client *ssh.Client, handler OutputHandler) error {
	handler(fmt.Sprintf("📤 上传本地脚本: %s", e.localScript), false)

	// 读取本地脚本
	scriptContent, err := os.ReadFile(e.localScript)
	if err != nil {
		return fmt.Errorf("读取本地脚本失败: %w", err)
	}

	// 创建远程临时脚本
	remotePath := "/tmp/xbuilder_script_" + fmt.Sprint(time.Now().UnixNano()) + ".sh"

	// 创建 session 上传脚本
	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("创建 SSH session 失败: %w", err)
	}

	// 使用 cat 写入脚本内容
	uploadCmd := fmt.Sprintf("cat > %s << 'XBUILDER_EOF'\n%s\nXBUILDER_EOF", remotePath, string(scriptContent))
	if err := session.Run(uploadCmd); err != nil {
		session.Close()
		return fmt.Errorf("上传脚本失败: %w", err)
	}
	session.Close()

	// 添加执行权限
	if err := e.runCommand(ctx, client, "chmod +x "+remotePath, handler); err != nil {
		return fmt.Errorf("设置脚本权限失败: %w", err)
	}

	handler("✅ 脚本上传成功", false)
	handler("📜 执行脚本...", false)
	handler("", false)

	// 执行脚本
	if err := e.runCommand(ctx, client, remotePath, handler); err != nil {
		return err
	}

	// 清理临时脚本
	e.runCommand(ctx, client, "rm -f "+remotePath, nil)

	return nil
}

// runCommand 运行单个命令
func (e *SSHExecutor) runCommand(ctx context.Context, client *ssh.Client, command string, handler OutputHandler) error {
	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("创建 SSH session 失败: %w", err)
	}
	defer session.Close()

	// 获取输出管道
	stdout, err := session.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := session.StderrPipe()
	if err != nil {
		return err
	}

	// 启动命令
	if err := session.Start(command); err != nil {
		return err
	}

	// 创建完成通道
	done := make(chan error, 1)

	// 异步读取输出
	go func() {
		e.readOutput(stdout, handler, false)
	}()
	go func() {
		e.readOutput(stderr, handler, true)
	}()

	// 等待命令完成
	go func() {
		done <- session.Wait()
	}()

	// 等待完成或取消
	select {
	case <-ctx.Done():
		session.Signal(ssh.SIGTERM)
		return ctx.Err()
	case err := <-done:
		return err
	}
}

// readOutput 读取输出
func (e *SSHExecutor) readOutput(r io.Reader, handler OutputHandler, isError bool) {
	if handler == nil {
		return
	}

	buf := make([]byte, 4096)
	var line strings.Builder

	for {
		n, err := r.Read(buf)
		if n > 0 {
			for i := 0; i < n; i++ {
				if buf[i] == '\n' {
					handler(line.String(), isError)
					line.Reset()
				} else {
					line.WriteByte(buf[i])
				}
			}
		}
		if err != nil {
			if line.Len() > 0 {
				handler(line.String(), isError)
			}
			break
		}
	}
}

// expandHomePath 展开 ~ 为 home 目录
func expandHomePath(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err == nil {
			return home + path[1:]
		}
	}
	return path
}

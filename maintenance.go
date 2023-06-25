package main

import (
	"bytes"
	"os/exec"
	"strings"
)

type DeployConfig struct {
	IsLocal           bool
	Path              string
	DockerComposeFile string
	DbName            string
	DbUser            string
	WebContainerName  string
	DbContainerName   string
	IsAutoDeploy      bool
	BackupScript      string
}

type Commands struct {
	Cd                  string
	Pull                string
	Down                string
	Up                  string
	Prune               string
	StartWeb            string
	StopWeb             string
	RestartWeb          string
	RestartAll          string
	DbReindex           string
	DbAnalyze           string
	DbVacuumFull        string
	DbVacuumFullAnalyze string
	DbVacuum            string
	DbVacuumAnalyze     string
}

var CommandsInstance Commands

func Deploy() error {
	command := CommandsInstance.Cd + " && " +
		CommandsInstance.Pull + " && " +
		CommandsInstance.Down + " && " +
		CommandsInstance.Up + " && " +
		CommandsInstance.Prune
	if DeployConfigInstance.IsLocal {
		err := exec.Command("sh", "-c", command).Run()
		if err == nil {
			// 如果部署成功则更新 status.json
			status := ReadStatus()
			status.IsDeployed = true
			WriteStatus(status)
		}
		return err
	} else {
		// 建立 SSH 连接
		sshClient, err := NewSSHClient(SSHConfigInstance)
		if err != nil {
			return err
		}
		// 新建 session
		session, err := sshClient.NewSession()
		if err != nil {
			return err
		}
		closeSSH := func() {
			_ = session.Close()
			_ = sshClient.Close()
		}
		defer closeSSH()
		// 执行部署命令
		_, err = RunCommand(session, command)
		if err == nil {
			// 如果部署成功则更新 status.json
			status := ReadStatus()
			status.IsDeployed = true
			WriteStatus(status)
		}
		return err
	}
}

func DbReindex() error {
	// 需要先停止 web 服务
	command := CommandsInstance.Cd + " && " + CommandsInstance.StopWeb + " && " + CommandsInstance.DbReindex
	if DeployConfigInstance.IsLocal {
		_, err := exec.Command("sh", "-c", command).Output()
		if err != nil {
			return err
		}
		return nil
	} else {
		// 建立 SSH 连接
		sshClient, err := NewSSHClient(SSHConfigInstance)
		if err != nil {
			return err
		}
		// 新建 session
		session, err := sshClient.NewSession()
		if err != nil {
			return err
		}
		closeSSH := func() {
			_ = session.Close()
			_ = sshClient.Close()
		}
		defer closeSSH()
		// 执行 db reindex 命令
		_, err = RunCommand(session, command)
		if err != nil {
			return err
		}
		return nil
	}
}

func DbVacuumFull() error {
	// 需要先停止 web 服务
	command := CommandsInstance.Cd + " && " + CommandsInstance.StopWeb + " && " + CommandsInstance.DbVacuumFull
	if DeployConfigInstance.IsLocal {
		out, err := exec.Command("sh", "-c", command).Output()
		if err != nil || !bytes.Contains(out, []byte("VACUUM")) {
			return err
		}
		return nil
	} else {
		// 建立 SSH 连接
		sshClient, err := NewSSHClient(SSHConfigInstance)
		if err != nil {
			return err
		}
		// 新建 session
		session, err := sshClient.NewSession()
		if err != nil {
			return err
		}
		closeSSH := func() {
			_ = session.Close()
			_ = sshClient.Close()
		}
		defer closeSSH()
		// 执行 db vacuum full 命令
		outBuffer, err := RunCommand(session, command)
		if err != nil || !strings.Contains(outBuffer.String(), "VACUUM") {
			return err
		}
		return nil
	}
}

func DbVacuumFullAnalyze() error {
	command := CommandsInstance.Cd + " && " + CommandsInstance.StopWeb + " && " + CommandsInstance.DbVacuumFullAnalyze
	if DeployConfigInstance.IsLocal {
		out, err := exec.Command("sh", "-c", command).Output()
		if err != nil || !bytes.Contains(out, []byte("VACUUM")) {
			return err
		}
		return nil
	} else {
		// 建立 SSH 连接
		sshClient, err := NewSSHClient(SSHConfigInstance)
		if err != nil {
			return err
		}
		// 新建 session
		session, err := sshClient.NewSession()
		if err != nil {
			return err
		}
		closeSSH := func() {
			_ = session.Close()
			_ = sshClient.Close()
		}
		defer closeSSH()
		// 执行 db vacuum full analyze 命令
		outBuffer, err := RunCommand(session, command)
		if err != nil || !strings.Contains(outBuffer.String(), "VACUUM") {
			return err
		}
		return nil
	}
}

func DbVacuum() error {
	command := CommandsInstance.Cd + " && " + CommandsInstance.StopWeb + " && " + CommandsInstance.DbVacuum
	if DeployConfigInstance.IsLocal {
		out, err := exec.Command("sh", "-c", command).Output()
		if err != nil || !bytes.Contains(out, []byte("VACUUM")) {
			return err
		}
		return nil
	} else {
		// 建立 SSH 连接
		sshClient, err := NewSSHClient(SSHConfigInstance)
		if err != nil {
			return err
		}
		// 新建 session
		session, err := sshClient.NewSession()
		if err != nil {
			return err
		}
		closeSSH := func() {
			_ = session.Close()
			_ = sshClient.Close()
		}
		defer closeSSH()
		// 执行 db vacuum 命令
		outBuffer, err := RunCommand(session, command)
		if err != nil || !strings.Contains(outBuffer.String(), "VACUUM") {
			return err
		}
		return nil
	}
}

func DbVacuumAnalyze() error {
	command := CommandsInstance.Cd + " && " + CommandsInstance.StopWeb + " && " + CommandsInstance.DbVacuumAnalyze
	if DeployConfigInstance.IsLocal {
		out, err := exec.Command("sh", "-c", command).Output()
		if err != nil || !bytes.Contains(out, []byte("VACUUM")) {
			return err
		}
		return nil
	} else {
		// 建立 SSH 连接
		sshClient, err := NewSSHClient(SSHConfigInstance)
		if err != nil {
			return err
		}
		// 新建 session
		session, err := sshClient.NewSession()
		if err != nil {
			return err
		}
		closeSSH := func() {
			_ = session.Close()
			_ = sshClient.Close()
		}
		defer closeSSH()
		// 执行 db vacuum 命令
		outBuffer, err := RunCommand(session, command)
		if err != nil || !strings.Contains(outBuffer.String(), "VACUUM") {
			return err
		}
		return nil
	}
}

func DbAnalyze() error {
	// ANALYZE 需要在关闭 Misskey 容器的情况下执行
	command := CommandsInstance.Cd + " && " + CommandsInstance.StopWeb + " && " + CommandsInstance.DbAnalyze
	if DeployConfigInstance.IsLocal {
		out, err := exec.Command("sh", "-c", command).Output()
		if err != nil || !bytes.Contains(out, []byte("ANALYZE")) {
			return err
		}
		return nil
	} else {
		// 建立 SSH 连接
		sshClient, err := NewSSHClient(SSHConfigInstance)
		if err != nil {
			return err
		}
		// 新建 session
		session, err := sshClient.NewSession()
		if err != nil {
			return err
		}
		closeSSH := func() {
			_ = session.Close()
			_ = sshClient.Close()
		}
		defer closeSSH()
		// 执行 db analyze 命令
		outBuffer, err := RunCommand(session, command)
		if err != nil || !strings.Contains(outBuffer.String(), "ANALYZE") {
			return err
		}
		return nil
	}
}

func ContainersDown() error {
	command := CommandsInstance.Cd + " && " + CommandsInstance.Down
	if DeployConfigInstance.IsLocal {
		err := exec.Command("sh", "-c", command).Run()
		return err
	} else {
		// 建立 SSH 连接
		sshClient, err := NewSSHClient(SSHConfigInstance)
		if err != nil {
			return err
		}
		// 新建 session
		session, err := sshClient.NewSession()
		if err != nil {
			return err
		}
		closeSSH := func() {
			_ = session.Close()
			_ = sshClient.Close()
		}
		defer closeSSH()
		// 执行 docker compose down 命令
		_, err = RunCommand(session, command)
		if err != nil {
			return err
		}
		return nil
	}
}

func ContainersUp() error {
	command := CommandsInstance.Cd + " && " + CommandsInstance.Up
	if DeployConfigInstance.IsLocal {
		err := exec.Command("sh", "-c", command).Run()
		return err
	} else {
		// 建立 SSH 连接
		sshClient, err := NewSSHClient(SSHConfigInstance)
		if err != nil {
			return err
		}
		// 新建 session
		session, err := sshClient.NewSession()
		if err != nil {
			return err
		}
		closeSSH := func() {
			_ = session.Close()
			_ = sshClient.Close()
		}
		defer closeSSH()
		// 执行 docker compose up -d 命令
		_, err = RunCommand(session, command)
		if err != nil {
			return err
		}
		return nil
	}
}

func ContainersRestart() error {
	command := CommandsInstance.Cd + " && " + CommandsInstance.RestartAll
	if DeployConfigInstance.IsLocal {
		err := exec.Command("sh", "-c", command).Run()
		return err
	} else {
		// 建立 SSH 连接
		sshClient, err := NewSSHClient(SSHConfigInstance)
		if err != nil {
			return err
		}
		// 新建 session
		session, err := sshClient.NewSession()
		if err != nil {
			return err
		}
		closeSSH := func() {
			_ = session.Close()
			_ = sshClient.Close()
		}
		defer closeSSH()
		// 执行 docker compose restart 命令
		_, err = RunCommand(session, command)
		if err != nil {
			return err
		}
		return nil
	}
}

func StopMisskeyContainer() error {
	command := CommandsInstance.Cd + " && " + CommandsInstance.StopWeb
	if DeployConfigInstance.IsLocal {
		err := exec.Command("sh", "-c", command).Run()
		return err
	} else {
		// 建立 SSH 连接
		sshClient, err := NewSSHClient(SSHConfigInstance)
		if err != nil {
			return err
		}
		// 新建 session
		session, err := sshClient.NewSession()
		if err != nil {
			return err
		}
		closeSSH := func() {
			_ = session.Close()
			_ = sshClient.Close()
		}
		defer closeSSH()
		// 执行 docker compose stop 命令
		_, err = RunCommand(session, command)
		if err != nil {
			return err
		}
		return nil
	}
}

func StartMisskeyContainer() error {
	command := CommandsInstance.Cd + " && " + CommandsInstance.StartWeb
	if DeployConfigInstance.IsLocal {
		err := exec.Command("sh", "-c", command).Run()
		return err
	} else {
		// 建立 SSH 连接
		sshClient, err := NewSSHClient(SSHConfigInstance)
		if err != nil {
			return err
		}
		// 新建 session
		session, err := sshClient.NewSession()
		if err != nil {
			return err
		}
		closeSSH := func() {
			_ = session.Close()
			_ = sshClient.Close()
		}
		defer closeSSH()
		// 执行 docker compose start 命令
		_, err = RunCommand(session, command)
		if err != nil {
			return err
		}
		return nil
	}
}

func DbBackup(isUpgrade bool) error {
	status := ReadStatus()
	command := "bash " + DeployConfigInstance.BackupScript + " " + status.LatestVersion
	if isUpgrade {
		command += " 1"
	} else {
		command += " 0"
	}
	if DeployConfigInstance.IsLocal {
		err := exec.Command("sh", "-c", command).Run()
		return err
	} else {
		// 建立 SSH 连接
		sshClient, err := NewSSHClient(SSHConfigInstance)
		if err != nil {
			return err
		}
		// 新建 session
		session, err := sshClient.NewSession()
		if err != nil {
			return err
		}
		closeSSH := func() {
			_ = session.Close()
			_ = sshClient.Close()
		}
		defer closeSSH()
		// 执行 docker compose backup 命令
		_, err = RunCommand(session, command)
		if err != nil {
			return err
		}
		return nil
	}
}

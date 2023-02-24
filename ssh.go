package main

import (
    "bytes"
    "golang.org/x/crypto/ssh"
    "strconv"
)

type SSHClientConfig struct {
    User                 string
    Host                 string
    Port                 int
    Password             string
    PrivateKey           string
    PrivateKeyPassPhrase string
}

func NewSSHClient(clientConfig SSHClientConfig) (*ssh.Client, error) {
    // 这里是连接到服务器的代码
    config := &ssh.ClientConfig{
        User: clientConfig.User,
        Auth: []ssh.AuthMethod{
            ssh.Password(clientConfig.Password),
        },
        HostKeyCallback: ssh.InsecureIgnoreHostKey(),
    }
    // PrivateKey 优先级高于 Password
    // 不带密码的私钥
    if clientConfig.PrivateKey != "" && clientConfig.PrivateKeyPassPhrase == "" {
        privateKey, err := ssh.ParsePrivateKey([]byte(clientConfig.PrivateKey))
        if err != nil {
            return nil, err
        }
        config.Auth = []ssh.AuthMethod{
            ssh.PublicKeys(privateKey),
        }
    } else if clientConfig.PrivateKey != "" && clientConfig.PrivateKeyPassPhrase != "" {
        // 带密码的私钥
        privateKey, err := ssh.ParsePrivateKeyWithPassphrase([]byte(clientConfig.PrivateKey), []byte(clientConfig.PrivateKeyPassPhrase))
        if err != nil {
            return nil, err
        }
        config.Auth = []ssh.AuthMethod{
            ssh.PublicKeys(privateKey),
        }
    }
    client, err := ssh.Dial("tcp", SSHConfigInstance.Host+":"+strconv.Itoa(SSHConfigInstance.Port), config)
    if err != nil {
        return nil, err
    }
    return client, nil
}

func RunCommand(session *ssh.Session, command string) (*bytes.Buffer, error) {
    // 执行命令并返回结果
    var stdout bytes.Buffer
    session.Stdout = &stdout
    // session.RequestPty("xterm", 80, 40, ssh.TerminalModes{})
    err := session.Run(command)
    if err != nil {
        return &bytes.Buffer{}, err
    }
    return &stdout, nil
}

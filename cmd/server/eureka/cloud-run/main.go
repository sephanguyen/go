package main

import (
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

type WithusRelayServerConfig struct {
	IPAddress        string `yaml:"ip_address"`
	Port             string `yaml:"port"`
	UserName         string `yaml:"user_name"`
	PrivateKeyBase64 string `yaml:"private_key"`
	Password         string `yaml:"password"`
	LMSToERPFilePath string `yaml:"lms_to_erp_file_path"`
	ReportFieldName  string `yaml:"report_field_name"`
}

func (c *WithusRelayServerConfig) loadForWithusServer() *WithusRelayServerConfig {
	c.IPAddress = os.Getenv("SERVER_IP")
	c.Port = os.Getenv("SERVER_PORT")
	c.UserName = os.Getenv("USERNAME")
	c.PrivateKeyBase64 = os.Getenv("PRIVATE_KEY")
	c.LMSToERPFilePath = os.Getenv("LMS_TO_ERP_FILE_PATH")
	c.ReportFieldName = "REPORTS"

	return c
}

func (c *WithusRelayServerConfig) loadForITeeServer() *WithusRelayServerConfig {
	c.IPAddress = os.Getenv("ITEE_SERVER_IP")
	c.Port = os.Getenv("ITEE_SERVER_PORT")
	c.UserName = os.Getenv("ITEE_USERNAME")
	c.Password = os.Getenv("ITEE_PASSWORD")
	c.LMSToERPFilePath = os.Getenv("ITEE_LMS_TO_ERP_FILE_PATH")
	c.ReportFieldName = "REPORTS"

	return c
}

func main() {
	router := gin.Default()

	router.POST("/upload-file/managara-base", UploadFileManagaraBaseHandler)
	router.POST("/upload-file/managara-hs", UploadFileManagaraHSHandler)
	err := router.Run(":1234")
	if err != nil {
		log.Fatal("unable to run app")
		return
	}
}

func generateSigner(privateKeyBase64 string) (ssh.Signer, error) {
	privateKeyText, err := base64.StdEncoding.DecodeString(privateKeyBase64)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64: %w", err)
	}
	signerKey, err := ssh.ParsePrivateKey(privateKeyText)
	if err != nil {
		return nil, fmt.Errorf("failed to parse signer key: %w", err)
	}
	return signerKey, nil
}

func UploadFileManagaraBaseHandler(c *gin.Context) {
	// Send L6 file to Withus server
	withusServerCfg := new(WithusRelayServerConfig).loadForWithusServer()
	withusServerSigner, err := generateSigner(withusServerCfg.PrivateKeyBase64)
	if err != nil {
		fmt.Printf("generateSigner: %v", err)
		c.String(http.StatusInternalServerError, fmt.Sprintf("generateSigner: %s", err.Error()))
		return
	}
	withusFileName, err := uploadFileToWithusServer(c, "Withus", withusServerCfg, ssh.PublicKeys(withusServerSigner))
	if err != nil {
		fmt.Printf("uploadFileToWithusServer: %v", err)
		c.String(http.StatusInternalServerError, fmt.Sprintf("uploadFileToWithusServer: %s", err.Error()))
		return
	}

	c.String(http.StatusOK, fmt.Sprintf("'%s' uploaded!", withusFileName))
}

func UploadFileManagaraHSHandler(c *gin.Context) {
	// Send M1 file to iTee server
	iTeeServerCfg := new(WithusRelayServerConfig).loadForITeeServer()
	iTeeFileName, err := uploadFileToWithusServer(c, "Withus", iTeeServerCfg, ssh.Password(iTeeServerCfg.Password))
	if err != nil {
		fmt.Printf("uploadFileToWithusServer: %v", err)
		c.String(http.StatusInternalServerError, fmt.Sprintf("uploadFileToWithusServer: %s", err.Error()))
		return
	}

	c.String(http.StatusOK, fmt.Sprintf("'%s' uploaded!", iTeeFileName))
}

func uploadFileToWithusServer(c *gin.Context, serverName string, cfg *WithusRelayServerConfig, sshAuthMethod ssh.AuthMethod) (string, error) {
	log.Printf("-----Start uploading data file to %s Server-----\n", serverName)

	file, err := c.FormFile(cfg.ReportFieldName)
	if err != nil {
		return "", fmt.Errorf("c.FormFile: %w", err)
	}

	sshClient, err := ssh.Dial("tcp", fmt.Sprintf("%s:%s", cfg.IPAddress, cfg.Port), &ssh.ClientConfig{
		User:            cfg.UserName,
		Auth:            []ssh.AuthMethod{sshAuthMethod},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         10 * time.Second,
	})
	if err != nil {
		return "", fmt.Errorf("failed to dial to server: %w", err)
	}
	defer sshClient.Close()
	sftpClient, err := sftp.NewClient(sshClient)
	if err != nil {
		return "", fmt.Errorf("failed to create new sftp client: %w", err)
	}
	defer sftpClient.Close()

	dstFile, err := sftpClient.Create(cfg.LMSToERPFilePath + "/" + file.Filename)
	if err != nil {
		return "", fmt.Errorf("failed to create dst file: %w", err)
	}
	defer dstFile.Close()

	srcFile, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("failed to open file")
	}

	bytes, err := io.Copy(dstFile, srcFile)
	if err != nil {
		log.Printf("failed to copy src file to dst file: byte copied: %d", bytes)
		return "", fmt.Errorf("failed to copy src file to dst file: %w", err)
	}

	log.Printf("-----Finish uploading data file to %s Server-----\n", serverName)
	return file.Filename, nil
}
